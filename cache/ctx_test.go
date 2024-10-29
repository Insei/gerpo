package cache

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	Name string
}

func TestStorageGet(t *testing.T) {
	tests := []struct {
		name      string
		modelType reflect.Type
		cache     map[reflect.Type]map[string]any
		key       string
		disabled  []string
		wantValue any
		wantFound bool
	}{
		{
			name:      "Key is in disabled list",
			modelType: reflect.TypeOf(1),
			cache:     nil,
			key:       "disabledKey",
			disabled:  []string{"disabledKey"},
			wantValue: nil,
			wantFound: false,
		},
		{
			name:      "Model type not in cache",
			modelType: reflect.TypeOf(1),
			cache:     make(map[reflect.Type]map[string]any),
			key:       "someKey",
			disabled:  nil,
			wantValue: nil,
			wantFound: false,
		},
		{
			name:      "Key not in model cache",
			modelType: reflect.TypeOf(1),
			cache: map[reflect.Type]map[string]any{
				reflect.TypeOf(1): {},
			},
			key:       "someKey",
			disabled:  nil,
			wantValue: nil,
			wantFound: false,
		},
		{
			name:      "Key found in model cache",
			modelType: reflect.TypeOf(1),
			cache: map[reflect.Type]map[string]any{
				reflect.TypeOf(1): {"someKey": "someValue"},
			},
			key:       "someKey",
			disabled:  nil,
			wantValue: "someValue",
			wantFound: true,
		},
	}

	mtx := &sync.Mutex{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &cacheStorage{
				mtx:      mtx,
				c:        tt.cache,
				disabled: tt.disabled,
			}
			gotValue, gotFound := cs.Get(tt.modelType, tt.key)
			assert.Equal(t, tt.wantValue, gotValue)
			assert.Equal(t, tt.wantFound, gotFound)
		})
	}
}

func TestStorageSet(t *testing.T) {
	tests := []struct {
		name      string
		modelType reflect.Type
		initial   map[reflect.Type]map[string]any
		key       string
		value     any
		expected  map[reflect.Type]map[string]any
	}{
		{
			name:      "Set value for existing model type and new key",
			modelType: reflect.TypeOf(1),
			initial: map[reflect.Type]map[string]any{
				reflect.TypeOf(1): {"existingKey": "existingValue"},
			},
			key:   "newKey",
			value: "newValue",
			expected: map[reflect.Type]map[string]any{
				reflect.TypeOf(1): {
					"existingKey": "existingValue",
					"newKey":      "newValue",
				},
			},
		},
		{
			name:      "Set value for new model type",
			modelType: reflect.TypeOf("string"),
			initial:   map[reflect.Type]map[string]any{},
			key:       "newKey",
			value:     "newValue",
			expected: map[reflect.Type]map[string]any{
				reflect.TypeOf(""): {
					"newKey": "newValue",
				},
			},
		},
	}

	mtx := &sync.Mutex{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &cacheStorage{
				mtx: mtx,
				c:   tt.initial,
			}
			cs.Set(tt.modelType, tt.key, tt.value)
			assert.Equal(t, tt.expected, cs.c)
		})
	}
}

func TestCacheStorageClean(t *testing.T) {
	tests := []struct {
		name      string
		modelType reflect.Type
		initial   map[reflect.Type]map[string]any
		expected  map[reflect.Type]map[string]any
	}{
		{
			name:      "Clean existing model type cache",
			modelType: reflect.TypeOf(1),
			initial: map[reflect.Type]map[string]any{
				reflect.TypeOf(1): {
					"key1": "value1",
					"key2": "value2",
				},
			},
			expected: map[reflect.Type]map[string]any{
				reflect.TypeOf(1): {},
			},
		},
		{
			name:      "Clean non-existent model type cache",
			modelType: reflect.TypeOf("string"),
			initial: map[reflect.Type]map[string]any{
				reflect.TypeOf(1): {
					"key1": "value1",
				},
			},
			expected: map[reflect.Type]map[string]any{
				reflect.TypeOf(1):        {"key1": "value1"},
				reflect.TypeOf("string"): {},
			},
		},
	}

	mtx := &sync.Mutex{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &cacheStorage{
				mtx: mtx,
				c:   tt.initial,
			}
			cs.Clean(tt.modelType)
			assert.Equal(t, tt.expected, cs.c)
		})
	}
}

func TestNewCtxCache(t *testing.T) {
	// Create a base context
	baseCtx := context.Background()

	// Generate a new context with cacheStorage
	ctx := NewCtxCache(baseCtx)

	// Retrieve the cacheStorage from the context
	cache, ok := ctx.Value(contextCacheKey).(*cacheStorage)
	if !ok || cache == nil {
		t.Fatal("Expected cacheStorage in context, but got nil or wrong type")
	}

	// Verify the cacheStorage is properly initialized
	assert.NotNil(t, cache.mtx, "Expected non-nil mutex in cacheStorage")
	assert.NotNil(t, cache.c, "Expected empty map in cacheStorage")
}

