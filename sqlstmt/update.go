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

func (u *Update) sql() (string, error) {
	if u.table == "" {
		return "", ErrTableIsNoSet
	}
	cols := u.columns.GetAll()
	if len(cols) < 1 {
		return "", ErrEmptyColumnsInExecutionSet
	}
	colsStr := ""
	for _, col := range cols {
		colName, ok := col.Name()
		if !ok {
			continue
		}
		colsStr += colName + " = ?, "
	}
	if colsStr == "" {
		return "", fmt.Errorf("columns set is not empty, but no one column is not allowed to set")
	}
	return fmt.Sprintf("UPDATE %s SET %s", u.table, colsStr[:len(colsStr)-2]), nil
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

func (u *Update) SQL(opts ...Option) (string, []any, error) {
	if u.table == "" {
		return "", nil, ErrTableIsNoSet
	}
	cols := u.columns.GetAll()
	if len(cols) < 1 {
		return "", nil, ErrEmptyColumnsInExecutionSet
	}
	sb := strings.Builder{}
	sb.WriteString("UPDATE " + u.table + " SET ")
	lenAtStart := sb.Len()
	for _, col := range cols {
		colName, ok := col.Name()
		if !ok {
			continue
		}
		if sb.Len() > lenAtStart {
			sb.WriteString(", ")
		}
		sb.WriteString(colName + " = ?")
	}
	if sb.Len() == lenAtStart {
		return "", nil, fmt.Errorf("columns set is not empty, but no one column is not allowed to set")
	}
	sb.WriteString(u.where.SQL())
	for _, opt := range opts {
		opt(u.vals)
	}
	vals := append(u.vals.values, u.where.Values()...)
	return sb.String(), vals, nil
}
