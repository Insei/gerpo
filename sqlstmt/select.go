package sqlstmt

import (
	"context"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type sqlselect struct {
	columnsStorage types.ColumnsStorage

	where *sqlpart.WhereBuilder
	join  *sqlpart.JoinBuilder
	order *sqlpart.OrderBuilder
	group *sqlpart.GroupBuilder
}

func newSelect(ctx context.Context, storage types.ColumnsStorage) *sqlselect {
	f := &sqlselect{
		columnsStorage: storage,
		where:          sqlpart.NewWhereBuilder(ctx),
		join:           sqlpart.NewJoinBuilder(ctx),
		order:          sqlpart.NewOrderBuilder(ctx),
		group:          sqlpart.NewGroupBuilder(ctx),
	}
	return f
}

func (f *sqlselect) ColumnsStorage() types.ColumnsStorage {
	return f.columnsStorage
}

func (f *sqlselect) Where() sqlpart.Where {
	return f.where
}

func (f *sqlselect) Join() sqlpart.Join {
	return f.join
}

func (f *sqlselect) Order() sqlpart.Order {
	return f.order
}

func (f *sqlselect) Group() sqlpart.Group {
	return f.group
}
