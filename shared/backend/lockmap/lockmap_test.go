package lockmap

import (
	"sync"
	"testing"
	"time"
)

func TestSerialisesOneKey(t *testing.T) {
	m := New[string]()
	var wg sync.WaitGroup
	counter := 0
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Lock("g")
			c := counter
			time.Sleep(time.Microsecond)
			counter = c + 1
			m.Unlock("g")
		}()
	}
	wg.Wait()
	if counter != 50 {
		t.Fatalf("counter = %d, want 50 (lost updates)", counter)
	}
	if m.Len() != 0 {
		t.Fatalf("Len = %d after all unlocks, want 0", m.Len())
	}
}

func TestKeysAreIndependent(t *testing.T) {
	m := New[string]()
	m.Lock("a")
	done := make(chan struct{})
	go func() {
		m.Lock("b")
		m.Unlock("b")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("locking b waited for a")
	}
	if m.Len() != 1 {
		t.Fatalf("Len = %d, want 1", m.Len())
	}
	m.Unlock("a")
}
