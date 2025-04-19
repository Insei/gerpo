package types

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
