package pgx5

import (
	"github.com/jackc/pgx/v5"
)

type rowsWrap struct {
	rows pgx.Rows
}

func (r *rowsWrap) Next() bool {
	return r.rows.Next()
}

func (r *rowsWrap) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *rowsWrap) Close() error {
	r.rows.Close()
	return nil
}
