package gerpo

import (
	"context"
	dbsql "database/sql"
	"errors"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type Builder[TModel any] interface {
	WithQuery(queryFn func(m *TModel, h query.PersistentUserHelper[TModel])) Builder[TModel]
	BeforeInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	BeforeUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	AfterSelect(fn func(ctx context.Context, models []*TModel)) Builder[TModel]
	AfterUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	AfterInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	Build() (Repository[TModel], error)
}

type builder[TModel any] struct {
	db              *dbsql.DB
	driver          string
	placeholder     sql.Placeholder
	table           string
	opts            []Option[TModel]
	model           *TModel
	fields          fmap.Storage
	columns         *types.ColumnsStorage
	columnBuilderFn func(m *TModel, columns *ColumnBuilder[TModel])
}

type TableChooser[TModel any] interface {
	Table(table string) ColumnsAppender[TModel]
}

type DbChooser[TModel any] interface {
	DB(db *dbsql.DB) TableChooser[TModel]
}

type ColumnsAppender[TModel any] interface {
	Columns(fn func(m *TModel, columns *ColumnBuilder[TModel])) Builder[TModel]
}

func NewBuilder[TModel any]() DbChooser[TModel] {
	model, fields, err := getModelAndFields[TModel]()
	if err != nil {
		panic(err)
	}
	return &builder[TModel]{
		model:  model,
		fields: fields,
	}
}

func (b *builder[TModel]) Table(table string) ColumnsAppender[TModel] {
	b.table = table
	return b
}

func (b *builder[TModel]) DB(db *dbsql.DB) TableChooser[TModel] {
	b.db = db
	return b
}

func (b *builder[TModel]) Columns(fn func(m *TModel, columns *ColumnBuilder[TModel])) Builder[TModel] {
	b.columnBuilderFn = fn
	return b
}

func (b *builder[TModel]) WithQuery(queryFn func(m *TModel, h query.PersistentUserHelper[TModel])) Builder[TModel] {
	b.opts = append(b.opts, WithQuery[TModel](queryFn))
	return b
}

func (b *builder[TModel]) BeforeInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithBeforeInsert[TModel](fn))
	return b
}

func (b *builder[TModel]) BeforeUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithBeforeUpdate[TModel](fn))
	return b
}

func (b *builder[TModel]) AfterSelect(fn func(ctx context.Context, models []*TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithAfterSelect[TModel](fn))
	return b
}

func (b *builder[TModel]) AfterUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithAfterUpdate[TModel](fn))
	return b
}

func (b *builder[TModel]) AfterInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithAfterInsert[TModel](fn))
	return b
}

func (b *builder[TModel]) Build() (Repository[TModel], error) {
	if b.db == nil {
		return nil, errors.New("no database found")
	}
	if b.table == "" {
		return nil, errors.New("no table found")
	}
	return New(b.db, b.table, b.columnBuilderFn, b.opts...)
}
