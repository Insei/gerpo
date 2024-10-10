package gerpo

import (
	"context"
	dbsql "database/sql"
	"errors"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type Builder[TModel any] struct {
	db          *dbsql.DB
	driver      string
	placeholder sql.Placeholder
	table       string
	opts        []Option[TModel]
	model       *TModel
	fields      fmap.Storage
	columns     *types.ColumnsStorage
}

func NewBuilder[TModel any](db *dbsql.DB, placeholder sql.Placeholder, table string) (*Builder[TModel], error) {
	model, fields, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	return &Builder[TModel]{
		table:       table,
		db:          db,
		placeholder: placeholder,
		model:       model,
		fields:      fields,
	}, nil
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

//func (b *Builder[TModel]) SoftDelete(fieldPtrFn func(m *TModel) any, valueFn func(ctx context.Context) any) *Builder[TModel] {
//	b.opts = append(b.opts, WithSoftDelete[TModel](fieldPtrFn, valueFn))
//	return b
//}

func (b *Builder[TModel]) AfterSelect(fn func(ctx context.Context, models []*TModel)) *Builder[TModel] {
	b.opts = append(b.opts, WithAfterSelect[TModel](fn))
	return b
}

func (b *Builder[TModel]) Build() (*Repository[TModel], error) {
	if b.columns == nil || len(b.columns.AsSlice()) < 1 {
		return nil, errors.New("no columns found")
	}
	return New(b.db, b.placeholder, b.table, b.columns, b.opts...)
}
