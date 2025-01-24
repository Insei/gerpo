package sqlstmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type GetFirst struct {
	ctx context.Context

	table   string
	columns types.ExecutionColumns

	*sqlselect
}

func NewGetFirst(ctx context.Context, table string, columnsStorage *types.ColumnsStorage) *GetFirst {
	f := &GetFirst{
		table:     table,
		sqlselect: newSelect(ctx, columnsStorage),
	}
	return f
}

func (f *GetFirst) Columns() types.ExecutionColumns {
	return f.columns
}

func (f *GetFirst) sql() string {
	sql := ""
	for _, col := range f.columns.GetAll() {
		if sql != "" {
			sql += ", "
		}
		sql += col.ToSQL(f.ctx)
	}
	if strings.TrimSpace(sql) == "" {
		panic("empty sql")
	}
	return fmt.Sprintf("SELECT %s FROM %s", sql, f.table)
}

func (f *GetFirst) SQL(_ ...Option) (string, []any) {
	sql := f.sql()
	sql += f.join.SQL()
	sql += f.where.SQL()
	sql += f.order.SQL()
	sql += f.group.SQL()
	limitOffset := sqlpart.NewLimitOffsetBuilder()
	limitOffset.SetLimit(1)
	limitOffset.SetOffset(0)
	sql += limitOffset.SQL()
	return sql, f.where.Values()
}
