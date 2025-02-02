package gerpo

import (
	"context"
	"errors"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
)

var ErrNotFound = errors.New("model(s) not found in the repository")

// Repository represents a generic data repository interface for managing models in the database.
type Repository[TModel any] interface {
	// GetColumns returns the column storage associated with the repository.
	GetColumns() types.ColumnsStorage
	// Tx creates a new repository instance bound to the provided database transaction.
	Tx(tx *executor.Tx) (Repository[TModel], error)
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
	// BeforeInsert registers a function to modify the model before insert operations.
	BeforeInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	// BeforeUpdate registers a function to modify the model before update operations.
	BeforeUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	// AfterSelect registers a function to process or transform models after selection operations.
	AfterSelect(fn func(ctx context.Context, models []*TModel)) Builder[TModel]
	// AfterInsert registers a function to process or transform the model after insert operations.
	AfterInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	// AfterUpdate registers a function to process or transform the model after update operations.
	AfterUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	// WithErrorTransformer allows customizing or wrapping errors during repository operations.
	WithErrorTransformer(fn func(err error) error) Builder[TModel]
	// Build finalizes and constructs the configured repository for the model.
	Build() (Repository[TModel], error)
}
