package data

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewSimpleCache(t *testing.T) {
	cache := NewSimpleCache(10)

	var wg sync.WaitGroup
	for j := 0; j < 10; j++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 10_000; i++ {
				cache.Add(fmt.Sprintf("%d", i), []byte{})
			}
			wg.Done()
		}()
	}

	wg.Wait()

	found := 0
	for i := 0; i < 10_000; i++ {
		_, ok := cache.Get(fmt.Sprintf("%d", i))
		if ok {
			found++
		}
	}

	if found != 10 {
		t.Error("expected only ten items", found)
	}
}

func TestSimpleCache_Add(t *testing.T) {
	cache := NewSimpleCache(5)
	cache.Add("oldest", []byte{})

	_, ok := cache.Get("oldest")
	if !ok {
		t.Error("should exist")
	}

	// sleep just long enough to ensure we expire this one
	time.Sleep(2 * time.Millisecond)

	for i := 0; i < 10000; i++ {
		cache.Add(fmt.Sprintf("%v", i), []byte{})
	}

	_, ok = cache.Get("oldest")
	if ok {
		t.Error("should not exist")
	}
}

func TestSimpleCache_AddUpdate(t *testing.T) {
	cache := NewSimpleCache(5)
	cache.Add("oldest", []byte{})

	_, ok := cache.Get("oldest")
	if !ok {
		t.Error("should exist")
	}

	// sleep just long enough to ensure we expire this one
	time.Sleep(2 * time.Millisecond)

	for i := 0; i < 1000; i++ {
		cache.Add(fmt.Sprintf("%v", i), []byte{})
		_, ok = cache.Get("oldest")
		if !ok {
			t.Error("should exist")
		}
	}

	_, ok = cache.Get("oldest")
	if !ok {
		t.Error("should still exist")
	}
}
