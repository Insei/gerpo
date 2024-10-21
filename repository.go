package gerpo

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"time"

	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type test struct {
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

	softDelete map[types.Column]func(ctx context.Context) any

	// Columns and fields
	strSQLBuilderFactory sql.StringBuilderFactory

	// SQL Query, execution and dependency
	linqCore   *linq.CoreBuilder
	executor   *sql.Executor[TModel]
	persistent query.PersistentHelper[TModel]
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
	coreBuilder := linq.NewCoreBuilder(model, columns)

	repo := &repository[TModel]{
		persistent:           query.NewPersistentHelper[TModel](coreBuilder),
		executor:             sql.NewExecutor[TModel](db),
		linqCore:             linq.NewCoreBuilder(model, columns),
		strSQLBuilderFactory: sql.NewStringBuilderFactory(table, columns),
		softDelete:           make(map[types.Column]func(ctx context.Context) any),
	}
	for _, opt := range opts {
		opt.apply(repo)
	}
	replaceNilCallbacks(repo)
	return repo, nil
}

func (r *repository[TModel]) applyPersistentQuery(sqlBuilder *sql.StringBuilder) {
	r.persistent.Apply(sqlBuilder)
}

func (r *repository[TModel]) GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstUserHelper[TModel])) (model *TModel, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	h := query.NewGetFirstHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	r.persistent.Apply(strSQLBuilder)
	model, err = r.executor.GetOne(ctx, strSQLBuilder)
	if err != nil {
		return nil, err
	}
	r.afterSelect(ctx, []*TModel{model})
	return model, nil
}

func (r *repository[TModel]) GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListUserHelper[TModel])) (models []*TModel, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	h := query.NewGetListHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	r.persistent.Apply(strSQLBuilder)
	models, err = r.executor.GetMultiple(ctx, strSQLBuilder)
	if err != nil {
		return nil, err
	}
	r.afterSelect(ctx, models)
	return models, nil
}

func (r *repository[TModel]) Count(ctx context.Context, qFns ...func(m *TModel, h query.CountUserHelper[TModel])) (count uint64, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	h := query.NewCountHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	r.persistent.Apply(strSQLBuilder)
	return r.executor.Count(ctx, strSQLBuilder)
}

func (r *repository[TModel]) Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertUserHelper[TModel])) (err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.beforeInsert(ctx, model)
	h := query.NewInsertHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	r.persistent.Apply(strSQLBuilder)
	return r.executor.InsertOne(ctx, model, strSQLBuilder)
}

func (r *repository[TModel]) Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateUserHelper[TModel])) (err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	r.beforeUpdate(ctx, model)
	h := query.NewUpdateHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	r.persistent.Apply(strSQLBuilder)
	_, err = r.executor.Update(ctx, model, strSQLBuilder)
	return err
}

func (r *repository[TModel]) Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error) {
	strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
	h := query.NewDeleteHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	r.persistent.Apply(strSQLBuilder)
	return r.executor.Delete(ctx, strSQLBuilder)
}
