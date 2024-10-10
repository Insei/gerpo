package query

import (
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type UpdateBuilderFactory[TModel any] struct {
	model   *TModel
	columns *types.ColumnsStorage
}

func NewUpdateBuilderFactory[TModel any](model *TModel, columns *types.ColumnsStorage) *UpdateBuilderFactory[TModel] {
	return &UpdateBuilderFactory[TModel]{
		model:   model,
		columns: columns,
	}
}

func (f *UpdateBuilderFactory[TModel]) New() *UpdateBuilder[TModel] {
	return &UpdateBuilder[TModel]{
		fabric: f,
		opts:   nil,
	}
}

type UpdateBuilder[TModel any] struct {
	fabric *UpdateBuilderFactory[TModel]
	opts   []func(b *sql.StringUpdateBuilder)
}

func (q *UpdateBuilder[TModel]) Exclude(fieldPtr any) *UpdateBuilder[TModel] {
	col, err := q.fabric.columns.GetByFieldPtr(q.fabric.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	q.opts = append(q.opts, func(b *sql.StringUpdateBuilder) {
		b.Exclude(col)
	})
	return q
}

func (q *UpdateBuilder[TModel]) Apply(strSQLBuilder *sql.StringUpdateBuilder) {
	strSQLBuilder.Update(q.fabric.columns.AsSlice()...)
	for _, opt := range q.opts {
		opt(strSQLBuilder)
	}
}
