package query

import "github.com/insei/gerpo/types"

type SelectBuilderFactory[TModel any] struct {
	model   *TModel
	columns *types.ColumnsStorage
}

func NewSelectBuilderFabric[TModel any](model *TModel, columns *types.ColumnsStorage) *SelectBuilderFactory[TModel] {
	return &SelectBuilderFactory[TModel]{
		model:   model,
		columns: columns,
	}
}

func (f *SelectBuilderFactory[TModel]) New() *SelectBuilder[TModel] {
	return &SelectBuilder[TModel]{
		fabric: f,
		opts:   nil,
	}
}

type SelectBuilder[TModel any] struct {
	fabric *SelectBuilderFactory[TModel]
	opts   []func(b *StringSQLSelectBuilder)
}

func (q *SelectBuilder[TModel]) Exclude(fieldPtr any) *SelectBuilder[TModel] {
	col, err := q.fabric.columns.GetByFieldPtr(q.fabric.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	q.opts = append(q.opts, func(b *StringSQLSelectBuilder) {
		b.Exclude(col)
	})
	return q
}

func (q *SelectBuilder[TModel]) Apply(strSQLBuilder *StringSQLSelectBuilder) {
	for _, opt := range q.opts {
		opt(strSQLBuilder)
	}
	strSQLBuilder.Select(q.fabric.columns.AsSlice()...)
}
