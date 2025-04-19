package ctx

import (
	"context"
	"slices"
	"sync"

	"github.com/insei/gerpo/executor/cache/types"
)

type ctxCacheKeyType struct {
	key string
}

var ctxCacheKey = &ctxCacheKeyType{
	key: "ctx_cache_key",
}

type cacheStorage struct {
	mtx      *sync.Mutex
	c        map[string]map[string]any
	disabled []string
}

func (s *cacheStorage) Get(sourceKey string, key string) (any, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if slices.Contains(s.disabled, key) {
		return nil, types.ErrNotFound
	}
	modelCache, ok := s.c[sourceKey]
	if !ok {
		return nil, types.ErrNotFound
	}
	cached, ok := modelCache[key]
	if !ok {
		return nil, types.ErrNotFound
	}
	return cached, nil
}

func (s *cacheStorage) Set(sourceKey string, key string, value any) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	modelCache, ok := s.c[sourceKey]
	if !ok {
		modelCache = make(map[string]any)
		s.c[sourceKey] = modelCache
	}
	modelCache[key] = value
}

func (s *cacheStorage) Clean(sourceKey string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.c[sourceKey] = make(map[string]any)
}

func NewCtxCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxCacheKey, &cacheStorage{mtx: &sync.Mutex{}, c: make(map[string]map[string]any)})
}
