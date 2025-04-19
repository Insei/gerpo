package sqlstmt

import (
	"context"
	"strings"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type Delete struct {
	ctx context.Context

	table          string
	columnsStorage types.ColumnsStorage

	join  *sqlpart.JoinBuilder
	where *sqlpart.WhereBuilder
}

func NewDelete(ctx context.Context, table string, columnsStorage types.ColumnsStorage) *Delete {
	return &Delete{
		ctx:            ctx,
		table:          table,
		columnsStorage: columnsStorage,
		join:           sqlpart.NewJoinBuilder(ctx),
		where:          sqlpart.NewWhereBuilder(ctx),
	}
}

func (d *Delete) Where() sqlpart.Where {
	return d.where
}

func (d *Delete) Join() sqlpart.Join {
	return d.join
}

func (d *Delete) ColumnsStorage() types.ColumnsStorage {
	return d.columnsStorage
}

func (d *Delete) SQL(_ ...Option) (string, []any, error) {
	if strings.TrimSpace(d.table) == "" {
		return "", nil, ErrTableIsNoSet
	}
	sb := strings.Builder{}
	sb.WriteString("DELETE FROM " + d.table)
	sb.WriteString(d.join.SQL())
	sb.WriteString(d.where.SQL())
	return sb.String(), d.where.Values(), nil
}
