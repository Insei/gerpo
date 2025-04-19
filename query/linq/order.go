package linq

import (
	"fmt"

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
	opts  []func(applier OrderApplier) error
}

type OrderApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Order() sqlpart.Order
}

func (q *OrderBuilder) Apply(applier OrderApplier) error {
	for _, opt := range q.opts {
		err := opt(applier)
		if err != nil {
			return err
		}
	}
	return nil
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
		q.opts = append(q.opts, func(applier OrderApplier) error {
			if column == nil {
				return fmt.Errorf("column is nil")
			}
			applier.Order().OrderByColumn(column, direction)
			return nil
		})
		return q
	})
}

func (q *OrderBuilder) Field(fieldPtr any) types.OrderOperation {
	return OrderDirectionFn(func(direction types.OrderDirection) *OrderBuilder {
		q.opts = append(q.opts, func(applier OrderApplier) error {
			column, err := applier.ColumnsStorage().GetByFieldPtr(q.model, fieldPtr)
			if err != nil {
				return err
			}
			applier.Order().OrderByColumn(column, direction)
			return nil
		})
		return q
	})
}
