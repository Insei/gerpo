package query

import (
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type OrderBuilderFactory[TModel any] struct {
	model   *TModel
	columns *types.ColumnsStorage
}

func NewOrderBuilderFabric[TModel any](model *TModel, columns *types.ColumnsStorage) *OrderBuilderFactory[TModel] {
	return &OrderBuilderFactory[TModel]{
		model:   model,
		columns: columns,
	}
}

func (f *OrderBuilderFactory[TModel]) New() *OrderBuilder[TModel] {
	return &OrderBuilder[TModel]{
		fabric: f,
		opts:   nil,
	}
}

type OrderBuilder[TModel any] struct {
	fabric *OrderBuilderFactory[TModel]
	opts   []func(b *sql.StringOrderBuilder)
}

func (q *OrderBuilder[TModel]) Apply(strSQLBuilder *sql.StringOrderBuilder) {
	for _, opt := range q.opts {
		opt(strSQLBuilder)
	}
}

type OrderDirectionFn[TModel any] func(operation types.OrderDirection) *OrderBuilder[TModel]

func (f OrderDirectionFn[TModel]) ASC() types.OrderTarget[TModel] {
	return f(types.OrderDirectionASC)
}

func (f OrderDirectionFn[TModel]) DESC() types.OrderTarget[TModel] {
	return f(types.OrderDirectionDESC)
}

func (q *OrderBuilder[TModel]) Field(fieldPtr any) types.OrderOperation[TModel] {
	col, err := q.fabric.columns.GetByFieldPtr(q.fabric.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	return OrderDirectionFn[TModel](func(direction types.OrderDirection) *OrderBuilder[TModel] {
		q.opts = append(q.opts, func(b *sql.StringOrderBuilder) {
			err := b.OrderByColumn(col, direction)
			if err != nil {
				panic(err)
			}
		})
		return q
	})
}
