package query

import (
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type InsertBuilderFactory[TModel any] struct {
	model   *TModel
	columns *types.ColumnsStorage
}

func NewInsertBuilderFactory[TModel any](model *TModel, columns *types.ColumnsStorage) *InsertBuilderFactory[TModel] {
	return &InsertBuilderFactory[TModel]{
		model:   model,
		columns: columns,
	}
}

func (f *InsertBuilderFactory[TModel]) New() *InsertBuilder[TModel] {
	return &InsertBuilder[TModel]{
		fabric: f,
		opts:   nil,
	}
}

type InsertBuilder[TModel any] struct {
	fabric *InsertBuilderFactory[TModel]
	opts   []func(b *sql.StringInsertBuilder)
}

func (q *InsertBuilder[TModel]) Exclude(fieldPtr any) *InsertBuilder[TModel] {
	col, err := q.fabric.columns.GetByFieldPtr(q.fabric.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	q.opts = append(q.opts, func(b *sql.StringInsertBuilder) {
		b.Exclude(col)
	})
	return q
}

func (q *InsertBuilder[TModel]) Apply(strSQLBuilder *sql.StringInsertBuilder) {
	strSQLBuilder.Insert(q.fabric.columns.AsSlice()...)
	for _, opt := range q.opts {
		opt(strSQLBuilder)
	}
}
