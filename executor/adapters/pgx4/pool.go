package pgx4

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/executor/adapters/placeholder"
	"github.com/insei/gerpo/executor/types"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type poolWrap struct {
	pool *pgxpool.Pool
}

func (p *poolWrap) ExecContext(ctx context.Context, query string, args ...any) (types.Result, error) {
	sql, err := placeholder.Dollar.ReplacePlaceholders(query)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	res, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrap{res: res}, nil
}

func (p *poolWrap) QueryContext(ctx context.Context, query string, args ...any) (types.Rows, error) {
	sql, err := placeholder.Dollar.ReplacePlaceholders(query)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrap{rows: rows}, nil
}

func (p *poolWrap) BeginTx(ctx context.Context) (types.Tx, error) {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &txWrap{
		rollbackUnlessCommittedNeeded: true,
		tx:                            tx,
	}, err
}

func NewPoolAdapter(pool *pgxpool.Pool) types.DBAdapter {
	return &poolWrap{pool}
}
