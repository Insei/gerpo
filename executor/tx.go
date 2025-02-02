package executor

import (
	"context"
	"database/sql"
)

type Tx struct {
	commited                      bool
	rollbackUnlessCommittedNeeded bool
	tx                            *sql.Tx
	db                            *sql.DB
}

func (t *Tx) Commit() error {
	err := t.tx.Commit()
	if err != nil {
		return err
	}
	t.commited = true
	return nil
}

func (t *Tx) Rollback() error {
	t.rollbackUnlessCommittedNeeded = false
	return t.tx.Rollback()
}

func (t *Tx) RollbackUnlessCommitted() {
	if !t.commited && t.rollbackUnlessCommittedNeeded {
		err := t.Rollback()
		if err != nil {
			panic(err)
		}
	}
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
		tx:                            tx,
		db:                            db,
		rollbackUnlessCommittedNeeded: true,
	}, nil
}
