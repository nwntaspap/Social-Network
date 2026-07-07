package cache

import "errors"

var (
	ErrKeyNotFound = errors.New("cache: key not found")
	ErrNoTTL       = errors.New("cache: key has no ttl")
)
