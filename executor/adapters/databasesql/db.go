package databasesql

import (
	"context"
	"database/sql"

	"github.com/insei/gerpo/executor/types"
)

type dbWrap struct {
	db *sql.DB
}

func (w *dbWrap) BeginTx(ctx context.Context) (types.Tx, error) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &txWrap{
		tx:                            tx,
		rollbackUnlessCommittedNeeded: true,
	}, nil
}

func (w *dbWrap) ExecContext(ctx context.Context, query string, args ...any) (types.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

func (w *dbWrap) QueryContext(ctx context.Context, query string, args ...any) (types.Rows, error) {
	return w.db.QueryContext(ctx, query, args...)
}

func NewAdapter(db *sql.DB) types.DBAdapter {
	return &dbWrap{db}
}
