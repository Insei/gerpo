package ctx

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/insei/gerpo/cache/types"
	"github.com/stretchr/testify/assert"
)

func TestStorageGet(t *testing.T) {
	tests := []struct {
		name      string
		modelKey  string
		cache     map[string]map[string]any
		key       string
		disabled  []string
		wantValue any
		wantErr   error
	}{
		{
			name:      "Key is in disabled list",
			modelKey:  "any",
			cache:     nil, //TODO: add cache with disabledKey modelKey
			key:       "disabledKey",
			disabled:  []string{"disabledKey"},
			wantValue: nil,
			wantErr:   types.ErrNotFound,
		},
		{
			name:      "Model type not in cache",
			modelKey:  "any",
			cache:     make(map[string]map[string]any),
			key:       "someKey",
			disabled:  nil,
			wantValue: nil,
			wantErr:   types.ErrNotFound,
		},
		{
			name:     "Key not in model cache",
			modelKey: "modelKey",
			cache: map[string]map[string]any{
				"modelKey": {},
			},
			key:       "cacheKey",
			disabled:  nil,
			wantValue: nil,
			wantErr:   types.ErrNotFound,
		},
		{
			name:     "Key found in model cache",
			modelKey: "modelKey",
			cache: map[string]map[string]any{
				"modelKey": {"someKey": "someValue"},
			},
			key:       "someKey",
			disabled:  nil,
			wantValue: "someValue",
			wantErr:   nil,
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
			gotValue, gotFound := cs.Get(tt.modelKey, tt.key)
			assert.Equal(t, tt.wantValue, gotValue)
			assert.Equal(t, true, errors.Is(gotFound, tt.wantErr))
		})
	}
}

func TestStorageSet(t *testing.T) {
	tests := []struct {
		name     string
		modelKey string
		initial  map[string]map[string]any
		key      string
		value    any
		expected map[string]map[string]any
	}{
		{
			name:     "Set value for existing model key and new key",
			modelKey: "modelKey",
			initial: map[string]map[string]any{
				"modelKey": {"existingKey": "existingValue"},
			},
			key:   "newKey",
			value: "newValue",
			expected: map[string]map[string]any{
				"modelKey": {
					"existingKey": "existingValue",
					"newKey":      "newValue",
				},
			},
		},
		{
			name:     "Set value for new model type",
			modelKey: "modelKey",
			initial:  map[string]map[string]any{},
			key:      "newKey",
			value:    "newValue",
			expected: map[string]map[string]any{
				"modelKey": {
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
			cs.Set(tt.modelKey, tt.key, tt.value)
			assert.Equal(t, tt.expected, cs.c)
		})
	}
}

func TestCacheStorageClean(t *testing.T) {
	tests := []struct {
		name     string
		modelKey string
		initial  map[string]map[string]any
		expected map[string]map[string]any
	}{
		{
			name:     "Clean existing model type cache",
			modelKey: "modelKey",
			initial: map[string]map[string]any{
				"modelKey": {
					"key1": "value1",
					"key2": "value2",
				},
			},
			expected: map[string]map[string]any{
				"modelKey": {},
			},
		},
		{
			name:     "Clean non-existent model type cache",
			modelKey: "newButCleaned",
			initial: map[string]map[string]any{
				"modelKey": {
					"key1": "value1",
				},
			},
			expected: map[string]map[string]any{
				"modelKey":      {"key1": "value1"},
				"newButCleaned": {},
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
			cs.Clean(tt.modelKey)
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
	cache, ok := ctx.Value(ctxCacheKey).(*cacheStorage)
	if !ok || cache == nil {
		t.Fatal("Expected cacheStorage in context, but got nil or wrong type")
	}

	// Verify the cacheStorage is properly initialized
	assert.NotNil(t, cache.mtx, "Expected non-nil mutex in cacheStorage")
	assert.NotNil(t, cache.c, "Expected empty map in cacheStorage")
}
