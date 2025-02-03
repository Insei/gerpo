package sqlstmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type Update struct {
	ctx context.Context

	table string
	vals  *values

	colsStorage types.ColumnsStorage
	columns     types.ExecutionColumns
	where       *sqlpart.WhereBuilder
}

func NewUpdate(ctx context.Context, colStorage types.ColumnsStorage, table string) *Update {
	columns := colStorage.NewExecutionColumns(ctx, types.SQLActionUpdate)
	return &Update{
		ctx: ctx,

		colsStorage: colStorage,
		table:       table,
		vals:        newValues(columns),
		columns:     columns,

		where: sqlpart.NewWhereBuilder(ctx),
	}
}

func (u *Update) sql() string {
	cols := u.columns.GetAll()
	colsStr := ""
	if len(cols) < 1 {
		return colsStr
	}
	for _, col := range cols {
		colName, ok := col.Name()
		if !ok {
			continue
		}
		colsStr += colName + " = ?, "
	}
	if strings.TrimSpace(colsStr) == "" {
		return ""
	}
	return fmt.Sprintf("UPDATE %s SET %s", u.table, colsStr[:len(colsStr)-2])
}

func (u *Update) ColumnsStorage() types.ColumnsStorage {
	return u.colsStorage
}

func (u *Update) Columns() types.ExecutionColumns {
	return u.columns
}

func (u *Update) Where() sqlpart.Where {
	return u.where
}

func (u *Update) SQL(opts ...Option) (string, []any) {
	sql := u.sql()
	sql += u.where.SQL()
	for _, opt := range opts {
		opt(u.vals)
	}
	vals := append(u.vals.values, u.where.Values()...)
	return sql, vals
}
