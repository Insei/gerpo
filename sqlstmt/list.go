package sqlstmt

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type GetList struct {
	ctx context.Context

	*sqlselect
	table       string
	columns     types.ExecutionColumns
	limitOffset *sqlpart.LimitOffsetBuilder
}

func NewGetList(ctx context.Context, table string, colStorage types.ColumnsStorage) *GetList {
	executionColumns := colStorage.NewExecutionColumns(ctx, types.SQLActionSelect)
	f := &GetList{
		sqlselect:   newSelect(ctx, colStorage),
		limitOffset: sqlpart.NewLimitOffsetBuilder(),
		columns:     executionColumns,
		table:       table,
	}
	return f
}

func (f *GetList) Columns() types.ExecutionColumns {
	return f.columns
}

func (f *GetList) LimitOffset() sqlpart.LimitOffset {
	return f.limitOffset
}

func (f *GetList) sql() (string, error) {
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

func (f *GetList) SQL(_ ...Option) (string, []any, error) {
	sql, err := f.sql()
	if err != nil {
		return "", nil, err
	}
	sql += f.join.SQL()
	sql += f.where.SQL()
	sql += f.order.SQL()
	sql += f.group.SQL()
	sql += f.limitOffset.SQL()
	return sql, f.where.Values(), nil
}
