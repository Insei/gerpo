package databasesql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/insei/gerpo/executor/adapters/placeholder"
	"github.com/insei/gerpo/executor/types"
)

type dbWrap struct {
	db          *sql.DB
	placeholder placeholder.PlaceholderFormat
}

func (w *dbWrap) BeginTx(ctx context.Context) (types.Tx, error) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &txWrap{
		tx:                            tx,
		rollbackUnlessCommittedNeeded: true,
		placeholder:                   w.placeholder,
	}, nil
}

func (w *dbWrap) ExecContext(ctx context.Context, query string, args ...any) (types.Result, error) {
	sql, err := w.placeholder.ReplacePlaceholders(query)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	return w.db.ExecContext(ctx, sql, args...)
}

func (w *dbWrap) QueryContext(ctx context.Context, query string, args ...any) (types.Rows, error) {
	sql, err := w.placeholder.ReplacePlaceholders(query)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	return w.db.QueryContext(ctx, sql, args...)
}

func NewAdapter(db *sql.DB, opts ...Option) types.DBAdapter {
	wrappedDb := &dbWrap{
		db:          db,
		placeholder: placeholder.Question,
	}
	for _, opt := range opts {
		opt.apply(wrappedDb)
	}
	return wrappedDb
}
