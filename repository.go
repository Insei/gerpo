package gerpo

import (
	"context"
	dbsql "database/sql"
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

type Repository[TModel any] struct {
	linqCore *linq.CoreBuilder

	columns   *types.ColumnsStorage
	leftJoins func(ctx context.Context) string

	beforeInsert func(ctx context.Context, model *TModel)
	beforeUpdate func(ctx context.Context, model *TModel)
	afterSelect  func(ctx context.Context, models []*TModel)

	softDelete map[types.Column]func(ctx context.Context) any

	table string

	executor *sql.Executor[TModel]
}

func replaceNilCallbacks[TModel any](repo *Repository[TModel]) {
	if repo.afterSelect == nil {
		repo.afterSelect = func(_ context.Context, models []*TModel) {}
	}
	if repo.beforeInsert == nil {
		repo.beforeInsert = func(_ context.Context, model *TModel) {}
	}
	if repo.beforeUpdate == nil {
		repo.beforeUpdate = func(_ context.Context, model *TModel) {}
	}
	if repo.leftJoins == nil {
		repo.leftJoins = func(_ context.Context) string { return "" }
	}
}

func New[TModel any](db *dbsql.DB, table string, columns *types.ColumnsStorage, opts ...Option[TModel]) (*Repository[TModel], error) {
	model, _, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	repo := &Repository[TModel]{
		executor:   sql.NewExecutor[TModel](db, sql.DeterminePlaceHolder(db)),
		linqCore:   linq.NewCoreBuilder(model, columns),
		table:      table,
		columns:    columns,
		softDelete: make(map[types.Column]func(ctx context.Context) any),
	}
	for _, opt := range opts {
		opt.apply(repo)
	}
	replaceNilCallbacks(repo)
	return repo, nil
}

func (r *Repository[TModel]) GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstUserHelper[TModel])) (model *TModel, err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	h := query.NewGetFirstHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	model, err = r.executor.GetOne(ctx, strSQLBuilder)
	if err != nil {
		return nil, err
	}
	r.afterSelect(ctx, []*TModel{model})
	return model, nil
}

func (r *Repository[TModel]) GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListUserHelper[TModel])) (models []*TModel, err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	h := query.NewGetListHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	models, err = r.executor.GetMultiple(ctx, strSQLBuilder)
	if err != nil {
		return nil, err
	}
	r.afterSelect(ctx, models)
	return models, nil
}

func (r *Repository[TModel]) Count(ctx context.Context, qFns ...func(m *TModel, h query.CountUserHelper[TModel])) (count uint64, err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	h := query.NewCountHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	return r.executor.Count(ctx, strSQLBuilder)
}

func (r *Repository[TModel]) Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertUserHelper[TModel])) (err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	r.beforeInsert(ctx, model)
	h := query.NewInsertHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	return r.executor.InsertOne(ctx, model, strSQLBuilder)
}

func (r *Repository[TModel]) Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateUserHelper[TModel])) (err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	r.beforeUpdate(ctx, model)
	h := query.NewUpdateHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	return r.executor.Update(ctx, model, strSQLBuilder)
}

func (r *Repository[TModel]) Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count uint64, err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	h := query.NewDeleteHelper[TModel](r.linqCore)
	h.HandleFn(qFns...)
	h.Apply(strSQLBuilder)
	return r.executor.Delete(ctx, strSQLBuilder)
}
