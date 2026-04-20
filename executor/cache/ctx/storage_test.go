package ctx

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/insei/gerpo/executor/cache/types"
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

// TestCacheStorage_Concurrent — covers the mutex contract on the three entry
// points. With -race the test fails immediately if any code path touches the
// map outside the lock.
func TestCacheStorage_Concurrent(t *testing.T) {
	cs := &cacheStorage{
		mtx: &sync.Mutex{},
		c:   make(map[string]map[string]any),
	}

	const writers, readers, cleaners, iterations = 8, 8, 2, 500
	var wg sync.WaitGroup

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			model := "repo-" + string(rune('A'+id%writers))
			for i := 0; i < iterations; i++ {
				cs.Set(model, "sql-"+string(rune(i%26+'a')), i)
			}
		}(w)
	}

	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			model := "repo-" + string(rune('A'+id%writers))
			for i := 0; i < iterations; i++ {
				_, _ = cs.Get(model, "sql-"+string(rune(i%26+'a')))
			}
		}(r)
	}

	for c := 0; c < cleaners; c++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations/10; i++ {
				cs.Clean()
			}
		}()
	}

	wg.Wait()
	// No assertions on the final state — the mix of writers/cleaners leaves it
	// non-deterministic. The point is: the race detector stays quiet.
}

func TestCacheStorageClean_WipesAllModels(t *testing.T) {
	cs := &cacheStorage{
		mtx: &sync.Mutex{},
		c: map[string]map[string]any{
			"users":    {"select users": "row-1"},
			"posts":    {"select posts": []any{"row-a"}},
			"comments": {"count": uint64(7)},
		},
	}
	cs.Clean()
	assert.Empty(t, cs.c,
		"any write must invalidate every repository's cache in the current context "+
			"— cross-repo dependencies (JOINs, virtual columns) make a per-repo clean unsafe")
}

func TestNewCtxCache(t *testing.T) {
	// Create a base context
	baseCtx := context.Background()

	// Generate a new context with cacheStorage
	ctx := WrapContext(baseCtx)

	// Retrieve the cacheStorage from the context
	cache, ok := ctx.Value(ctxCacheKey).(*cacheStorage)
	if !ok || cache == nil {
		t.Fatal("Expected cacheStorage in context, but got nil or wrong type")
	}

	// Verify the cacheStorage is properly initialized
	assert.NotNil(t, cache.mtx, "Expected non-nil mutex in cacheStorage")
	assert.NotNil(t, cache.c, "Expected empty map in cacheStorage")
}
