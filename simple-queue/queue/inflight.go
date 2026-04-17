package queue

import (
	"sync"
	"time"
)

type InFlightMessage struct {
	Msg          Message
	ShardIdx     int
	DequeueCount int
	ExpiresAt    time.Time
}

type InFlightStore struct {
	mu           sync.Mutex
	items        map[string]*InFlightMessage
	lockedGroups map[string]int // groupID -> count of in-flight msgs
}

func NewInFlightStore() *InFlightStore {
	return &InFlightStore{
		items:        make(map[string]*InFlightMessage),
		lockedGroups: make(map[string]int),
	}
}

func (s *InFlightStore) Add(msg Message, shardIdx int, dequeueCount int, timeout time.Duration) {
	s.mu.Lock()
	s.items[msg.ID] = &InFlightMessage{
		Msg:          msg,
		ShardIdx:     shardIdx,
		DequeueCount: dequeueCount,
		ExpiresAt:    time.Now().Add(timeout),
	}
	s.lockedGroups[msg.GroupID]++
	s.mu.Unlock()
}

func (s *InFlightStore) Ack(id string) bool {
	s.mu.Lock()
	item, ok := s.items[id]
	if ok {
		s.lockedGroups[item.Msg.GroupID]--
		if s.lockedGroups[item.Msg.GroupID] <= 0 {
			delete(s.lockedGroups, item.Msg.GroupID)
		}
		delete(s.items, id)
	}
	s.mu.Unlock()
	return ok
}

func (s *InFlightStore) IsGroupLocked(groupID string) bool {
	s.mu.Lock()
	locked := s.lockedGroups[groupID] > 0
	s.mu.Unlock()
	return locked
}

func (s *InFlightStore) Reap() []*InFlightMessage {
	now := time.Now()
	var expired []*InFlightMessage

	s.mu.Lock()
	for id, item := range s.items {
		if now.After(item.ExpiresAt) {
			expired = append(expired, item)
			s.lockedGroups[item.Msg.GroupID]--
			if s.lockedGroups[item.Msg.GroupID] <= 0 {
				delete(s.lockedGroups, item.Msg.GroupID)
			}
			delete(s.items, id)
		}
	}
	s.mu.Unlock()

	return expired
}

func (s *InFlightStore) Size() int {
	s.mu.Lock()
	n := len(s.items)
	s.mu.Unlock()
	return n
}
