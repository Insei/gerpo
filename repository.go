package gerpo

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"time"

	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type test struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt *time.Time
	Name      string
	Age       int
	Bool      bool
	DeletedAt *time.Time
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
	beforeInsert func(ctx context.Context, model *TModel)
	beforeUpdate func(ctx context.Context, model *TModel)
	afterSelect  func(ctx context.Context, models []*TModel)

	// Columns and fields
	strSQLBuilderFactory sql.StringBuilderFactory
	columns              *types.ColumnsStorage

	// SQL Query, execution and dependency
	executor *sql.Executor[TModel]
	query    *query.Bundle[TModel]
}

func replaceNilCallbacks[TModel any](repo *repository[TModel]) {
	if repo.afterSelect == nil {
		repo.afterSelect = func(_ context.Context, models []*TModel) {}
	}
	if repo.beforeInsert == nil {
		repo.beforeInsert = func(_ context.Context, model *TModel) {}
	}
	if repo.beforeUpdate == nil {
		repo.beforeUpdate = func(_ context.Context, model *TModel) {}
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
		columns:              columns,
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
		return nil, err
	}
	r.afterSelect(ctx, models)
	return models, nil
}

func (r *repository[TModel]) Count(ctx context.Context, qFns ...func(m *TModel, h query.CountUserHelper[TModel])) (count uint64, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.query.ApplyCount(strSQLBuilder, qFns...)
	return r.executor.Count(ctx, strSQLBuilder)
}

func (r *repository[TModel]) Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertUserHelper[TModel])) (err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.beforeInsert(ctx, model)
	r.query.ApplyInsert(strSQLBuilder, qFns...)
	return r.executor.InsertOne(ctx, model, strSQLBuilder)
}

func (r *repository[TModel]) Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateUserHelper[TModel])) (err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.beforeUpdate(ctx, model)
	r.query.ApplyUpdate(strSQLBuilder, qFns...)
	_, err = r.executor.Update(ctx, model, strSQLBuilder)
	return err
}

func (r *repository[TModel]) Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.query.ApplyDelete(strSQLBuilder, qFns...)
	return r.executor.Delete(ctx, strSQLBuilder)
}
