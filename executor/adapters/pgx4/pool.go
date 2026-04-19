package pgx4

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/insei/gerpo/executor/adapters/internal"
	"github.com/insei/gerpo/executor/adapters/placeholder"
	extypes "github.com/insei/gerpo/executor/types"
)

// poolBackend implements internal.Backend on top of a pgx v4 connection pool.
type poolBackend struct {
	pool *pgxpool.Pool
}

func (b *poolBackend) Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error) {
	res, err := b.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrap{res: res}, nil
}

func (b *poolBackend) Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error) {
	rows, err := b.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrap{rows: rows}, nil
}

func (b *poolBackend) BeginTx(ctx context.Context) (internal.TxBackend, error) {
	tx, err := b.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &txBackend{tx: tx}, nil
}

// txBackend implements internal.TxBackend on top of pgx.Tx.
type txBackend struct {
	tx pgx.Tx
}

func (t *txBackend) Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error) {
	res, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrap{res: res}, nil
}

func (t *txBackend) Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error) {
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrap{rows: rows}, nil
}

func (t *txBackend) Commit() error   { return t.tx.Commit(context.Background()) }
func (t *txBackend) Rollback() error { return t.tx.Rollback(context.Background()) }

// NewPoolAdapter wraps a pgx v4 pool with the gerpo DB adapter contract.
// SQL placeholders are rewritten from `?` to PostgreSQL's `$1, $2, …` form.
func NewPoolAdapter(pool *pgxpool.Pool) extypes.DBAdapter {
	return internal.New(&poolBackend{pool: pool}, placeholder.Dollar)
}