func TestGetFromCtxCache(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "Cache does not exist in context",
			ctx:  context.Background(),
		},
		{
			name: "Model type not in cache",
			ctx:  NewCtxCache(context.Background()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetFromCtxCache[TestModel](tt.ctx, "key")
		})
	}
}

func TestAppendToCtxCache(t *testing.T) {
	tests := []struct {
		name       string
		storageCtx context.Context
	}{
		{
			name:       "Append to missing cacheStorage in context",
			storageCtx: context.WithValue(context.Background(), contextCacheKey, nil),
		},
		{
			name:       "Append to missing cacheStorage in context",
			storageCtx: context.Background(),
		},
		{
			name:       "Append to correct cacheStorage in context",
			storageCtx: NewCtxCache(context.Background()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AppendToCtxCache[TestModel](tt.storageCtx, "anyString", nil)
		})
	}
}

func TestCleanupCtxCache(t *testing.T) {
	tests := []struct {
		name     string
		ctxFunc  func() context.Context
		expected map[reflect.Type]map[string]any
	}{
		{
			name: "Clean existing model type cache",
			ctxFunc: func() context.Context {
				ctx := context.WithValue(context.Background(), contextCacheKey, &cacheStorage{
					mtx: &sync.Mutex{},
					c:   make(map[reflect.Type]map[string]any),
				})
				storage := ctx.Value(contextCacheKey).(*cacheStorage)
				storage.c[reflect.TypeOf((*TestModel)(nil)).Elem()] = map[string]any{
					"existingKey": TestModel{Name: "Existing"},
				}
				return ctx
			},
			expected: map[reflect.Type]map[string]any{
				reflect.TypeOf((*TestModel)(nil)).Elem(): {},
			},
		},
		{
			name: "Clean missing cacheStorage in context",
			ctxFunc: func() context.Context {
				return context.Background()
			},
			expected: map[reflect.Type]map[string]any{
				// Should be empty because cacheStorage is missing
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.ctxFunc()
			CleanupCtxCache[TestModel](ctx)
			storage, _ := ctx.Value(contextCacheKey).(*cacheStorage)
			if storage != nil {
				assert.Equal(t, tt.expected, storage.c)
			} else {
				assert.Empty(t, tt.expected, "Expected empty cache storage")
			}
		})
	}
}

func TestDisableKey(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		key      string
		expected []string
	}{
		{
			name:     "Disable key with existing cacheStorage",
			ctx:      context.WithValue(context.Background(), contextCacheKey, &cacheStorage{mtx: &sync.Mutex{}}),
			key:      "testKey",
			expected: []string{"testKey"},
		},
		{
			name:     "Disabled key without cacheStorage in context",
			ctx:      context.Background(),
			key:      "testKey",
			expected: nil, // No change since cacheStorage is not in context
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DisableCtxKey(tt.ctx, tt.key)
			storage, _ := tt.ctx.Value(contextCacheKey).(*cacheStorage)
			if storage != nil {
				assert.Equal(t, tt.expected, storage.disabled)
			} else {
				assert.Nil(t, tt.expected, "Expected nil since no cache storage should be present")
			}
		})
	}
}

func TestRemoveDisabledKey(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		key      string
		expected []string
	}{
		{
			name: "Remove existing disabled key",
			ctx: context.WithValue(context.Background(), contextCacheKey, &cacheStorage{
				mtx:      &sync.Mutex{},
				disabled: []string{"testKey1", "testKey2"},
			}),
			key:      "testKey1",
			expected: []string{"testKey2"},
		},
		{
			name: "Remove non-existing key",
			ctx: context.WithValue(context.Background(), contextCacheKey, &cacheStorage{
				mtx:      &sync.Mutex{},
				disabled: []string{"testKey1", "testKey2"},
			}),
			key:      "nonExistingKey",
			expected: []string{"testKey1", "testKey2"},
		},
		{
			name:     "Remove key without cacheStorage in context",
			ctx:      context.Background(),
			key:      "testKey",
			expected: nil, // No change since cacheStorage is not in context
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnableCtxKey(tt.ctx, tt.key)
			storage, _ := tt.ctx.Value(contextCacheKey).(*cacheStorage)
			if storage != nil {
				assert.Equal(t, tt.expected, storage.disabled)
			} else {
				assert.Nil(t, tt.expected, "Expected nil since no cache storage should be present")
			}
		})
	}
}
