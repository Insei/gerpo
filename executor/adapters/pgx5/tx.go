package pgx4

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/executor/adapters/placeholder"
	extypes "github.com/insei/gerpo/executor/types"
	"github.com/jackc/pgx/v5"
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

func (t txWrap) RollbackUnlessCommitted() error {
	if !t.commited && t.rollbackUnlessCommittedNeeded {
		err := t.Rollback()
		if err != nil {
			return err
		}
	}
	return nil
}

func (t txWrap) ExecContext(ctx context.Context, query string, args ...any) (extypes.Result, error) {
	sql, err := placeholder.Dollar.ReplacePlaceholders(query)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	res, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrap{res: res}, nil
}

func (t txWrap) QueryContext(ctx context.Context, query string, args ...any) (extypes.Rows, error) {
	sql, err := placeholder.Dollar.ReplacePlaceholders(query)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrap{rows: rows}, nil
}
