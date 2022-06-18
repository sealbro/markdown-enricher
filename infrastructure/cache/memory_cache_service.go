package cache

import (
	"context"
	"errors"
	"github.com/jellydator/ttlcache/v3"
	"markdown-enricher/pkg/closer"
	"time"
)

type MemoryCacheService struct {
	cache *ttlcache.Cache[string, any]
}

func MakeMemoryCacheService(collection *closer.CloserCollection) CacheService {
	cache := ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](24 * time.Hour),
	)

	go cache.Start()

	service := &MemoryCacheService{
		cache: cache,
	}

	collection.Add(service)

	return service
}

func (s *MemoryCacheService) Close(ctx context.Context) error {
	s.cache.Stop()
	return nil
}

func (s *MemoryCacheService) Set(key string, value any, ttl time.Duration) error {
	s.cache.Set(key, value, ttl)

	return nil
}

func (s *MemoryCacheService) Get(key string) (any, error) {
	get := s.cache.Get(key)

	if get == nil {
		return nil, errors.New("key not found")
	}

	return get.Value(), nil
}
