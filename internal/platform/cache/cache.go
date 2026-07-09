package cache

import "time"

type Cache interface {
	// set with ttl=0 if no expiration wanted
	Set(key string, value any, ttl time.Duration)
	// set a value only if it doesnt already exist
	// with a normal Set(...) it would overwrite the existsing value for the respective key
	SetNX(key string, value any, ttl time.Duration)

	Get(key string) (any, error)
	Delete(key string)

	//this method is used to define the expiration time of a value already
	//present in the cache.
	//again, use ttl=0 in order to make the value live forever
	Expire(key string, ttl time.Duration) error
	//in order to minimize error handling
	//returns -2 if key does not exist
	//returns -1 if there is no expiration set
	//return the remaing time to live otherwise
	TTL(key string) time.Duration
	Exists(key string) bool
	Stop()
}
