package service

import (
	"sync"

	"github.com/martin-pouls/regicide/backend/internal/core/domain"
)

type lockEntry struct {
	mu   sync.Mutex
	refs int
}

// keyedMutex is a per-game mutex; entries are removed when no one holds or waits for them.
type keyedMutex struct {
	mu      sync.Mutex
	entries map[domain.GameID]*lockEntry
}

func newKeyedMutex() *keyedMutex {
	return &keyedMutex{entries: make(map[domain.GameID]*lockEntry)}
}

func (k *keyedMutex) lock(id domain.GameID) {
	k.mu.Lock()
	e, ok := k.entries[id]
	if !ok {
		e = &lockEntry{}
		k.entries[id] = e
	}
	e.refs++
	k.mu.Unlock()
	e.mu.Lock()
}

func (k *keyedMutex) unlock(id domain.GameID) {
	k.mu.Lock()
	e := k.entries[id]
	e.mu.Unlock()
	e.refs--
	if e.refs == 0 {
		delete(k.entries, id)
	}
	k.mu.Unlock()
}

func (k *keyedMutex) size() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.entries)
}
