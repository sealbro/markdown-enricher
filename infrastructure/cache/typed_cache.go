package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

const (
	methodLabelName = "method"
	prefixLabelName = "prefix"
)

var (
	cacheSuccessCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_request_success",
		Help: "The total number of success cache get",
	}, []string{prefixLabelName})
	cacheMissCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_request_miss",
		Help: "The total number of miss cache get",
	}, []string{prefixLabelName})
	cacheCallCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_request",
		Help: "The total number of cache calls",
	}, []string{prefixLabelName, methodLabelName})
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
	cacheCallCounter.WithLabelValues(s.prefix, "set").Inc()

	key = s.buildKey(key)

	return s.cache.Set(key, value, ttl)
}

func (s *TypedCache[V]) Get(key string) (V, error) {
	cacheCallCounter.WithLabelValues(s.prefix, "get").Inc()

	key = s.buildKey(key)

	valueAny, err := s.cache.Get(key)

	var value V
	if err == nil {
		cacheSuccessCounter.WithLabelValues(s.prefix).Inc()
		value = valueAny.(V)
	} else {
		cacheMissCounter.WithLabelValues(s.prefix).Inc()
	}

	return value, err
}

func (s *TypedCache[V]) GetOrSet(key string, getValue func() (V, error), ttl time.Duration) (V, error) {
	value, err := s.Get(key)

	if err != nil {
		value, err = getValue()
		if err != nil {
			return value, err
		}

		err = s.Set(key, value, ttl)
	}

	return value, err
}

func (s *TypedCache[V]) buildKey(key string) string {
	return s.prefix + "=>" + key
}
