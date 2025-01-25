package linq

import (
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

func NewOrderBuilder(baseModel any) *OrderBuilder {
	return &OrderBuilder{
		model: baseModel,
	}
}

type OrderBuilder struct {
	model any
	opts  []func(applier OrderApplier)
}

type OrderApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Order() sqlpart.Order
}

func (q *OrderBuilder) Apply(applier OrderApplier) {
	for _, opt := range q.opts {
		opt(applier)
	}
}

type OrderDirectionFn func(operation types.OrderDirection) *OrderBuilder

func (f OrderDirectionFn) ASC() types.OrderTarget {
	return f(types.OrderDirectionASC)
}

func (f OrderDirectionFn) DESC() types.OrderTarget {
	return f(types.OrderDirectionDESC)
}

func (q *OrderBuilder) Column(column types.Column) types.OrderOperation {
	return OrderDirectionFn(func(direction types.OrderDirection) *OrderBuilder {
		q.opts = append(q.opts, func(applier OrderApplier) {
			applier.Order().OrderByColumn(column, direction)
		})
		return q
	})
}

func (q *OrderBuilder) Field(fieldPtr any) types.OrderOperation {
	return OrderDirectionFn(func(direction types.OrderDirection) *OrderBuilder {
		q.opts = append(q.opts, func(applier OrderApplier) {
			column, err := applier.ColumnsStorage().GetByFieldPtr(q.model, fieldPtr)
			if err != nil {
				panic(err)
			}
			applier.Order().OrderByColumn(column, direction)
		})
		return q
	})
}
