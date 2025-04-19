package sqlstmt

import (
	"context"
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

func (f *GetList) SQL(_ ...Option) (string, []any, error) {
	if f.table == "" {
		return "", nil, ErrTableIsNoSet
	}
	columns := f.columns.GetAll()
	if len(columns) < 1 {
		return "", nil, ErrEmptyColumnsInExecutionSet
	}
	sb := strings.Builder{}
	sb.WriteString("SELECT ")
	for _, col := range columns {
		if sb.Len() > 8 {
			sb.WriteString(", ")
		}
		sb.WriteString(col.ToSQL(f.ctx))
	}
	sb.WriteString(" FROM " + f.table)
	sb.WriteString(f.join.SQL())
	sb.WriteString(f.where.SQL())
	sb.WriteString(f.order.SQL())
	sb.WriteString(f.group.SQL())
	sb.WriteString(f.limitOffset.SQL())
	return sb.String(), f.where.Values(), nil
}
