package executor

import (
	"context"
	"database/sql"
)

type Tx struct {
	tx *sql.Tx
	db *sql.DB
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

func BeginTx(ctx context.Context, db *sql.DB, opts ...*sql.TxOptions) (*Tx, error) {
	var opt *sql.TxOptions
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}
	tx, err := db.BeginTx(ctx, opt)
	if err != nil {
		return nil, err
	}
	return &Tx{
		tx: tx,
		db: db,
	}, nil
}
