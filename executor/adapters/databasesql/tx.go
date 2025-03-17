package databasesql

import (
	"context"
	"database/sql"

	"github.com/insei/gerpo/executor/types"
)

type txWrap struct {
	commited                      bool
	rollbackUnlessCommittedNeeded bool
	tx                            *sql.Tx
}

func (t *txWrap) ExecContext(ctx context.Context, query string, args ...any) (types.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *txWrap) QueryContext(ctx context.Context, query string, args ...any) (types.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *txWrap) Commit() error {
	err := t.tx.Commit()
	if err != nil {
		return err
	}
	t.commited = true
	return nil
}

func (t *txWrap) Rollback() error {
	t.rollbackUnlessCommittedNeeded = false
	return t.tx.Rollback()
}

func (t *txWrap) RollbackUnlessCommitted() {
	if !t.commited && t.rollbackUnlessCommittedNeeded {
		err := t.Rollback()
		if err != nil {
			panic(err)
		}
	}
}
