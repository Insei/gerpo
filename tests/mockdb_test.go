package tests

import (
	"context"

	"github.com/insei/gerpo/executor"
	extypes "github.com/insei/gerpo/executor/types"
)

type mockDB struct {
	executor.DBAdapter
	ExecContextFn  func(ctx context.Context, query string, args ...any) (extypes.Result, error)
	QueryContextFn func(ctx context.Context, query string, args ...any) (extypes.Rows, error)
	BeginTxFn      func(ctx context.Context) (extypes.Tx, error)
}

func (m *mockDB) ExecContext(ctx context.Context, query string, args ...any) (extypes.Result, error) {
	if m.ExecContextFn != nil {
		return m.ExecContextFn(ctx, query, args...)
	}
	panic("implement me")
}

func (m *mockDB) QueryContext(ctx context.Context, query string, args ...any) (extypes.Rows, error) {
	if m.QueryContextFn != nil {
		return m.QueryContextFn(ctx, query, args...)
	}
	panic("implement me")
}

type mockRows struct {
	alwaysNext   bool
	current, max int
}

func (m *mockRows) Next() bool {
	if m.alwaysNext {
		return true
	}
	if m.current < m.max {
		m.current++
		return true
	}
	return false
}

func (m *mockRows) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockRows) Close() error {
	return nil
}

func (m *mockDB) BeginTx(ctx context.Context) (extypes.Tx, error) {
	if m.BeginTxFn != nil {
		return m.BeginTxFn(ctx)
	}
	panic("implement me")
}

func newMockDB() *mockDB {
	return &mockDB{}
}
