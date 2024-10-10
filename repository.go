package gerpo

import (
	"context"
	dbsql "database/sql"
	"time"

	"github.com/insei/gerpo/query"
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
	helperFactory *query.HelperFactory[TModel]

	columns   *types.ColumnsStorage
	leftJoins func(ctx context.Context) string

	beforeInsert func(ctx context.Context, model *TModel)
	beforeUpdate func(ctx context.Context, model *TModel)
	afterSelect  func(ctx context.Context, models []*TModel)

	softDelete map[types.Column]func(ctx context.Context) any

	table string

	executor *sql.Executor[TModel]
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
	if repo.leftJoins == nil {
		repo.leftJoins = func(_ context.Context) string { return "" }
	}
}

func New[TModel any](db *dbsql.DB, placeholder sql.Placeholder, table string, columns *types.ColumnsStorage, opts ...Option[TModel]) (*repository[TModel], error) {
	model, _, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	//apiConnector, err := query.NewAPIConnectorFactory[test, test](columns)
	//if err != nil {
	//	return nil, err
	//}
	helperFactory := query.NewHelperFactory(model, columns, nil)
	repo := &repository[TModel]{
		executor:      sql.NewExecutor[TModel](db, placeholder),
		helperFactory: helperFactory,
		table:         table,
		columns:       columns,
		softDelete:    make(map[types.Column]func(ctx context.Context) any),
	}
	for _, opt := range opts {
		opt.apply(repo)
	}
	replaceNilCallbacks(repo)
	return repo, nil
}

func (r *repository[TModel]) GetFirst(ctx context.Context, qFns ...func(m *TModel, h *query.Helper[TModel])) (model *TModel, err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	h := r.helperFactory.New()
	for _, fn := range qFns {
		h.Handle(fn)
	}
	h.Apply(strSQLBuilder)
	model, err = r.executor.GetOne(ctx, strSQLBuilder)
	if err != nil {
		return nil, err
	}
	r.afterSelect(ctx, []*TModel{model})
	return model, nil
}

func (r *repository[TModel]) GetList(ctx context.Context, qFns ...func(m *TModel, h *query.Helper[TModel])) (models []*TModel, err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	h := r.helperFactory.New()
	for _, fn := range qFns {
		h.Handle(fn)
	}
	h.Apply(strSQLBuilder)
	models, err = r.executor.GetMultiple(ctx, strSQLBuilder)
	if err != nil {
		return nil, err
	}
	r.afterSelect(ctx, models)
	return models, nil
}

func (r *repository[TModel]) Count(ctx context.Context, qFns ...func(m *TModel, h *query.Helper[TModel])) (count uint64, err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	h := r.helperFactory.New()
	for _, fn := range qFns {
		h.Handle(fn)
	}
	h.Apply(strSQLBuilder)
	return r.executor.Count(ctx, strSQLBuilder)
}

func (r *repository[TModel]) Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h *query.Helper[TModel])) (err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	r.beforeInsert(ctx, model)
	h := r.helperFactory.New()
	for _, fn := range qFns {
		h.Handle(fn)
	}
	h.Apply(strSQLBuilder)
	return r.executor.InsertOne(ctx, model, strSQLBuilder)
}

func (r *repository[TModel]) Update(ctx context.Context, model *TModel, qFn func(m *TModel, h *query.Helper[TModel])) (err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	r.beforeUpdate(ctx, model)
	h := r.helperFactory.New()
	h.Handle(qFn)
	h.Apply(strSQLBuilder)
	return r.executor.Update(ctx, model, strSQLBuilder)
}

func (r *repository[TModel]) Delete(ctx context.Context, qFn func(m *TModel, h *query.Helper[TModel])) (count uint64, err error) {
	strSQLBuilder := sql.NewStringBuilder(ctx, r.table)
	h := r.helperFactory.New()
	h.Handle(qFn)
	h.Apply(strSQLBuilder)
	return r.executor.Delete(ctx, strSQLBuilder)
}
