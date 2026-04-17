package queue

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"example.com/simple-queue/metrics"
)

const ShardsQuantity = 10

type Message struct {
	ID           string `json:"id"`
	GroupID      string `json:"groupId"`
	Body         string `json:"body"`
	DequeueCount int    `json:"dequeueCount"`
}

type Shard struct {
	mu    sync.Mutex
	items []Message
}

const MaxReceiveCount = 3

type Queue struct {
	shards            []Shard
	size              uint64
	nextID            uint64
	inFlight          *InFlightStore
	dlq               *DeadLetterQueue
	visibilityTimeout time.Duration
	Metrics           *metrics.Metrics
}

var rrCounter uint64

func NewQueue(numShards int, visibilityTimeout time.Duration, ctx context.Context) *Queue {
	q := &Queue{
		shards:            make([]Shard, numShards),
		inFlight:          NewInFlightStore(),
		dlq:               NewDeadLetterQueue(),
		visibilityTimeout: visibilityTimeout,
		Metrics:           metrics.New(),
	}
	for i := range q.shards {
		q.shards[i] = Shard{
			items: make([]Message, 0),
		}
	}

	go q.reapLoop(ctx)
	go q.gaugeLoop(ctx)
	return q
}

func (q *Queue) reapLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.requeueExpired()
		}
	}
}

func (q *Queue) gaugeLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.updateGauges()
		}
	}
}

func (q *Queue) updateGauges() {
	q.Metrics.QueueSize.Set(float64(q.Size()))
	q.Metrics.InFlightSize.Set(float64(q.InFlightSize()))
	q.Metrics.DLQSize.Set(float64(q.DLQSize()))
	for i, size := range q.ShardSizes() {
		q.Metrics.ShardSize.WithLabelValues(strconv.Itoa(i)).Set(float64(size))
	}
}

func (q *Queue) requeueExpired() {
	expired := q.inFlight.Reap()
	for _, item := range expired {
		msg := item.Msg
		msg.DequeueCount = item.DequeueCount

		if msg.DequeueCount >= MaxReceiveCount {
			q.dlq.Add(msg)
			q.Metrics.RecordDLQ()
			continue
		}

		shard := &q.shards[item.ShardIdx]
		shard.mu.Lock()
		shard.items = append(shard.items, msg)
		shard.mu.Unlock()

		atomic.AddUint64(&q.size, 1)
		q.Metrics.RecordRequeue()
	}
}

func (q *Queue) getShard(groupID string) (*Shard, int) {
	h := fnv.New32a()
	h.Write([]byte(groupID))
	idx := int(h.Sum32() % uint32(len(q.shards)))
	return &q.shards[idx], idx
}

func (q *Queue) Enqueue(msg Message) string {
	id := fmt.Sprintf("%d", atomic.AddUint64(&q.nextID, 1))
	msg.ID = id

	shard, _ := q.getShard(msg.GroupID)
	shard.mu.Lock()
	shard.items = append(shard.items, msg)
	shard.mu.Unlock()

	atomic.AddUint64(&q.size, 1)
	q.Metrics.RecordEnqueue()
	return id
}

func (q *Queue) Dequeue() (Message, bool) {
	n := len(q.shards)
	for i := 0; i < n; i++ {
		idx := int(atomic.AddUint64(&rrCounter, 1) % uint64(n))
		shard := &q.shards[idx]

		shard.mu.Lock()
		for j := 0; j < len(shard.items); j++ {
			if q.inFlight.IsGroupLocked(shard.items[j].GroupID) {
				continue
			}

			msg := shard.items[j]
			shard.items = append(shard.items[:j], shard.items[j+1:]...)
			shard.mu.Unlock()

			atomic.AddUint64(&q.size, ^uint64(0))
			msg.DequeueCount++
			q.inFlight.Add(msg, idx, msg.DequeueCount, q.visibilityTimeout)
			q.Metrics.RecordDequeue()
			return msg, true
		}
		shard.mu.Unlock()
	}
	return Message{}, false
}

func (q *Queue) Ack(id string) bool {
	ok := q.inFlight.Ack(id)
	if ok {
		q.Metrics.RecordAck()
	}
	return ok
}

func (q *Queue) Size() uint64 {
	return atomic.LoadUint64(&q.size)
}

func (q *Queue) InFlightSize() int {
	return q.inFlight.Size()
}

func (q *Queue) DLQSize() int {
	return q.dlq.Size()
}

func (q *Queue) DLQList(limit int) []Message {
	return q.dlq.List(limit)
}

func (q *Queue) ShardSizes() []int {
	sizes := make([]int, len(q.shards))
	for i := range q.shards {
		q.shards[i].mu.Lock()
		sizes[i] = len(q.shards[i].items)
		q.shards[i].mu.Unlock()
	}
	return sizes
}
