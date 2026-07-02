package store

import (
	"sync"

)

type memoryStore struct {
	logs     []*RequestLog
	index    map[string]int // ID -> position in logs (O(1) Get)
	capacity int
	size     int
	next     int
	mu       sync.RWMutex
}

func NewMemory(capacity int) Store {
	if capacity <= 0 {
		capacity = 100
	}

	return &memoryStore{
		logs:     make([]*RequestLog, capacity),
		index:    make(map[string]int, capacity),
		capacity: capacity,
	}
}

func (s *memoryStore) Add(log *RequestLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If the ring buffer is full, the slot we're about to overwrite holds
	// the oldest entry. Evict it from the ID index first.
	if old := s.logs[s.next]; old != nil {
		delete(s.index, old.ID)
	}

	s.logs[s.next] = log
	s.index[log.ID] = s.next
	s.next = (s.next + 1) % s.capacity

	if s.size < s.capacity {
		s.size++
	}
	return nil
}

func (s *memoryStore) Get(id string) (*RequestLog, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pos, ok := s.index[id]
	if !ok {
		return nil, false
	}
	return s.logs[pos], true
}

// GetAll returns all stored logs in newest-first order. This matches the
// ordering contract of the SQL/Mongo/Redis backends, so callers can treat the
// returned slice uniformly regardless of which store backs them.
func (s *memoryStore) GetAll() []*RequestLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*RequestLog, 0, s.size)
	// Walk backwards from the most-recently-written slot (s.next-1) for s.size
	// steps, wrapping. This yields newest-first without an extra sort.
	for i := 0; i < s.size; i++ {
		idx := (s.next - 1 - i + s.capacity) % s.capacity
		result = append(result, s.logs[idx])
	}
	return result
}

// GetLatest returns the n most recent logs, newest-first.
func (s *memoryStore) GetLatest(n int) []*RequestLog {
	all := s.GetAll()
	if len(all) <= n {
		return all
	}
	return all[:n]
}

// Clear clears all stored request logs
func (s *memoryStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.size == 0 {
		return nil
	}

	s.logs = make([]*RequestLog, s.capacity)
	s.index = make(map[string]int, s.capacity)
	s.size = 0
	s.next = 0
	return nil
}

// Close implements the Store interface but does nothing for in-memory store
func (s *memoryStore) Close() error {
	return nil
}
