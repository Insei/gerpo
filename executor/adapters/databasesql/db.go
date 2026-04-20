package databasesql

import (
	"context"
	"database/sql"

	"github.com/insei/gerpo/executor/adapters/internal"
	"github.com/insei/gerpo/executor/adapters/placeholder"
	extypes "github.com/insei/gerpo/executor/types"
)

// dbBackend implements internal.Driver on top of a standard *sql.DB.
// *sql.Result and *sql.Rows already satisfy executor/types.Result and
// executor/types.Rows respectively, so no extra wrapper types are needed.
type dbBackend struct {
	db *sql.DB
}

func (b *dbBackend) Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error) {
	return b.db.ExecContext(ctx, sql, args...)
}

func (b *dbBackend) Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error) {
	return b.db.QueryContext(ctx, sql, args...)
}

func (b *dbBackend) BeginTx(ctx context.Context) (internal.TxDriver, error) {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &txBackend{tx: tx}, nil
}

// txBackend implements internal.TxDriver on top of *sql.Tx.
type txBackend struct {
	tx *sql.Tx
}

func (t *txBackend) Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error) {
	return t.tx.ExecContext(ctx, sql, args...)
}

func (t *txBackend) Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error) {
	return t.tx.QueryContext(ctx, sql, args...)
}

func (t *txBackend) Commit() error   { return t.tx.Commit() }
func (t *txBackend) Rollback() error { return t.tx.Rollback() }

// adapterConfig collects the optional knobs for NewAdapter.
type adapterConfig struct {
	placeholder placeholder.PlaceholderFormat
}

// NewAdapter wraps a database/sql DB with the gerpo DB adapter contract.
// The placeholder format defaults to `?` (MySQL); use WithPlaceholder to
// switch to `$1, $2, …` for PostgreSQL.
func NewAdapter(db *sql.DB, opts ...Option) extypes.Adapter {
	cfg := adapterConfig{placeholder: placeholder.Question}
	for _, opt := range opts {
		opt.apply(&cfg)
	}
	return internal.New(&dbBackend{db: db}, cfg.placeholder)
}
