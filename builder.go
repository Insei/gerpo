package gerpo

import (
	"context"
	"errors"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/types"
)

type Builder[TModel any] struct {
	table   string
	opts    []Option[TModel]
	model   *TModel
	fields  fmap.Storage
	columns *types.ColumnsStorage
}

func NewBuilder[TModel any]() (*Builder[TModel], error) {
	model, fields, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	return &Builder[TModel]{
		opts:    nil,
		model:   model,
		fields:  fields,
		columns: nil,
	}, nil
}

func (b *Builder[TModel]) Table(table string) *Builder[TModel] {
	b.table = table
	return b
}

func (b *Builder[TModel]) Columns(fn func(m *TModel, columns *ColumnBuilder[TModel])) *Builder[TModel] {
	columnsBuilder := NewColumnBuilder[TModel](b.table, b.model, b.fields)
	fn(b.model, columnsBuilder)
	b.columns = columnsBuilder.build()
	return b
}

func (b *Builder[TModel]) LeftJoin(leftJoinFn func(ctx context.Context) string) *Builder[TModel] {
	b.opts = append(b.opts, WithLeftJoin[TModel](leftJoinFn))
	return b
}

func (b *Builder[TModel]) BeforeInsert(fn func(ctx context.Context, m *TModel)) *Builder[TModel] {
	b.opts = append(b.opts, WithBeforeInsert[TModel](fn))
	return b
}

func (b *Builder[TModel]) BeforeUpdate(fn func(ctx context.Context, m *TModel)) *Builder[TModel] {
	b.opts = append(b.opts, WithBeforeUpdate[TModel](fn))
	return b
}

func (b *Builder[TModel]) SoftDelete(fieldPtrFn func(m *TModel) any, valueFn func(ctx context.Context) any) *Builder[TModel] {
	b.opts = append(b.opts, WithSoftDelete[TModel](fieldPtrFn, valueFn))
	return b
}

func (b *Builder[TModel]) AfterSelect(fn func(ctx context.Context, models []*TModel)) *Builder[TModel] {
	b.opts = append(b.opts, WithAfterSelect[TModel](fn))
	return b
}

func (b *Builder[TModel]) Build() (*repository[TModel], error) {
	if b.columns == nil || len(b.columns.AsSlice()) < 1 {
		return nil, errors.New("no columns found")
	}
	return New(b.table, b.columns, b.opts...)
}
