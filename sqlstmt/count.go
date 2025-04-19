package sqlstmt

import (
	"context"
	"strings"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type Count struct {
	*sqlselect

	table string
}

func NewCount(ctx context.Context, table string, storage types.ColumnsStorage) *Count {
	f := &Count{
		sqlselect: newSelect(ctx, storage),
		table:     table,
	}
	return f
}

func (c *Count) SQL(_ ...Option) (string, []any, error) {
	if strings.TrimSpace(c.table) == "" {
		return "", nil, ErrTableIsNoSet
	}
	sb := strings.Builder{}
	sb.WriteString("SELECT count(*) over() AS count FROM " + c.table)
	sb.WriteString(c.join.SQL())
	sb.WriteString(c.where.SQL())
	sb.WriteString(c.group.SQL())
	limitOffset := sqlpart.NewLimitOffsetBuilder()
	limitOffset.SetLimit(1)
	sb.WriteString(limitOffset.SQL())
	return sb.String(), c.where.Values(), nil
}
