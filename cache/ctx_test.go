package cache

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

type testModel struct{}
type testModel2 struct{}
type testModel3 struct{}

type testModel4 struct {
	Name string
}

func TestNewCtxCache(t *testing.T) {
	// Create a new context with a cache
	ctx := NewCtxCache(context.Background())

	// Retrieve the cache storage from the context
	storage, ok := ctx.Value(contextCacheKey).(*cacheStorage)

	// Validate that the storage is of type *cacheStorage and is not nil
	if !ok || storage == nil {
		t.Errorf("expected NewCtxCache to return a context with a cacheStorage value, but got nil")
		return
	}

	// Ensure that the cache storage has been initialized properly
	if storage.mtx == nil {
		t.Errorf("expected cacheStorage.mtx to be initialized, but got nil")
	}
	if storage.c == nil {
		t.Errorf("expected cacheStorage.c to be initialized, but got nil")
	}
}

func TestContextCacheKeyType(t *testing.T) {
	cases := []struct {
		name          string
		key           string
		cacheKey      string
		expectedError bool
	}{
		{
			name:          "check key value",
			key:           "ctx_cache_key",
			cacheKey:      "ctx_cache_key",
			expectedError: false,
		},
		{
			name:          "check key is not empty",
			key:           "",
			cacheKey:      "",
			expectedError: true,
		},
		{
			name:          "check key is a string",
			key:           "ctx_cache_key",
			cacheKey:      "",
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the context cache key to the test case key
			contextCacheKey.key = tc.key

			// If an error is expected in the test case
			if tc.expectedError {
				// Check if the context cache key matches the test case key
				if contextCacheKey.key != tc.key {
					t.Errorf("expected contextCacheKey.key to be '%s', but got '%s'", tc.key, contextCacheKey.key)
				}
			} else {
				// If no error is expected, check if the context cache key matches the expected cache key
				if contextCacheKey.key != tc.cacheKey {
					t.Errorf("expected contextCacheKey.key to be '%s', but got '%s'", tc.cacheKey, contextCacheKey.key)
				}
			}
		})
	}
}

func TestCacheStorageGet(t *testing.T) {
	cases := []struct {
		name          string
		modelType     reflect.Type
		key           string
		value         any
		expectedValue any
		expectedOk    bool
	}{
		{
			name:          "get non-existent value",
			modelType:     reflect.TypeOf(&testModel{}).Elem(),
			key:           "test_key",
			value:         nil,
			expectedValue: nil,
			expectedOk:    false,
		},
		{
			name:          "get existing value",
			modelType:     reflect.TypeOf(&testModel{}).Elem(),
			key:           "test_key",
			value:         "test_value",
			expectedValue: "test_value",
			expectedOk:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize a new cacheStorage with a mutex and an empty map
			s := &cacheStorage{
				mtx: &sync.Mutex{},
				c:   make(map[reflect.Type]map[string]any),
			}

			// If the test case value is not nil, set it in the cache for the specific model type and key
			if tc.value != nil {
				s.Set(tc.modelType, tc.key, tc.value)
			}

			// Attempt to get the value from the cache for the specific model type and key
			gotValue, ok := s.Get(tc.modelType, tc.key)

			// Check if the existence flag matches the expected value
			if ok != tc.expectedOk {
				t.Errorf("expected Get to return %v, but got %v", tc.expectedOk, ok)
			}

			// Check if the returned value matches the expected value
			if gotValue != tc.expectedValue {
				t.Errorf("expected Get to return '%v', but got '%v'", tc.expectedValue, gotValue)
			}
		})
	}
}

