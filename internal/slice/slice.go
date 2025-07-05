package slice

import (
	"slices"

	mutex "sync"
)

type Sync[T comparable] struct {
	mu    mutex.RWMutex
	items []T
}

func NewSync[T comparable]() *Sync[T] {
	return &Sync[T]{
		items: []T{},
	}
}

func (s *Sync[T]) Append(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = append(s.items, item)
}

func (s *Sync[T]) AppendMany(items []T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = append(s.items, items...)
}

func (s *Sync[T]) Get() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.items
}

func (s *Sync[T]) Contains(item T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return slices.Contains(s.items, item)
}

func (s *Sync[T]) Delete(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = slices.DeleteFunc(s.items, func(sliceItem T) bool {
		return sliceItem == item
	})
}

func (s *Sync[T]) Length() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.items)
}

func (s *Sync[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = []T{}
}
