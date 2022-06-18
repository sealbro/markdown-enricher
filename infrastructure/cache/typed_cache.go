package cache

import (
	"time"
)

type TypedCache[V any] struct {
	cache  CacheService
	prefix string
}

func NewTypedCache[V any](cache CacheService, prefix string) *TypedCache[V] {
	return &TypedCache[V]{
		cache:  cache,
		prefix: prefix,
	}
}

func (s *TypedCache[V]) Set(key string, value V, ttl time.Duration) error {
	key = s.prefix + key
	return s.cache.Set(key, value, ttl)
}

func (s *TypedCache[V]) Get(key string) (V, error) {
	valueAny, err := s.cache.Get(key)

	var value V
	if err == nil {
		value = valueAny.(V)
	}

	return value, err
}

func (s *TypedCache[V]) GetOrSet(key string, getValue func() (V, error), ttl time.Duration) (V, error) {
	key = s.prefix + key
	var value V
	valueAny, err := s.cache.Get(key)

	if err != nil {
		value, err = getValue()
		if err != nil {
			return value, err
		}

		err = s.cache.Set(key, value, ttl)
	} else {
		value = valueAny.(V)
	}

	return value, err
}
