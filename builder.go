package gerpo

import (
	"context"
	"errors"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
)

type builder[TModel any] struct {
	adapter         executor.Adapter
	executorOptions []executor.Option
	table           string
	opts            []Option[TModel]
	columnBuilderFn func(m *TModel, columns *ColumnBuilder[TModel])
}

type TableChooser[TModel any] interface {
	Table(table string) ColumnsAppender[TModel]
}

// AdapterChooser is the first step in the fluent builder chain returned by
// gerpo.New[T](). Pick the executor.Adapter that wraps your SQL driver (pgx5,
// pgx4, database/sql) with Adapter(...); executor-level options (cache,
// tracing) attach here too.
type AdapterChooser[TModel any] interface {
	Adapter(a executor.Adapter, opts ...executor.Option) TableChooser[TModel]
}

type ColumnsAppender[TModel any] interface {
	Columns(fn func(m *TModel, columns *ColumnBuilder[TModel])) Builder[TModel]
}

// New starts the fluent builder for a Repository[TModel]. The returned chain
// is AdapterChooser → TableChooser → ColumnsAppender → Builder; finalize with
// Build() to receive the Repository.
func New[TModel any]() AdapterChooser[TModel] {
	return &builder[TModel]{}
}

// Table sets the name of the database table to be used for the model and returns a ColumnsAppender for further configuration.
func (b *builder[TModel]) Table(table string) ColumnsAppender[TModel] {
	b.table = table
	return b
}

// Adapter binds the executor.Adapter (wrapping pgx5, pgx4, or database/sql)
// that the repository will execute through; executor-level options — cache
// storage, tracing, etc. — are applied here as well.
func (b *builder[TModel]) Adapter(a executor.Adapter, opts ...executor.Option) TableChooser[TModel] {
	b.adapter = a
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
// Returning a non-nil error aborts the INSERT — the SQL does not run.
func (b *builder[TModel]) WithBeforeInsert(fn func(ctx context.Context, m *TModel) error) Builder[TModel] {
	b.opts = append(b.opts, WithBeforeInsert[TModel](fn))
	return b
}

// WithBeforeInsertMany registers a function executed before a batched INSERT.
// The callback receives the full slice in one call — a non-nil error aborts
// the call without running any SQL.
func (b *builder[TModel]) WithBeforeInsertMany(fn func(ctx context.Context, models []*TModel) error) Builder[TModel] {
	b.opts = append(b.opts, WithBeforeInsertMany[TModel](fn))
	return b
}

// WithBeforeUpdate registers a function to be executed before performing an update operation on the model in the database.
// Returning a non-nil error aborts the UPDATE — the SQL does not run.
func (b *builder[TModel]) WithBeforeUpdate(fn func(ctx context.Context, m *TModel) error) Builder[TModel] {
	b.opts = append(b.opts, WithBeforeUpdate[TModel](fn))
	return b
}

// WithAfterSelect registers a callback executed after GetFirst/GetList with the
// scanned models. A non-nil error is surfaced to the caller after the rows are
// already fetched.
func (b *builder[TModel]) WithAfterSelect(fn func(ctx context.Context, models []*TModel) error) Builder[TModel] {
	b.opts = append(b.opts, WithAfterSelect[TModel](fn))
	return b
}

// WithAfterUpdate registers a callback executed after a successful UPDATE. A
// non-nil error is surfaced after the row was already modified — the caller
// decides whether to roll back an ambient transaction.
func (b *builder[TModel]) WithAfterUpdate(fn func(ctx context.Context, m *TModel) error) Builder[TModel] {
	b.opts = append(b.opts, WithAfterUpdate[TModel](fn))
	return b
}

// WithAfterInsert registers a callback executed after a successful INSERT. A
// non-nil error is surfaced after the row was already written — the caller
// decides whether to roll back an ambient transaction.
func (b *builder[TModel]) WithAfterInsert(fn func(ctx context.Context, m *TModel) error) Builder[TModel] {
	b.opts = append(b.opts, WithAfterInsert[TModel](fn))
	return b
}

// WithAfterInsertMany registers a callback executed after a successful batched
// INSERT. The callback receives the full slice — use for cascade inserts in
// one batched child query rather than calling the single-row hook per parent.
func (b *builder[TModel]) WithAfterInsertMany(fn func(ctx context.Context, models []*TModel) error) Builder[TModel] {
	b.opts = append(b.opts, WithAfterInsertMany[TModel](fn))
	return b
}

// WithErrorTransformer registers a function to transform or customize errors during repository operations.
func (b *builder[TModel]) WithErrorTransformer(fn func(err error) error) Builder[TModel] {
	b.opts = append(b.opts, WithErrorTransformer[TModel](fn))
	return b
}

// WithTracer installs a tracing hook called around every Repository operation.
// The hook receives SpanInfo carrying the operation name and the bound table.
// Pass nil to disable tracing (this is the default).
func (b *builder[TModel]) WithTracer(tracer Tracer) Builder[TModel] {
	b.opts = append(b.opts, WithTracer[TModel](tracer))
	return b
}

// Build finalizes the builder configuration and returns a Repository instance or an error if essential elements are missing.
func (b *builder[TModel]) Build() (Repository[TModel], error) {
	if b.adapter == nil {
		return nil, errors.New("no adapter found")
	}
	if b.table == "" {
		return nil, errors.New("no table found")
	}
	exec := executor.New[TModel](b.adapter, b.executorOptions...)
	return newRepository(exec, b.table, b.columnBuilderFn, b.opts...)
}