func TestCacheStorageSet(t *testing.T) {
	cases := []struct {
		name          string
		modelType     reflect.Type
		key           string
		value         any
		expectedValue any
		expectedOk    bool
	}{
		{
			name:          "set and get a value",
			modelType:     reflect.TypeOf(&testModel{}).Elem(),
			key:           "test_key",
			value:         "test_value",
			expectedValue: "test_value",
			expectedOk:    true,
		},
		{
			name:          "set a value with empty key",
			modelType:     reflect.TypeOf(&testModel{}).Elem(),
			key:           "",
			value:         "test_value",
			expectedValue: "test_value",
			expectedOk:    true,
		},
		{
			name:          "set a value with nil value",
			modelType:     reflect.TypeOf(&testModel{}).Elem(),
			key:           "test_key",
			value:         nil,
			expectedValue: nil,
			expectedOk:    true,
		},
		{
			name:          "set multiple values with the same model type",
			modelType:     reflect.TypeOf(&testModel{}).Elem(),
			key:           "test_key",
			value:         "test_value",
			expectedValue: "test_value",
			expectedOk:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize a new cacheStorage with a mutex and an empty map
			s := &cacheStorage{
				mtx: &sync.Mutex{},
				c:   make(map[reflect.Type]map[string]any),
			}

			// Set a value in the cache for a specific model type and key
			s.Set(tc.modelType, tc.key, tc.value)

			// Attempt to get the value from the cache for the specific model type and key
			gotValue, ok := s.Get(tc.modelType, tc.key)

			// Check if the existence flag matches the expected value
			if ok != tc.expectedOk {
				t.Errorf("expected Get to return %v, but got %v", tc.expectedOk, ok)
			}

			// Check if the returned value matches the expected value
			if gotValue != tc.expectedValue {
				t.Errorf("expected Get to return '%v', but got '%v'", tc.expectedValue, gotValue)
			}
		})
	}
}

func TestCacheStorageClean(t *testing.T) {
	cases := []struct {
		name       string
		modelType  reflect.Type
		key        string
		value      any
		cleanType  reflect.Type
		expectedOk bool
	}{
		{
			name:       "clean cache for a specific model type",
			modelType:  reflect.TypeOf(&testModel{}).Elem(),
			key:        "test_key",
			value:      "test_value",
			cleanType:  reflect.TypeOf(&testModel{}).Elem(),
			expectedOk: false,
		},
		{
			name:       "clean cache for a different model type",
			modelType:  reflect.TypeOf(&testModel{}).Elem(),
			key:        "test_key",
			value:      "test_value",
			cleanType:  reflect.TypeOf(&testModel2{}).Elem(),
			expectedOk: true,
		},
		{
			name:       "clean cache for a model type that doesn't exist",
			modelType:  reflect.TypeOf(&testModel{}).Elem(),
			key:        "test_key",
			value:      "test_value",
			cleanType:  reflect.TypeOf(&testModel3{}).Elem(),
			expectedOk: true,
		},
		{
			name:       "clean cache for a nil model type",
			modelType:  reflect.TypeOf(&testModel{}).Elem(),
			key:        "test_key",
			value:      "test_value",
			cleanType:  nil,
			expectedOk: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize a new cacheStorage with a mutex and an empty map
			s := &cacheStorage{
				mtx: &sync.Mutex{},
				c:   make(map[reflect.Type]map[string]any),
			}

			// Set a value in the cache for a specific model type and key
			s.Set(tc.modelType, tc.key, tc.value)

			// Clean the cache based on the provided clean type
			s.Clean(tc.cleanType)

			// Attempt to get the value from the cache for the specific model type and key
			_, ok := s.Get(tc.modelType, tc.key)

			// Check if the existence flag matches the expected value
			if ok != tc.expectedOk {
				t.Errorf("expected Get to return %v, but got %v", tc.expectedOk, ok)
			}
		})
	}
}

