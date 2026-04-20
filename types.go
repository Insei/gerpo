package gerpo

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
)

var (
	ErrNotFound             = executor.ErrNoRows
	ErrApplyQuery           = fmt.Errorf("failed to apply query")
	ErrApplyPersistentQuery = fmt.Errorf("failed to apply persistent query")
)

// Repository represents a generic data repository interface for managing models in the database.
type Repository[TModel any] interface {
	// GetColumns returns the column storage associated with the repository.
	GetColumns() types.ColumnsStorage
	// GetFirst retrieves the first record matching the query conditions.
	GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstHelper[TModel])) (model *TModel, err error)
	// GetList retrieves a list of records matching the query conditions.
	GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListHelper[TModel])) (models []*TModel, err error)
	// Count returns the count of records matching the query conditions.
	Count(ctx context.Context, qFns ...func(m *TModel, h query.CountHelper[TModel])) (count uint64, err error)
	// Insert adds a new record to the database using the provided model and query options.
	Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertHelper[TModel])) (err error)
	// Update modifies an existing record in the database based on the provided model and query options.
	Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateHelper[TModel])) (count int64, err error)
	// Delete removes records from the database based on the query conditions and returns the count of deleted records.
	Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteHelper[TModel])) (count int64, err error)
}

// Builder represents a generic interface for building and configuring a repository for a specific model type.
type Builder[TModel any] interface {
	// WithQuery applies a persistent query function to customize the query process for the model.
	WithQuery(queryFn func(m *TModel, h query.PersistentHelper[TModel])) Builder[TModel]
	// WithBeforeInsert registers a hook called before the INSERT SQL. Returning a
	// non-nil error aborts the call; the SQL does not run.
	WithBeforeInsert(fn func(ctx context.Context, m *TModel) error) Builder[TModel]
	// WithBeforeUpdate registers a hook called before the UPDATE SQL. Returning a
	// non-nil error aborts the call; the SQL does not run.
	WithBeforeUpdate(fn func(ctx context.Context, m *TModel) error) Builder[TModel]
	// WithAfterSelect registers a hook called after GetFirst/GetList with the
	// scanned models. A non-nil error is surfaced after the rows are already
	// fetched.
	WithAfterSelect(fn func(ctx context.Context, models []*TModel) error) Builder[TModel]
	// WithAfterInsert registers a hook called after a successful INSERT.
	// A non-nil error is surfaced after the row has been written — use for
	// cascade inserts in the same ctx-bound transaction; the caller decides
	// whether to roll back.
	WithAfterInsert(fn func(ctx context.Context, m *TModel) error) Builder[TModel]
	// WithAfterUpdate registers a hook called after a successful UPDATE.
	// Same error contract as WithAfterInsert.
	WithAfterUpdate(fn func(ctx context.Context, m *TModel) error) Builder[TModel]
	// WithErrorTransformer allows customizing or wrapping errors during repository operations.
	WithErrorTransformer(fn func(err error) error) Builder[TModel]
	// WithTracer installs a tracing hook called around every Repository operation.
	WithTracer(tracer Tracer) Builder[TModel]
	// WithSoftDeletion configures soft deletion behavior for the model using the provided function and SoftDeletionBuilder.
	WithSoftDeletion(fn func(m *TModel, softDeletion *SoftDeletionBuilder[TModel])) Builder[TModel]
	// Build finalizes and constructs the configured repository for the model.
	Build() (Repository[TModel], error)
}
