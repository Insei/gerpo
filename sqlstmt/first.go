package sqlstmt

import (
	"context"
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

func (f *GetFirst) SQL(_ ...Option) (string, []any, error) {
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
	sb.WriteString(" FROM ")
	sb.WriteString(f.table)
	sb.WriteString(f.join.SQL())
	sb.WriteString(f.where.SQL())
	sb.WriteString(f.order.SQL())
	sb.WriteString(f.group.SQL())
	limitOffset := sqlpart.NewLimitOffsetBuilder()
	limitOffset.SetLimit(1)
	sb.WriteString(limitOffset.SQL())
	return sb.String(), f.where.Values(), nil
}
