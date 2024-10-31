package gerpo

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"
	"time"

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

type testDto struct {
	ID        int        `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Name      string     `json:"name"`
	Age       int        `json:"age"`
	Bool      bool       `json:"bool"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type repository[TModel any] struct {
	// Callbacks and Hooks
	beforeInsert     func(ctx context.Context, model *TModel)
	beforeUpdate     func(ctx context.Context, model *TModel)
	afterInsert      func(ctx context.Context, model *TModel)
	afterUpdate      func(ctx context.Context, model *TModel)
	afterSelect      func(ctx context.Context, models []*TModel)
	errorTransformer func(err error) error

	// Columns and fields
	strSQLBuilderFactory sql.StringBuilderFactory
	columns              *types.ColumnsStorage
	sdColumnsMap         map[types.Column]SoftDeleteGetValueFn

	// SQL Query, execution and dependency
	executor *sql.Executor[TModel]
	query    *query.Bundle[TModel]
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

func New[TModel any](db *dbsql.DB, table string, columnsFn func(m *TModel, builder *ColumnBuilder[TModel]), sdColumnsFn func(m *TModel, builder *SoftDeleteBuilder[TModel]), opts ...Option[TModel]) (Repository[TModel], error) {
	model, fields, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	columnsBuilder := newColumnBuilder(table, model, fields)
	columnsFn(model, columnsBuilder)
	columns := columnsBuilder.build()

	var sdColumnsMap map[types.Column]SoftDeleteGetValueFn
	if sdColumnsFn != nil {
		sdColumnsBuilder := newSoftDeleteBuilder(model, columns)
		sdColumnsFn(model, sdColumnsBuilder)
		sdColumnsMap = sdColumnsBuilder.build()
	}

	if len(columns.AsSlice()) < 1 {
		return nil, fmt.Errorf("failed to create repository with empty columns")
	}
	repo := &repository[TModel]{
		columns:              columns,
		sdColumnsMap:         sdColumnsMap,
		query:                query.NewBundle(model, columns),
		executor:             sql.NewExecutor[TModel](db),
		strSQLBuilderFactory: sql.NewStringBuilderFactory(table, columns),
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
	deleteFn := r.getDeleteFn()
	count, err = deleteFn(ctx, qFns...)
	if err != nil {
		return 0, r.errorTransformer(err)
	}
	if count < 1 {
		return 0, fmt.Errorf("nothing to delete: %w", ErrNotFound)
	}
	return count, r.errorTransformer(err)
}

func (r *repository[TModel]) getDeleteFn() func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error) {
	if r.sdColumnsMap != nil && len(r.sdColumnsMap) > 0 {
		return r.softDelete
	}
	return r.delete
}

func (r *repository[TModel]) delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.query.ApplyDelete(strSQLBuilder, qFns...)
	return r.executor.Delete(ctx, strSQLBuilder)
}

func (r *repository[TModel]) softDelete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)

	model := new(TModel)
	for col, getValFn := range r.sdColumnsMap {
		col.GetField().Set(model, getValFn(ctx))
	}

	r.query.ApplyUpdate(strSQLBuilder, func(m *TModel, h query.UpdateUserHelper[TModel]) {
		pointers := r.getExcludePointers(m, r.columns.AsSlice(), r.sdColumnsMap)
		h.Exclude(pointers...)
	})
	// Apply WHERE filters
	r.query.ApplyDelete(strSQLBuilder, qFns...)

	return r.executor.Update(ctx, model, strSQLBuilder)
}

func (r *repository[TModel]) getExcludePointers(m *TModel, allCols []types.Column, sdCols map[types.Column]SoftDeleteGetValueFn) []any {
	pointers := make([]any, 0, len(allCols)-len(sdCols))
	for _, col := range allCols {
		_, ok := sdCols[col]
		if !ok {
			pointers = append(pointers, col.GetField().GetPtr(m))
		}
	}

	return pointers
}
