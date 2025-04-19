package pgx4

import "github.com/jackc/pgx/v5/pgconn"

type resultWrap struct {
	res pgconn.CommandTag
}

func (e *resultWrap) RowsAffected() (int64, error) {
	return e.res.RowsAffected(), nil
}
