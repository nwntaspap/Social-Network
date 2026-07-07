package cache

import "sync"

type InMemoryCache struct {
	cache sync.Map
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{}
}
