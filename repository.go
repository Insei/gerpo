package gerpo

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/sqlstmt"
	"github.com/insei/gerpo/types"
)

type repository[TModel any] struct {
	// Callbacks and Hooks
	beforeInsert     func(ctx context.Context, model *TModel)
	beforeUpdate     func(ctx context.Context, model *TModel)
	afterInsert      func(ctx context.Context, model *TModel)
	afterUpdate      func(ctx context.Context, model *TModel)
	afterSelect      func(ctx context.Context, models []*TModel)
	errorTransformer func(err error) error

	// Columns and fields
	baseModel *TModel
	table     string
	columns   *types.ColumnsStorage

	// SQL Query, execution and dependency
	executor   executor.Executor[TModel]
	persistent *query.Persistent[TModel]
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
	if repo.errorTransformer == nil {
		repo.errorTransformer = func(err error) error { return err }
	}
}

func New[TModel any](db *dbsql.DB, table string, columnsFn func(m *TModel, builder *ColumnBuilder[TModel]), opts ...Option[TModel]) (Repository[TModel], error) {
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
		columns:    columns,
		executor:   executor.New[TModel](db),
		table:      table,
		baseModel:  model,
		persistent: query.NewPersistent(model),
	}

	for _, opt := range opts {
		opt.apply(repo)
	}

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

func (r *repository[TModel]) GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstHelper[TModel])) (model *TModel, err error) {
	stmt := sqlstmt.NewGetFirst(ctx, r.table, r.columns)
	q := query.NewGetFirst(r.baseModel)
	q.HandleFn(qFns...)
	r.persistent.Apply(stmt)
	q.Apply(stmt)
	model, err = r.executor.GetOne(ctx, stmt)
	if err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, r.errorTransformer(fmt.Errorf("%w: %w", ErrNotFound, err))
		}
		return nil, err
	}
	r.afterSelect(ctx, []*TModel{model})
	return model, nil
}

func (r *repository[TModel]) GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListHelper[TModel])) (models []*TModel, err error) {
	stmt := sqlstmt.NewGetList(ctx, r.table, r.columns)
	r.persistent.Apply(stmt)
	q := query.NewGetList(r.baseModel)
	q.HandleFn(qFns...)
	q.Apply(stmt)
	models, err = r.executor.GetMultiple(ctx, stmt)
	if err != nil {
		return nil, r.errorTransformer(err)
	}
	r.afterSelect(ctx, models)
	return models, nil
}

func (r *repository[TModel]) Count(ctx context.Context, qFns ...func(m *TModel, h query.CountHelper[TModel])) (count uint64, err error) {
	stmt := sqlstmt.NewCount(ctx, r.table, r.columns)
	q := query.NewCount(r.baseModel)
	q.HandleFn(qFns...)
	r.persistent.Apply(stmt)
	q.Apply(stmt)
	count, err = r.executor.Count(ctx, stmt)
	if err != nil {
		return 0, r.errorTransformer(err)
	}
	return count, nil
}

func (r *repository[TModel]) Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertHelper[TModel])) (err error) {
	r.beforeInsert(ctx, model)
	stmt := sqlstmt.NewInsert(ctx, r.table, r.columns)
	q := query.NewInsert(r.baseModel)
	q.HandleFn(qFns...)
	r.persistent.Apply(stmt)
	q.Apply(stmt)
	err = r.executor.InsertOne(ctx, stmt, model)
	if err != nil {
		return r.errorTransformer(err)
	}
	r.afterInsert(ctx, model)
	return nil
}

func (r *repository[TModel]) Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateHelper[TModel])) (err error) {
	r.beforeUpdate(ctx, model)
	stmt := sqlstmt.NewUpdate(ctx, r.columns, r.table)
	q := query.NewUpdate(r.baseModel)
	q.HandleFn(qFns...)
	r.persistent.Apply(stmt)
	q.Apply(stmt)
	updatedCount, err := r.executor.Update(ctx, stmt, model)
	if err != nil {
		return r.errorTransformer(err)
	}
	if updatedCount < 1 {
		return r.errorTransformer(fmt.Errorf("nothing to update: %w", ErrNotFound))
	}
	r.afterUpdate(ctx, model)
	return nil
}

func (r *repository[TModel]) Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteHelper[TModel])) (count int64, err error) {
	stmt := sqlstmt.NewDelete(ctx, r.table, r.columns)
	q := query.NewDelete(r.baseModel)
	q.HandleFn(qFns...)
	r.persistent.Apply(stmt)
	q.Apply(stmt)
	count, err = r.executor.Delete(ctx, stmt)
	if err != nil {
		return count, r.errorTransformer(err)
	}
	if count < 1 {
		return 0, r.errorTransformer(fmt.Errorf("nothing to delete: %w", ErrNotFound))
	}
	return count, r.errorTransformer(err)
}
