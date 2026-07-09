package cache

import (
	"sync"
	"time"
)

type entry struct {
	value      any
	expiration time.Time
}

type InMemoryCache struct {
	mu              sync.RWMutex
	data            map[string]*entry
	stopCh          chan struct{}
	wg              sync.WaitGroup
	stopOnce        sync.Once
	cleanupInterval time.Duration
}

type CacheOption func(*InMemoryCache)

func WithCleanupInterval(d time.Duration) CacheOption {
	return func(c *InMemoryCache) {
		if d > 0 {
			c.cleanupInterval = d
		}
	}
}

// options are basically like closures
// for example
//
//	func WithCleanupInterval(d time.Duration) CacheOption {
//		return func(c *InMemoryCache) {
//			if d > 0 {
//				c.cleanupInterval = d
//			}
//		}
//	}
//
// it is a closure so that
//
//	NewInMemoryCache(WithCleanupInternal(50 * time.Millisecond))
//
// eventually passes the interval in the struct instance
func NewInMemoryCache(opts ...CacheOption) *InMemoryCache {
	c := &InMemoryCache{
		data:            make(map[string]*entry),
		stopCh:          make(chan struct{}),
		cleanupInterval: 1 * time.Minute,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.wg.Add(1)
	go c.cleanup()
	return c
}

func (c *InMemoryCache) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
		c.wg.Wait()
	})
}

func (c *InMemoryCache) cleanup() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, e := range c.data {
				if !e.expiration.IsZero() && now.After(e.expiration) {
					delete(c.data, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *InMemoryCache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	c.data[key] = &entry{value: value, expiration: exp}
}

func (c *InMemoryCache) SetNX(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.data[key]
	if ok && !isExpired(e) {
		return
	}

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	c.data[key] = &entry{value: value, expiration: exp}
}

func (c *InMemoryCache) Get(key string) (any, error) {
	c.mu.RLock()
	e, ok := c.data[key]
	if !ok {
		c.mu.RUnlock()
		return nil, ErrKeyNotFound
	}

	if isExpired(e) {
		c.mu.RUnlock()
		//we check a second time because RLock does not block writes
		//to happen somewhere else
		//so between seeing that it is expired and trying to delete it
		//someone could have written this value.
		//thats why we Lock now to block writes, and delete
		c.mu.Lock()
		if e2, ok := c.data[key]; ok && isExpired(e2) {
			delete(c.data, key)
		}
		c.mu.Unlock()
		return nil, ErrKeyNotFound
	}

	c.mu.RUnlock()
	return e.value, nil
}

func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *InMemoryCache) Expire(key string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.data[key]
	if !ok {
		return ErrKeyNotFound
	}

	if isExpired(e) {
		delete(c.data, key)
		return ErrKeyNotFound
	}

	if ttl > 0 {
		e.expiration = time.Now().Add(ttl)
	} else {
		e.expiration = time.Time{}
	}

	return nil
}

func (c *InMemoryCache) TTL(key string) time.Duration {
	c.mu.RLock()
	e, ok := c.data[key]
	if !ok {
		c.mu.RUnlock()
		return -2
	}

	if isExpired(e) {
		c.mu.RUnlock()
		c.mu.Lock()
		if e2, ok := c.data[key]; ok && isExpired(e2) {
			delete(c.data, key)
		}
		c.mu.Unlock()
		return -2
	}

	if e.expiration.IsZero() {
		c.mu.RUnlock()
		return -1
	}

	remaining := time.Until(e.expiration)
	c.mu.RUnlock()
	return remaining
}

func (c *InMemoryCache) Exists(key string) bool {
	c.mu.RLock()
	e, ok := c.data[key]
	if !ok {
		c.mu.RUnlock()
		return false
	}

	if isExpired(e) {
		c.mu.RUnlock()
		c.mu.Lock()
		if e2, ok := c.data[key]; ok && isExpired(e2) {
			delete(c.data, key)
		}
		c.mu.Unlock()
		return false
	}

	c.mu.RUnlock()
	return true
}

func isExpired(e *entry) bool {
	return !e.expiration.IsZero() && time.Now().After(e.expiration)
}
