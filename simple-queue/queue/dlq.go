package queue

import "sync"

type DeadLetterQueue struct {
	mu    sync.Mutex
	items []Message
}

func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		items: make([]Message, 0),
	}
}

func (d *DeadLetterQueue) Add(msg Message) {
	d.mu.Lock()
	d.items = append(d.items, msg)
	d.mu.Unlock()
}

func (d *DeadLetterQueue) List(limit int) []Message {
	d.mu.Lock()
	defer d.mu.Unlock()

	if limit <= 0 || limit > len(d.items) {
		limit = len(d.items)
	}
	result := make([]Message, limit)
	copy(result, d.items[:limit])
	return result
}

func (d *DeadLetterQueue) Size() int {
	d.mu.Lock()
	n := len(d.items)
	d.mu.Unlock()
	return n
}
