package cache

import "time"

type CacheService interface {
	Set(key string, value any, ttl time.Duration) error
	Get(key string) (any, error)
}
