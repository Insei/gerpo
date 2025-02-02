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
			s := &CtxCache{log: logger.NoopLogger}
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
			ctx:         NewCtxCache(context.Background()),
			expectedErr: types.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CtxCache{log: logger.NoopLogger}
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
			ctx:           NewCtxCache(context.Background()),
			modelKey:      "testKey",
			cache:         "fakeCache",
			statement:     "setCache",
			statementArgs: []any{"arg1", "arg2"},
		},
		// add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CtxCache{log: logger.NoopLogger, key: tt.modelKey}
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
			ctx:  NewCtxCache(context.Background()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CtxCache{log: logger.NoopLogger, key: "testKey"}
			s.Clean(tt.ctx)
			// Since method doesn't return anything, no assertions made.
		})
	}
}
