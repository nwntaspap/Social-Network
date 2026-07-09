package cache

import (
	"sync"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", 0)
	got, err := c.Get("key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != "value" {
		t.Fatalf("got %v, want value", got)
	}
}

func TestGetMissing(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	_, err := c.Get("noexist")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestGetExpired(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	_, err := c.Get("key")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound after expiry, got %v", err)
	}
}

func TestSetNXNew(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.SetNX("key", "value", 0)
	got, err := c.Get("key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != "value" {
		t.Fatalf("got %v, want value", got)
	}
}

func TestSetNXExisting(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "original", 0)
	c.SetNX("key", "overwrite", 0)

	got, _ := c.Get("key")
	if got != "original" {
		t.Fatalf("SetNX overwrote key: got %v, want original", got)
	}
}

func TestSetNXExpired(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "original", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	c.SetNX("key", "newvalue", 0)
	got, err := c.Get("key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != "newvalue" {
		t.Fatalf("expected newvalue after SetNX on expired key, got %v", got)
	}
}

func TestDelete(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", 0)
	c.Delete("key")

	_, err := c.Get("key")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound after Delete, got %v", err)
	}
}

func TestDeleteMissingIsIdempotent(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Delete("noexist")
}

func TestExpire(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", 0)
	err := c.Expire("key", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	_, err = c.Get("key")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound after Expire, got %v", err)
	}
}

func TestExpireMissing(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	err := c.Expire("noexist", time.Second)
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestExpireRemoveTTL(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", 10*time.Millisecond)
	c.Expire("key", 0)

	time.Sleep(20 * time.Millisecond)
	got, err := c.Get("key")
	if err != nil {
		t.Fatalf("key should persist after removing TTL: %v", err)
	}
	if got != "value" {
		t.Fatalf("got %v, want value", got)
	}
}

func TestTTLMissing(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	got := c.TTL("noexist")
	if got != -2 {
		t.Fatalf("expected -2 for missing key, got %v", got)
	}
}

func TestTTLNoExpiry(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", 0)
	got := c.TTL("key")
	if got != -1 {
		t.Fatalf("expected -1 for no expiry, got %v", got)
	}
}

func TestTTLWithExpiry(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", time.Hour)
	got := c.TTL("key")
	if got <= 0 || got > time.Hour {
		t.Fatalf("expected ~1 hour, got %v", got)
	}
}

func TestExists(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	if c.Exists("key") {
		t.Fatal("expected false for missing key")
	}

	c.Set("key", "value", 0)
	if !c.Exists("key") {
		t.Fatal("expected true for existing key")
	}
}

func TestExistsExpired(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	if c.Exists("key") {
		t.Fatal("expected false for expired key")
	}
}

func TestLazyDeleteOnGet(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	_, _ = c.Get("key")

	_, err := c.Get("key")
	if err != ErrKeyNotFound {
		t.Fatal("expected ErrKeyNotFound after lazy delete")
	}
}

func TestBackgroundCleanup(t *testing.T) {
	c := NewInMemoryCache(WithCleanupInterval(50 * time.Millisecond))
	defer c.Stop()

	c.Set("key1", "value", 10*time.Millisecond)
	c.Set("key2", "value", 10*time.Millisecond)
	c.Set("perm", "value", 0)

	time.Sleep(100 * time.Millisecond)

	if c.Exists("key1") {
		t.Fatal("expected key1 to be cleaned up")
	}
	if c.Exists("key2") {
		t.Fatal("expected key2 to be cleaned up")
	}
	if !c.Exists("perm") {
		t.Fatal("expected perm to persist")
	}
}

func TestConcurrentSetGet(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := string(rune('a' + n%26))
			if n%2 == 0 {
				c.Set(key, n, 0)
			}
			_, err := c.Get(key)
			if err != nil && err != ErrKeyNotFound {
				t.Logf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentSetNX(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "original", 0)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetNX("key", "overwrite", 0)
		}()
	}
	wg.Wait()

	got, _ := c.Get("key")
	if got != "original" {
		t.Fatal("SetNX should not overwrite existing key")
	}
}

func TestStopIdempotent(t *testing.T) {
	c := NewInMemoryCache()
	c.Stop()
	c.Stop()
}

func TestNilValue(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", nil, 0)
	got, err := c.Get("key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestNegativeTTL(t *testing.T) {
	c := NewInMemoryCache()
	defer c.Stop()

	c.Set("key", "value", -time.Second)
	got := c.TTL("key")
	if got != -1 {
		t.Fatalf("expected -1 for negative TTL (no expiry), got %v", got)
	}
}
