package sqlstmt

import (
	"context"
	"strings"
	"sync"

	"github.com/insei/gerpo/types"
)

type GetFirst struct {
	ctx context.Context

	table   string
	columns types.ExecutionColumns

	*sqlselect
}

var getFirstPool = sync.Pool{
	New: func() any {
		return &GetFirst{sqlselect: newSelectEmpty()}
	},
}

func NewGetFirst(ctx context.Context, table string, columnsStorage types.ColumnsStorage) *GetFirst {
	f := getFirstPool.Get().(*GetFirst)
	f.ctx = ctx
	f.table = table
	f.columns = columnsStorage.NewExecutionColumns(ctx, types.SQLActionSelect)
	f.sqlselect.reset(ctx, columnsStorage)
	return f
}

// Release returns the statement to the pool. Must not be used after Release.
func (f *GetFirst) Release() {
	f.ctx = nil
	f.table = ""
	f.columns = nil
	f.sqlselect.columnsStorage = nil
	getFirstPool.Put(f)
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
	sb.Grow(128)
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
	sb.WriteString(f.group.SQL())
	sb.WriteString(f.order.SQL())
	sb.WriteString(" LIMIT 1")
	return sb.String(), mergeArgs(f.join.Values(), f.where.Values()), nil
}