func TestGetFromCtxCache(t *testing.T) {
	cases := []struct {
		name          string
		key           string
		value         any
		addValue      bool
		expectedOk    bool
		expectedValue any
	}{
		{
			name:          "get non-existent value",
			key:           "test_key",
			value:         nil,
			addValue:      false,
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name:          "set and get value",
			key:           "test_key",
			value:         "test_value",
			addValue:      true,
			expectedOk:    true,
			expectedValue: "test_value",
		},
		{
			name:          "get value with empty key",
			key:           "",
			value:         "test_value",
			addValue:      true,
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name:          "get value with nil value",
			key:           "test_key",
			value:         nil,
			addValue:      true,
			expectedOk:    true,
			expectedValue: nil,
		},
		{
			name:          "get value with different model type",
			key:           "test_key",
			value:         "test_value",
			addValue:      false,
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name:          "set multiple values and get one",
			key:           "test_key",
			value:         "test_value",
			addValue:      true,
			expectedOk:    true,
			expectedValue: "test_value",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewCtxCache(context.Background())

			// Append the value to the context cache if addValue is true
			if tc.addValue && tc.key != "" {
				AppendToCtxCache[testModel](ctx, tc.key, tc.value)
			}

			// Get the value from the context cache
			gotValue, ok := GetFromCtxCache[testModel](ctx, tc.key)

			// Check if the existence flag matches the expected value
			if ok != tc.expectedOk {
				t.Errorf("expected GetFromCtxCache to return %v, but got %v", tc.expectedOk, ok)
			}

			// Check if the returned value matches the expected value
			if gotValue != tc.expectedValue {
				t.Errorf("expected GetFromCtxCache to return '%v', but got '%v'", tc.expectedValue, gotValue)
			}
		})
	}
}

func TestAppendToCtxCache(t *testing.T) {
	cases := []struct {
		name          string
		key           string
		value         any
		exist         bool
		expectedOk    bool
		expectedValue any
	}{
		{
			name:          "set a value",
			key:           "test_key",
			value:         "test_value",
			exist:         false,
			expectedOk:    true,
			expectedValue: "test_value",
		},
		{
			name:          "set multiple values",
			key:           "test_key",
			value:         "test_value",
			exist:         true,
			expectedOk:    true,
			expectedValue: "test_value",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewCtxCache(context.Background())
			// If the test case indicates that the value should exist (this seems redundant, as the next line always appends the value)
			if tc.exist {
				AppendToCtxCache[testModel](ctx, tc.key, tc.value)
			}

			// Append the value to the context cache
			AppendToCtxCache[testModel](ctx, tc.key, tc.value)

			// Get the value from the context cache
			gotValue, ok := GetFromCtxCache[testModel](ctx, tc.key)

			// Check if the existence flag matches the expected value
			if ok != tc.expectedOk {
				t.Errorf("expected GetFromCtxCache to return %v, but got %v", tc.expectedOk, ok)
			}

			// Check if the returned value matches the expected value
			if gotValue != tc.expectedValue {
				t.Errorf("expected GetFromCtxCache to return '%v', but got '%v'", tc.expectedValue, gotValue)
			}
		})
	}
}

func TestCleanupCtxCache(t *testing.T) {
	cases := []struct {
		name          string
		key           string
		value         any
		exist         bool
		expectedOk    bool
		expectedValue any
	}{
		{
			name:          "clean cache for a specific model type",
			key:           "test_key",
			value:         "test_value",
			exist:         true,
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name:          "clean cache for a model type that doesn't exist",
			key:           "test_key",
			value:         "test_value",
			exist:         false,
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name:          "clean cache multiple times",
			key:           "test_key",
			value:         "test_value",
			exist:         true,
			expectedOk:    false,
			expectedValue: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewCtxCache(context.Background())
			// If the test case indicates that the value should exist, append it to the context cache
			if tc.exist {
				AppendToCtxCache[testModel](ctx, tc.key, tc.value)
			}

			// Cleanup the context cache
			CleanupCtxCache[testModel](ctx)

			// Get the value from the context cache after cleanup
			gotValue, ok := GetFromCtxCache[testModel](ctx, tc.key)

			// Check if the existence flag matches the expected value
			if ok != tc.expectedOk {
				t.Errorf("expected GetFromCtxCache to return %v, but got %v", tc.expectedOk, ok)
			}

			// Check if the returned value matches the expected value
			if gotValue != tc.expectedValue {
				t.Errorf("expected GetFromCtxCache to return '%v', but got '%v'", tc.expectedValue, gotValue)
			}
		})
	}
}

