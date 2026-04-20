package ctx

import (
	"context"
	"testing"

	"github.com/insei/gerpo/executor/cache/types"
	"github.com/insei/gerpo/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewSource(t *testing.T) {
	// Drill-down tests for ctx.New
	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "default options",
			opts: []Option{},
		},
		{
			name: "custom logger",
			opts: []Option{WithLogger(logger.NoopLogger)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := New(tt.opts...)
			assert.NotNil(t, src)
		})
	}
}

func TestGetStorage(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "nil Context",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "Context without storage",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "Context with storage",
			ctx:     context.WithValue(context.Background(), ctxCacheKey, &cacheStorage{}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Cache{log: logger.NoopLogger}
			_, err := s.getStorage(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		expectedErr error
	}{
		{
			name:        "Nil context",
			ctx:         nil,
			expectedErr: types.ErrNotFound,
		},
		{
			name:        "Context without storage",
			ctx:         context.Background(),
			expectedErr: types.ErrWrongConfiguration,
		},
		{
			name:        "Valid context with storage",
			ctx:         WrapContext(context.Background()),
			expectedErr: types.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Cache{log: logger.NoopLogger}
			_, err := s.Get(tt.ctx, "someKey", "someStatement")
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		modelKey      string
		cache         any
		statement     string
		statementArgs []any
	}{
		{
			name:          "Nil context",
			ctx:           nil,
			modelKey:      "testKey",
			cache:         "fakeCache",
			statement:     "setCache",
			statementArgs: []any{"arg1", "arg2"},
		},
		{
			name:          "OK",
			ctx:           WrapContext(context.Background()),
			modelKey:      "testKey",
			cache:         "fakeCache",
			statement:     "setCache",
			statementArgs: []any{"arg1", "arg2"},
		},
		// add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Cache{log: logger.NoopLogger, key: tt.modelKey}
			s.Set(tt.ctx, tt.cache, tt.statement, tt.statementArgs...)
		})
	}
}
func TestClean(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "nil Context",
			ctx:  nil,
		},
		{
			name: "Context without storage",
			ctx:  context.Background(),
		},
		{
			name: "Context with storage",
			ctx:  WrapContext(context.Background()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Cache{log: logger.NoopLogger, key: "testKey"}
			s.Clean(tt.ctx)
			// Since method doesn't return anything, no assertions made.
		})
	}
}

// TestClean_CrossRepoInvalidation — the public contract: a write through any
// Cache instance invalidates every other Cache that shares the same context.
// This is the fix for the cross-repo staleness gap: usersRepo.Update() used to
// leave postsRepo's cached JOIN result in place, now it wipes everything.
func TestClean_CrossRepoInvalidation(t *testing.T) {
	users := &Cache{log: logger.NoopLogger, key: "users"}
	posts := &Cache{log: logger.NoopLogger, key: "posts"}
	ctx := WrapContext(context.Background())

	users.Set(ctx, "user-row", "SELECT * FROM users WHERE id=?", 1)
	posts.Set(ctx,
		[]any{"post-1", "post-2"},
		"SELECT posts.* FROM posts JOIN users ON ...", 1,
	)

	// Sanity: both repositories see their cached value.
	got, err := users.Get(ctx, "SELECT * FROM users WHERE id=?", 1)
	assert.NoError(t, err)
	assert.Equal(t, "user-row", got)

	got, err = posts.Get(ctx, "SELECT posts.* FROM posts JOIN users ON ...", 1)
	assert.NoError(t, err)
	assert.Equal(t, []any{"post-1", "post-2"}, got)

	// A write through the users repo must purge the posts cache too: the JOIN
	// made posts depend on users' state, and gerpo cannot statically know which
	// caches are safe to keep.
	users.Clean(ctx)

	_, err = users.Get(ctx, "SELECT * FROM users WHERE id=?", 1)
	assert.ErrorIs(t, err, types.ErrNotFound, "writing repo must see its own cache wiped")

	_, err = posts.Get(ctx, "SELECT posts.* FROM posts JOIN users ON ...", 1)
	assert.ErrorIs(t, err, types.ErrNotFound,
		"other repositories on the same context must see their cache wiped too")
}

// TestClean_DifferentContextsIsolated — invalidation stays inside the
// originating context. Two independent requests never bleed into each other.
func TestClean_DifferentContextsIsolated(t *testing.T) {
	cache := &Cache{log: logger.NoopLogger, key: "users"}
	ctxA := WrapContext(context.Background())
	ctxB := WrapContext(context.Background())

	cache.Set(ctxA, "A", "SELECT ?", 1)
	cache.Set(ctxB, "B", "SELECT ?", 1)

	cache.Clean(ctxA)

	_, err := cache.Get(ctxA, "SELECT ?", 1)
	assert.ErrorIs(t, err, types.ErrNotFound)

	got, err := cache.Get(ctxB, "SELECT ?", 1)
	assert.NoError(t, err)
	assert.Equal(t, "B", got, "ctxB must not be affected by a clean on ctxA")
}

// TestGetSet_NamespaceIsolation — two repositories whose SQL keys collide must
// still return their own cached values. Per-repo namespacing in Set/Get keeps
// this property even after Clean dropped its per-repo granularity.
func TestGetSet_NamespaceIsolation(t *testing.T) {
	users := &Cache{log: logger.NoopLogger, key: "users"}
	posts := &Cache{log: logger.NoopLogger, key: "posts"}
	ctx := WrapContext(context.Background())

	// Same statement string, same args — but stored under different repo keys.
	users.Set(ctx, "users-row", "SELECT count(*)", 0)
	posts.Set(ctx, "posts-row", "SELECT count(*)", 0)

	got, err := users.Get(ctx, "SELECT count(*)", 0)
	assert.NoError(t, err)
	assert.Equal(t, "users-row", got, "users cache must see its own value")

	got, err = posts.Get(ctx, "SELECT count(*)", 0)
	assert.NoError(t, err)
	assert.Equal(t, "posts-row", got, "posts cache must see its own value")
}
