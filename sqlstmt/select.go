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

// newSelectEmpty allocates an sqlselect with empty builders intended for sync.Pool warmup.
// Builders are initialized with context.TODO() because a real ctx is injected via reset()
// before the instance is exposed to the caller.
func newSelectEmpty() *sqlselect {
	ctx := context.TODO()
	return &sqlselect{
		where: sqlpart.NewWhereBuilder(ctx),
		join:  sqlpart.NewJoinBuilder(ctx),
		order: sqlpart.NewOrderBuilder(ctx),
		group: sqlpart.NewGroupBuilder(ctx),
	}
}

// reset prepares an sqlselect for reuse by a new query.
func (f *sqlselect) reset(ctx context.Context, storage types.ColumnsStorage) {
	f.columnsStorage = storage
	f.where.Reset(ctx)
	f.join.Reset(ctx)
	f.order.Reset(ctx)
	f.group.Reset(ctx)
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