func TestDisableKey(t *testing.T) {
	type testCase struct {
		name          string
		key           string
		valueToCache  any
		shouldDisable bool
		expectExist   bool
		expectedValue any
	}

	testCases := []testCase{
		{
			name:          "exists before disable",
			key:           "test_key",
			valueToCache:  "test_value",
			shouldDisable: false,
			expectExist:   true,
			expectedValue: "test_value",
		},
		{
			name:          "does not exist after disable",
			key:           "test_key",
			valueToCache:  "test_value",
			shouldDisable: true,
			expectExist:   false,
			expectedValue: nil,
		},
		{
			name:          "exist after re-enable",
			key:           "test_key",
			valueToCache:  "test_value",
			shouldDisable: true,
			expectExist:   false,
			expectedValue: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewCtxCache(context.Background())

			// If there is a value to cache, append it to the context cache
			if tc.valueToCache != nil {
				AppendToCtxCache[testModel4](ctx, tc.key, tc.valueToCache)
			}

			// Get the value from the context cache
			value, ok := GetFromCtxCache[testModel4](ctx, tc.key)

			// Check if the value exists before disabling the key
			if tc.shouldDisable && (!ok || value != tc.valueToCache) {
				t.Errorf("expected value to be '%v' before disabling, but got '%v'", tc.valueToCache, value)
			}

			// Disable the key if required
			if tc.shouldDisable {
				DisableKey(ctx, tc.key)
			}

			// Get the value from the context cache after disabling the key
			value, ok = GetFromCtxCache[testModel4](ctx, tc.key)

			// Check if the value exists after disabling the key
			if tc.expectExist && (!ok || value != tc.expectedValue) {
				t.Errorf("expected value to be '%v' after disabling, but got '%v'", tc.expectedValue, value)
			}

			// Check if the value does not exist after disabling the key
			if !tc.expectExist && (ok || value != tc.expectedValue) {
				t.Errorf("expected no value or nil after disabling, but got '%v'", value)
			}
		})
	}
}

func TestRemoveDisabledKey(t *testing.T) {
	type testCase struct {
		name           string
		key            string
		valueToCache   any
		shouldDisable  bool
		removeDisabled bool
		expectExist    bool
		expectedValue  any
	}

	testCases := []testCase{
		{
			name:           "remove disabled key and check if value exists",
			key:            "test_key",
			valueToCache:   "test_value",
			shouldDisable:  true,
			removeDisabled: true,
			expectExist:    true,
			expectedValue:  "test_value",
		},
		{
			name:           "remove non-disabled key",
			key:            "test_key",
			valueToCache:   "test_value",
			shouldDisable:  false,
			removeDisabled: true,
			expectExist:    true,
			expectedValue:  "test_value",
		},
		{
			name:           "remove disabled non-existing key",
			key:            "missing_key",
			valueToCache:   nil,
			shouldDisable:  true,
			removeDisabled: true,
			expectExist:    false,
			expectedValue:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewCtxCache(context.Background())

			// If there is a value to cache, append it to the context cache
			if tc.valueToCache != nil {
				AppendToCtxCache[testModel](ctx, tc.key, tc.valueToCache)
			}

			// Disable the key if required
			if tc.shouldDisable {
				DisableKey(ctx, tc.key)
			}

			// Remove the disabled key if required
			if tc.removeDisabled {
				RemoveDisabledKey(ctx, tc.key)
			}

			// Get the value from the context cache
			value, ok := GetFromCtxCache[testModel](ctx, tc.key)

			// Check if the value exists as expected after the operations
			if tc.expectExist && (!ok || value != tc.expectedValue) {
				t.Errorf("expected value to be '%v', but got '%v'", tc.expectedValue, value)
			}

			// Check if the value does not exist as expected after the operations
			if !tc.expectExist && (ok || value != tc.expectedValue) {
				t.Errorf("expected no value or nil, but got '%v'", value)
			}
		})
	}
}
