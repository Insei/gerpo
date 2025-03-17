package pgx4

import (
	"context"

	extypes "github.com/insei/gerpo/executor/types"
	"github.com/jackc/pgx/v4"
)

type txWrap struct {
	commited                      bool
	rollbackUnlessCommittedNeeded bool
	tx                            pgx.Tx
}

func (t txWrap) Rollback() error {
	return t.tx.Rollback(context.Background())
}

func (t txWrap) Commit() error {
	return t.tx.Commit(context.Background())
}

func (t txWrap) RollbackUnlessCommitted() {
	if !t.commited && t.rollbackUnlessCommittedNeeded {
		err := t.Rollback()
		if err != nil {
			panic(err)
		}
	}
}

func (t txWrap) ExecContext(ctx context.Context, query string, args ...any) (extypes.Result, error) {
	res, err := t.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrap{res: res}, nil
}

func (t txWrap) QueryContext(ctx context.Context, query string, args ...any) (extypes.Rows, error) {
	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrap{rows: rows}, nil
}
