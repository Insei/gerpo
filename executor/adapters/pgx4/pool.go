package pgx4

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/insei/gerpo/executor/adapters/internal"
	"github.com/insei/gerpo/executor/adapters/placeholder"
	extypes "github.com/insei/gerpo/executor/types"
)

// poolDriver implements internal.Driver on top of a pgx v4 connection pool.
type poolDriver struct {
	pool *pgxpool.Pool
}

func (b *poolDriver) Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error) {
	res, err := b.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrap{res: res}, nil
}

func (b *poolDriver) Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error) {
	rows, err := b.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrap{rows: rows}, nil
}

func (b *poolDriver) BeginTx(ctx context.Context) (internal.TxDriver, error) {
	tx, err := b.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &txDriver{tx: tx}, nil
}

// txDriver implements internal.TxDriver on top of pgx.Tx.
type txDriver struct {
	tx pgx.Tx
}

func (t *txDriver) Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error) {
	res, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrap{res: res}, nil
}

func (t *txDriver) Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error) {
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrap{rows: rows}, nil
}

func (t *txDriver) Commit() error   { return t.tx.Commit(context.Background()) }
func (t *txDriver) Rollback() error { return t.tx.Rollback(context.Background()) }

// NewPoolAdapter wraps a pgx v4 pool with the gerpo DB adapter contract.
// SQL placeholders are rewritten from `?` to PostgreSQL's `$1, $2, …` form.
func NewPoolAdapter(pool *pgxpool.Pool) extypes.Adapter {
	return internal.New(&poolDriver{pool: pool}, placeholder.Dollar)
}
