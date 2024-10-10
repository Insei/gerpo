package query

import (
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type SelectBuilderFactory[TModel any] struct {
	model   *TModel
	columns *types.ColumnsStorage
}

func NewSelectBuilderFactory[TModel any](model *TModel, columns *types.ColumnsStorage) *SelectBuilderFactory[TModel] {
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
	opts   []func(b *sql.StringSelectBuilder)
}

func (q *SelectBuilder[TModel]) Exclude(fieldPtr any) *SelectBuilder[TModel] {
	col, err := q.fabric.columns.GetByFieldPtr(q.fabric.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	q.opts = append(q.opts, func(b *sql.StringSelectBuilder) {
		b.Exclude(col)
	})
	return q
}

func (q *SelectBuilder[TModel]) Page(page uint64) *SelectBuilder[TModel] {
	q.opts = append(q.opts, func(b *sql.StringSelectBuilder) {
		if page > 0 {
			page--
		}
		b.Offset(page)
	})
	return q
}

func (q *SelectBuilder[TModel]) Size(size uint64) *SelectBuilder[TModel] {
	q.opts = append(q.opts, func(b *sql.StringSelectBuilder) {
		if size == 0 {
			size = 10
		}
		b.Limit(size)
	})
	return q
}

func (q *SelectBuilder[TModel]) Apply(strSQLBuilder *sql.StringSelectBuilder) {
	strSQLBuilder.Select(q.fabric.columns.AsSlice()...)
	for _, opt := range q.opts {
		opt(strSQLBuilder)
	}
}
