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

func (s *cacheStorage) Get(modelKey string, key string) (any, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if slices.Contains(s.disabled, key) {
		return nil, types.ErrNotFound
	}
	modelCache, ok := s.c[modelKey]
	if !ok {
		return nil, types.ErrNotFound
	}
	cached, ok := modelCache[key]
	if !ok {
		return nil, types.ErrNotFound
	}
	return cached, nil
}

func (s *cacheStorage) Set(modelKey string, key string, value any) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	modelCache, ok := s.c[modelKey]
	if !ok {
		modelCache = make(map[string]any)
		s.c[modelKey] = modelCache
	}
	modelCache[key] = value
}

// Clean wipes every model's cache section inside this context. Per-repo
// isolation is only used for Get/Set (avoiding accidental key collisions
// between repositories that happen to encode the same SQL); invalidation
// fans out to the whole request context because cross-repo dependencies —
// virtual columns, JOINs — make per-repo clean unsafe. See docs/features/cache.md
// for the rationale.
func (s *cacheStorage) Clean() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for k := range s.c {
		delete(s.c, k)
	}
}

func WrapContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxCacheKey, &cacheStorage{mtx: &sync.Mutex{}, c: make(map[string]map[string]any)})
}
