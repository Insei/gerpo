package sqlstmt

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type Count struct {
	*sqlselect

	table string
}

func NewCount(ctx context.Context, table string, storage *types.ColumnsStorage) *Count {
	f := &Count{
		sqlselect: newSelect(ctx, storage),
		table:     table,
	}
	return f
}

func (c *Count) SQL(_ ...Option) (string, []any) {
	sql := fmt.Sprintf("SELECT %s FROM %s", "count(*) over() AS count", c.table)
	sql += c.join.SQL()
	sql += c.where.SQL()
	sql += c.order.SQL()
	sql += c.group.SQL()
	limitOffset := sqlpart.NewLimitOffsetBuilder()
	limitOffset.SetLimit(1)
	limitOffset.SetOffset(0)
	sql += limitOffset.SQL()
	return sql, c.where.Values()
}
