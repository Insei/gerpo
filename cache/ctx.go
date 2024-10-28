package cache

import (
	"context"
	"reflect"
	"slices"
	"sync"
)

type contextCacheKeyType struct {
	key string
}

var contextCacheKey = &contextCacheKeyType{
	key: "ctx_cache_key",
}

type cacheStorage struct {
	mtx      *sync.Mutex
	c        map[reflect.Type]map[string]any
	disabled []string
}

// Get returns the value for the given key in the cache.
func (s *cacheStorage) Get(modelType reflect.Type, key string) (any, bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if slices.Contains(s.disabled, key) {
		return nil, false
	}
	modelCache, ok := s.c[modelType]
	if !ok {
		return nil, false
	}
	cached, ok := modelCache[key]
	if !ok {
		return nil, false
	}
	return cached, true
}

// Set sets the value for the given key in the cache.
func (s *cacheStorage) Set(modelType reflect.Type, key string, value any) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	modelCache, ok := s.c[modelType]
	if !ok {
		modelCache = make(map[string]any)
		s.c[modelType] = modelCache
	}
	modelCache[key] = value
}

// Clean removes all entries for the given model type from the cache.
func (s *cacheStorage) Clean(modelType reflect.Type) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.c[modelType] = make(map[string]any)
}

// NewCtxCache creates a new context with a cache.
func NewCtxCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextCacheKey, &cacheStorage{mtx: &sync.Mutex{}, c: make(map[reflect.Type]map[string]any)})
}

// GetFromCtxCache returns the value for the given key in the cache.
func GetFromCtxCache[TModel any](ctx context.Context, key string) (any, bool) {
	storage, ok := ctx.Value(contextCacheKey).(*cacheStorage)
	if !ok || storage == nil {
		return nil, false
	}
	var model *TModel
	mTypeOf := reflect.TypeOf(model).Elem()
	return storage.Get(mTypeOf, key)
}

// AppendToCtxCache sets the value for the given key in the cache.
func AppendToCtxCache[TModel any](ctx context.Context, key string, value any) {
	storage, ok := ctx.Value(contextCacheKey).(*cacheStorage)
	if !ok || storage == nil {
		return
	}
	var model *TModel
	mTypeOf := reflect.TypeOf(model).Elem()
	storage.Set(mTypeOf, key, value)
}

// CleanupCtxCache removes all entries for the given model type from the cache.
func CleanupCtxCache[TModel any](ctx context.Context) {
	storage, ok := ctx.Value(contextCacheKey).(*cacheStorage)
	if !ok || storage == nil {
		return
	}
	var model *TModel
	mTypeOf := reflect.TypeOf(model).Elem()
	storage.Clean(mTypeOf)
}

// DisableCtxKey disables key for context cache reading
func DisableCtxKey(ctx context.Context, key string) {
	storage, ok := ctx.Value(contextCacheKey).(*cacheStorage)
	if !ok || storage == nil {
		return
	}
	storage.mtx.Lock()
	defer storage.mtx.Unlock()
	storage.disabled = append(storage.disabled, key)
}

// RemoveCtxDisabledKey removes disabled for context key caching reading usage
func RemoveCtxDisabledKey(ctx context.Context, key string) {
	storage, ok := ctx.Value(contextCacheKey).(*cacheStorage)
	if !ok || storage == nil {
		return
	}
	storage.mtx.Lock()
	defer storage.mtx.Unlock()
	storage.disabled = slices.DeleteFunc(storage.disabled, func(s string) bool {
		if s == key {
			return true
		}
		return false
	})
}
