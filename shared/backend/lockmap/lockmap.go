// Package lockmap is a mutex per key (one per game), so mutations of one game are serialised
// while different games run in parallel.
package lockmap

import "sync"

type entry struct {
	mu   sync.Mutex
	refs int
}

// Map is a per-key mutex; entries are removed when no one holds or waits for them.
// The zero value is not usable: create one with New.
type Map[K comparable] struct {
	mu      sync.Mutex
	entries map[K]*entry
}

// New returns an empty Map.
func New[K comparable]() *Map[K] {
	return &Map[K]{entries: make(map[K]*entry)}
}

// Lock locks key, waiting while another caller holds it.
func (m *Map[K]) Lock(key K) {
	m.mu.Lock()
	e, ok := m.entries[key]
	if !ok {
		e = &entry{}
		m.entries[key] = e
	}
	e.refs++
	m.mu.Unlock()
	e.mu.Lock()
}

// Unlock unlocks key. It must be held.
func (m *Map[K]) Unlock(key K) {
	m.mu.Lock()
	e := m.entries[key]
	e.mu.Unlock()
	e.refs--
	if e.refs == 0 {
		delete(m.entries, key)
	}
	m.mu.Unlock()
}

// Len returns the number of keys held or waited for.
func (m *Map[K]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}
