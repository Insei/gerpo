package databasesql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/insei/gerpo/executor/adapters/placeholder"
	"github.com/insei/gerpo/executor/types"
)

type txWrap struct {
	commited                      bool
	rollbackUnlessCommittedNeeded bool
	tx                            *sql.Tx
	placeholder                   placeholder.PlaceholderFormat
}

func (t *txWrap) ExecContext(ctx context.Context, query string, args ...any) (types.Result, error) {
	sql, err := t.placeholder.ReplacePlaceholders(query)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	return t.tx.ExecContext(ctx, sql, args...)
}

func (t *txWrap) QueryContext(ctx context.Context, query string, args ...any) (types.Rows, error) {
	sql, err := t.placeholder.ReplacePlaceholders(query)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	return t.tx.QueryContext(ctx, sql, args...)
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

func (t *txWrap) RollbackUnlessCommitted() error {
	if !t.commited && t.rollbackUnlessCommittedNeeded {
		err := t.Rollback()
		if err != nil {
			return err
		}
	}
	return nil
}
