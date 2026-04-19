package sqlstmt

import (
	"context"
	"strings"
	"sync"

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

var getListPool = sync.Pool{
	New: func() any {
		return &GetList{
			sqlselect:   newSelectEmpty(),
			limitOffset: sqlpart.NewLimitOffsetBuilder(),
		}
	},
}

func NewGetList(ctx context.Context, table string, colStorage types.ColumnsStorage) *GetList {
	f := getListPool.Get().(*GetList)
	f.ctx = ctx
	f.table = table
	f.columns = colStorage.NewExecutionColumns(ctx, types.SQLActionSelect)
	f.limitOffset.SetLimit(0)
	f.limitOffset.SetOffset(0)
	f.sqlselect.reset(ctx, colStorage)
	return f
}

// Release returns the statement to the pool. Must not be used after Release.
func (f *GetList) Release() {
	f.ctx = nil
	f.table = ""
	f.columns = nil
	f.sqlselect.columnsStorage = nil
	getListPool.Put(f)
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
	sb.Grow(160)
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
	sb.WriteString(f.limitOffset.SQL())
	return sb.String(), mergeArgs(f.join.Values(), f.where.Values()), nil
}
