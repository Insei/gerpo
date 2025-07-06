package gerpo

import (
	"context"
	"errors"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
)

type builder[TModel any] struct {
	db              executor.DBAdapter
	executorOptions []executor.Option
	table           string
	opts            []Option[TModel]
	columns         *types.ColumnsStorage
	columnBuilderFn func(m *TModel, columns *ColumnBuilder[TModel])
}

type TableChooser[TModel any] interface {
	Table(table string) ColumnsAppender[TModel]
}

type ExecutorChooser[TModel any] interface {
	DB(db executor.DBAdapter, opts ...executor.Option) TableChooser[TModel]
}

type ColumnsAppender[TModel any] interface {
	Columns(fn func(m *TModel, columns *ColumnBuilder[TModel])) Builder[TModel]
}

// NewBuilder creates a new instance of the repository builder using the specified generic type for model TModel.
func NewBuilder[TModel any]() ExecutorChooser[TModel] {
	return &builder[TModel]{}
}

// Table sets the name of the database table to be used for the model and returns a ColumnsAppender for further configuration.
func (b *builder[TModel]) Table(table string) ColumnsAppender[TModel] {
	b.table = table
	return b
}

// DB sets the database connection to be used for the builder and returns a TableChooser for further configuration.
func (b *builder[TModel]) DB(db executor.DBAdapter, opts ...executor.Option) TableChooser[TModel] {
	b.db = db
	b.executorOptions = opts
	return b
}

// Columns sets a custom function to configure columns for the model and returns the builder for further customization.
func (b *builder[TModel]) Columns(fn func(m *TModel, columns *ColumnBuilder[TModel])) Builder[TModel] {
	b.columnBuilderFn = fn
	return b
}

func (b *builder[TModel]) WithSoftDeletion(fn func(m *TModel, softDeletion *SoftDeletionBuilder[TModel])) Builder[TModel] {
	b.opts = append(b.opts, WithSoftDeletion[TModel](fn))
	return b
}

// WithQuery registers a persistent query function to modify or customize the query behavior in the repository configuration.
func (b *builder[TModel]) WithQuery(queryFn func(m *TModel, h query.PersistentHelper[TModel])) Builder[TModel] {
	b.opts = append(b.opts, WithQuery[TModel](queryFn))
	return b
}

// WithBeforeInsert registers a function that is executed before performing an insert operation on the model in the database.
func (b *builder[TModel]) WithBeforeInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithBeforeInsert[TModel](fn))
	return b
}

// WithBeforeUpdate registers a function to be executed before performing an update operation on the model in the database.
func (b *builder[TModel]) WithBeforeUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithBeforeUpdate[TModel](fn))
	return b
}

// WithAfterSelect registers a callback function to be executed after models are retrieved through a select operation.
func (b *builder[TModel]) WithAfterSelect(fn func(ctx context.Context, models []*TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithAfterSelect[TModel](fn))
	return b
}

// WithAfterUpdate registers a callback function to be executed after an update operation is performed on the model.
func (b *builder[TModel]) WithAfterUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithAfterUpdate[TModel](fn))
	return b
}

// WithAfterInsert registers a callback function to be executed after an insert operation is performed on the model.
func (b *builder[TModel]) WithAfterInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel] {
	b.opts = append(b.opts, WithAfterInsert[TModel](fn))
	return b
}

// WithErrorTransformer registers a function to transform or customize errors during repository operations.
func (b *builder[TModel]) WithErrorTransformer(fn func(err error) error) Builder[TModel] {
	b.opts = append(b.opts, WithErrorTransformer[TModel](fn))
	return b
}

// Build finalizes the builder configuration and returns a Repository instance or an error if essential elements are missing.
func (b *builder[TModel]) Build() (Repository[TModel], error) {
	if b.db == nil {
		return nil, errors.New("no database found")
	}
	if b.table == "" {
		return nil, errors.New("no table found")
	}
	exec := executor.New[TModel](b.db, b.executorOptions...)
	return New(exec, b.table, b.columnBuilderFn, b.opts...)
}
