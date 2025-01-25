package sqlstmt

import (
	"context"
	"fmt"
	"strings"

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

func (f *GetList) sql() string {
	sql := ""
	for _, col := range f.columns.GetAll() {
		if sql != "" {
			sql += ", "
		}
		sql += col.ToSQL(f.ctx)
	}
	if strings.TrimSpace(sql) == "" {
		return ""
	}
	return fmt.Sprintf("SELECT %s FROM %s", sql, f.table)
}

func (f *GetList) SQL(_ ...Option) (string, []any) {
	sql := f.sql()
	sql += f.join.SQL()
	sql += f.where.SQL()
	sql += f.order.SQL()
	sql += f.group.SQL()
	sql += f.limitOffset.SQL()
	return sql, f.where.Values()
}
