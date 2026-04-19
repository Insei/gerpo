// Package types holds the low-level interfaces that gerpo expects from a database driver:
// Result, Rows, Tx, ExecQuery and DBAdapter. The "types" name is kept for backwards
// compatibility with the public API.
package types //nolint:revive // public API package name kept for backwards compatibility

import "context"

type Result interface {
	RowsAffected() (int64, error)
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
}

type Tx interface {
	Rollback() error
	Commit() error
	RollbackUnlessCommitted() error
	ExecQuery
}

type DBAdapter interface {
	ExecQuery
	BeginTx(ctx context.Context) (Tx, error)
}

type ExecQuery interface {
	ExecContext(ctx context.Context, query string, args ...any) (Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
}
