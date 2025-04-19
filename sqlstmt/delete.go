package sqlstmt

import (
	"context"
	"strings"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type Delete struct {
	ctx context.Context

	table          string
	columnsStorage types.ColumnsStorage

	join  *sqlpart.JoinBuilder
	where *sqlpart.WhereBuilder
}

func NewDelete(ctx context.Context, table string, columnsStorage types.ColumnsStorage) *Delete {
	return &Delete{
		ctx:            ctx,
		table:          table,
		columnsStorage: columnsStorage,
		join:           sqlpart.NewJoinBuilder(ctx),
		where:          sqlpart.NewWhereBuilder(ctx),
	}
}

func (d *Delete) Where() sqlpart.Where {
	return d.where
}

func (d *Delete) Join() sqlpart.Join {
	return d.join
}

func (d *Delete) ColumnsStorage() types.ColumnsStorage {
	return d.columnsStorage
}

func (d *Delete) sql() (string, error) {
	if strings.TrimSpace(d.table) == "" {
		return "", ErrTableIsNoSet
	}
	return "DELETE FROM " + d.table, nil
}

func (d *Delete) SQL(_ ...Option) (string, []any, error) {
	sql, err := d.sql()
	if err != nil {
		return "", nil, err
	}
	sql += d.join.SQL()
	sql += d.where.SQL()
	return sql, d.where.Values(), nil
}
