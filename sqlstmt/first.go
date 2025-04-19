package sqlstmt

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type GetFirst struct {
	ctx context.Context

	table   string
	columns types.ExecutionColumns

	*sqlselect
}

func NewGetFirst(ctx context.Context, table string, columnsStorage types.ColumnsStorage) *GetFirst {
	f := &GetFirst{
		ctx:       ctx,
		table:     table,
		sqlselect: newSelect(ctx, columnsStorage),
		columns:   columnsStorage.NewExecutionColumns(ctx, types.SQLActionSelect),
	}
	return f
}

func (f *GetFirst) Columns() types.ExecutionColumns {
	return f.columns
}

func (f *GetFirst) sql() (string, error) {
	if f.table == "" {
		return "", ErrTableIsNoSet
	}
	columns := f.columns.GetAll()
	if len(columns) < 1 {
		return "", ErrEmptyColumnsInExecutionSet
	}
	sql := ""
	for _, col := range columns {
		if sql != "" {
			sql += ", "
		}
		sql += col.ToSQL(f.ctx)
	}
	return fmt.Sprintf("SELECT %s FROM %s", sql, f.table), nil
}

func (f *GetFirst) SQL(_ ...Option) (string, []any, error) {
	sql, err := f.sql()
	if err != nil {
		return "", nil, err
	}
	sql += f.join.SQL()
	sql += f.where.SQL()
	sql += f.order.SQL()
	sql += f.group.SQL()
	limitOffset := sqlpart.NewLimitOffsetBuilder()
	limitOffset.SetLimit(1)
	sql += limitOffset.SQL()
	return sql, f.where.Values(), nil
}
