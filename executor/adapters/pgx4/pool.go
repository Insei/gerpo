package pgx4

import (
	"context"

	"github.com/insei/gerpo/executor/types"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type poolWrap struct {
	pool *pgxpool.Pool
}

func (p *poolWrap) ExecContext(ctx context.Context, query string, args ...any) (types.Result, error) {
	res, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrap{res: res}, nil
}

func (p *poolWrap) QueryContext(ctx context.Context, query string, args ...any) (types.Rows, error) {
	rows, err := p.pool.Query(ctx, query, args...)
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
