package cache

import "time"

type Cache interface {
	Set(key string, value any, ttl time.Duration)
	SetNX(key string, value any, ttl time.Duration)

	Get(key string) (any, error)
	Delete(key string) error

	Expire(key string, ttl time.Duration) error
	TTL(key string) (time.Duration, error)
	Exists(key string) (bool, error)
}
