package executor

import (
	"context"
	"testing"

	"github.com/insei/gerpo/executor/cache"
	"github.com/stretchr/testify/mock"
)

type MockCacheSource struct {
	mock.Mock
}

func (m *MockCacheSource) Clean(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockCacheSource) Get(ctx context.Context, statement string, statementArgs ...any) (any, error) {
	args := m.Called(ctx, statement, statementArgs)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheSource) Set(ctx context.Context, cache any, statement string, statementArgs ...any) {
	m.Called(ctx, cache, statement, statementArgs)
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	b := new(MockCacheSource)
	tests := []struct {
		name        string
		stmt        string
		stmtArgs    []any
		mockReturns any
		mockErr     error
		want        any
		wantOk      bool
	}{
		{
			name:        "Found OK and Cast Ok",
			stmt:        "SELECT * FROM users WHERE id = ?",
			mockReturns: "any",
			want:        "any",
			wantOk:      true,
		},
		{
			name:        "Found OK and Cast Fail",
			stmt:        "SELECT * FROM users WHERE id = ?",
			mockReturns: 123,
			want:        "any",
			wantOk:      false,
		},
		{
			name:        "Found Fail",
			stmt:        "SELECT * FROM users WHERE id = ?",
			mockReturns: nil,
			mockErr:     cache.ErrNotFound,
			want:        nil,
			wantOk:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockReturns, tt.mockErr).Once()
			result, ok := get[string](ctx, b, tt.stmt, tt.stmtArgs)
			if ok != tt.wantOk {
				t.Errorf("get() = %v, %v, want %v, %v", result, ok, tt.want, tt.wantOk)
				return
			}
			if ok && *result != tt.want {
				t.Errorf("get() = %v, want %v", *result, tt.want)
			}
			b.AssertExpectations(t)
		})
	}
}

func TestSet(t *testing.T) {
	ctx := context.Background()
	b := new(MockCacheSource)
	tests := []struct {
		name     string
		stmt     string
		cache    any
		stmtArgs []any
	}{
		{
			name:  "Value set OK",
			cache: "any",
			stmt:  "INSERT INTO users WHERE id = ?",
		},
		{
			name:  "Value set Empty Statement",
			cache: "any",
			stmt:  "",
		},
		{
			name:  "Value set Nil cache",
			cache: nil,
			stmt:  "INSERT INTO users WHERE id = ?",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Once()
			set(ctx, b, tt.cache, tt.stmt, tt.stmtArgs...)
			b.AssertExpectations(t)
		})
	}
}

func TestClean(t *testing.T) {
	ctx := context.Background()
	b := new(MockCacheSource)
	tests := []struct {
		name string
	}{
		{
			name: "clean OK",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b.On("Clean", mock.Anything).Return().Once()
			clean(ctx, b)
			b.AssertExpectations(t)
		})
	}
}
