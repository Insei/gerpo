package gerpo

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type test struct {
	ID          int
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Name        string
	Age         int
	Bool        bool
	DeletedAt   *time.Time
	DeletedTest bool
}

type repository[TModel any] struct {
	// Callbacks and Hooks
	beforeInsert     func(ctx context.Context, model *TModel)
	beforeUpdate     func(ctx context.Context, model *TModel)
	afterInsert      func(ctx context.Context, model *TModel)
	afterUpdate      func(ctx context.Context, model *TModel)
	afterDelete      func(ctx context.Context, model []*TModel)
	afterSelect      func(ctx context.Context, models []*TModel)
	errorTransformer func(err error) error

	deleteFn        func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error)
	getDeleteModels func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) ([]*TModel, error)

	// Columns and fields
	columns *types.ColumnsStorage

	// SQL Query, execution and dependency
	strSQLBuilderFactory sql.StringBuilderFactory
	executor             executor.Executor[TModel]
	query                *query.Bundle[TModel]
}

func replaceNilCallbacks[TModel any](repo *repository[TModel]) {
	if repo.beforeInsert == nil {
		repo.beforeInsert = func(_ context.Context, model *TModel) {}
	}
	if repo.beforeUpdate == nil {
		repo.beforeUpdate = func(_ context.Context, model *TModel) {}
	}
	if repo.afterInsert == nil {
		repo.afterInsert = func(_ context.Context, model *TModel) {}
	}
	if repo.afterUpdate == nil {
		repo.afterUpdate = func(_ context.Context, model *TModel) {}
	}
	if repo.afterSelect == nil {
		repo.afterSelect = func(_ context.Context, models []*TModel) {}
	}
	if repo.afterDelete == nil {
		repo.getDeleteModels = func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) ([]*TModel, error) {
			return []*TModel{}, nil
		}
		repo.afterDelete = func(ctx context.Context, model []*TModel) {}
	}
	if repo.errorTransformer == nil {
		repo.errorTransformer = func(err error) error { return err }
	}
}

func New[TModel any](db *dbsql.DB, table string, columnsFn func(m *TModel, builder *ColumnBuilder[TModel]), sdColumnsFn func(m *TModel, builder *SoftDeleteBuilder[TModel]), opts ...Option[TModel]) (Repository[TModel], error) {
	model, fields, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	columnsBuilder := newColumnBuilder(table, model, fields)
	columnsFn(model, columnsBuilder)
	columns := columnsBuilder.build()

	if len(columns.AsSlice()) < 1 {
		return nil, fmt.Errorf("failed to create repository with empty columns")
	}
	repo := &repository[TModel]{
		columns:              columns,
		query:                query.NewBundle(model, columns),
		executor:             executor.New[TModel](db),
		strSQLBuilderFactory: sql.NewStringBuilderFactory(table, columns),
	}

	repo.deleteFn = getDeleteFn(repo, model, sdColumnsFn)

	for _, opt := range opts {
		opt.apply(repo)
	}

	useDeleteHook(repo)
	replaceNilCallbacks(repo)
	return repo, nil
}

func (r *repository[TModel]) GetColumns() *types.ColumnsStorage {
	return r.columns
}

func (r *repository[TModel]) Tx(tx *executor.Tx) (Repository[TModel], error) {
	txExecutor, err := r.executor.Tx(tx)
	if err != nil {
		return nil, err
	}
	repocp := *r
	repocp.executor = txExecutor
	return &repocp, nil
}

func (r *repository[TModel]) GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstUserHelper[TModel])) (model *TModel, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.query.ApplyGetFirst(strSQLBuilder, qFns...)
	model, err = r.executor.GetOne(ctx, strSQLBuilder)
	if err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, r.errorTransformer(fmt.Errorf("%w: %w", ErrNotFound, err))
		}
		return nil, err
	}
	r.afterSelect(ctx, []*TModel{model})
	return model, nil
}

func (r *repository[TModel]) GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListUserHelper[TModel])) (models []*TModel, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.query.ApplyGetList(strSQLBuilder, qFns...)
	models, err = r.executor.GetMultiple(ctx, strSQLBuilder)
	if err != nil {
		return nil, r.errorTransformer(err)
	}
	r.afterSelect(ctx, models)
	return models, nil
}

func (r *repository[TModel]) Count(ctx context.Context, qFns ...func(m *TModel, h query.CountUserHelper[TModel])) (count uint64, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.query.ApplyCount(strSQLBuilder, qFns...)
	count, err = r.executor.Count(ctx, strSQLBuilder)
	if err != nil {
		return 0, r.errorTransformer(err)
	}
	return count, nil
}

func (r *repository[TModel]) Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertUserHelper[TModel])) (err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.beforeInsert(ctx, model)
	r.query.ApplyInsert(strSQLBuilder, qFns...)
	err = r.executor.InsertOne(ctx, model, strSQLBuilder)
	if err != nil {
		return r.errorTransformer(err)
	}
	r.afterInsert(ctx, model)
	return nil
}

func (r *repository[TModel]) Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateUserHelper[TModel])) (err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.beforeUpdate(ctx, model)
	r.query.ApplyUpdate(strSQLBuilder, qFns...)
	updatedCount, err := r.executor.Update(ctx, model, strSQLBuilder)
	if err != nil {
		return r.errorTransformer(err)
	}
	if updatedCount < 1 {
		return r.errorTransformer(fmt.Errorf("nothing to update: %w", ErrNotFound))
	}
	r.afterUpdate(ctx, model)
	return nil
}

func (r *repository[TModel]) Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error) {
	models, err := r.getDeleteModels(ctx, qFns...)
	if err != nil {
		return 0, r.errorTransformer(err)
	}

	count, err = r.deleteFn(ctx, qFns...)
	if err != nil {
		return 0, r.errorTransformer(err)
	}
	if count < 1 {
		return 0, r.errorTransformer(fmt.Errorf("nothing to delete: %w", ErrNotFound))
	}
	r.afterDelete(ctx, models)
	return count, r.errorTransformer(err)
}
