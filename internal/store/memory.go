package store

import (
	"sync"

	"github.com/doganarif/govisual/internal/model"
)

type InMemoryStore struct {
	logs     []*model.RequestLog
	capacity int
	size     int
	next     int
	mu       sync.RWMutex
}

func NewInMemoryStore(capacity int) *InMemoryStore {
	if capacity <= 0 {
		capacity = 100
	}

	return &InMemoryStore{
		logs:     make([]*model.RequestLog, capacity),
		capacity: capacity,
		size:     0,
		next:     0,
	}
}

func (s *InMemoryStore) Add(log *model.RequestLog) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs[s.next] = log

	s.next = (s.next + 1) % s.capacity

	if s.size < s.capacity {
		s.size++
	}
}

func (s *InMemoryStore) Get(id string) (*model.RequestLog, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, log := range s.logs[:s.size] {
		if log != nil && log.ID == id {
			return log, true
		}
	}

	return nil, false
}

func (s *InMemoryStore) GetAll() []*model.RequestLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*model.RequestLog, 0, s.size)

	if s.size < s.capacity {
		for i := 0; i < s.size; i++ {
			result = append(result, s.logs[i])
		}
		return result
	}

	for i := s.next; i < s.capacity; i++ {
		result = append(result, s.logs[i])
	}
	for i := 0; i < s.next; i++ {
		result = append(result, s.logs[i])
	}

	return result
}

func (s *InMemoryStore) GetLatest(n int) []*model.RequestLog {
	all := s.GetAll()

	if len(all) <= n {
		return all
	}

	return all[len(all)-n:]
}

// Clear clears all stored request logs
func (s *InMemoryStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.size == 0 {
		return nil
	}

	s.logs = make([]*model.RequestLog, 0)
	s.size = 0
	s.next = 0
	return nil
}

// Close implements the Store interface but does nothing for in-memory store
func (s *InMemoryStore) Close() error {
	// Nothing to do for in-memory store
	return nil
}
