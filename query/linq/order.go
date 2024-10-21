package linq

import (
	"github.com/insei/gerpo/types"
)

type Order interface {
	OrderByColumn(column types.Column, direction types.OrderDirection) error
}

func NewOrderBuilder(core *CoreBuilder) *OrderBuilder {
	return &OrderBuilder{
		core: core,
	}
}

type OrderBuilder struct {
	core *CoreBuilder
	opts []func(o Order)
}

func (q *OrderBuilder) Apply(order Order) {
	for _, opt := range q.opts {
		opt(order)
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
		q.opts = append(q.opts, func(o Order) {
			err := o.OrderByColumn(column, direction)
			if err != nil {
				panic(err)
			}
		})
		return q
	})
}

func (q *OrderBuilder) Field(fieldPtr any) types.OrderOperation {
	col := q.core.GetColumn(fieldPtr)
	return q.Column(col)
}
