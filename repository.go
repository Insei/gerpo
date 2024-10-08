package gerpo

import (
	"context"
	"database/sql"
	"time"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/query"
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
	whereBuilderFactory  *query.WhereBuilderFactory[TModel]
	orderBuilderFactory  *query.OrderBuilderFactory[TModel]
	selectBuilderFactory *query.SelectBuilderFactory[TModel]
	apiConnectorFactory  *query.APIConnectorFactory

	model     *TModel
	fields    fmap.Storage
	columns   *types.ColumnsStorage
	leftJoins func(ctx context.Context) string

	beforeInsert func(ctx context.Context, model *TModel)
	beforeUpdate func(ctx context.Context, model *TModel)
	afterSelect  func(ctx context.Context, models []*TModel)

	softDelete map[types.Column]func(ctx context.Context) any

	table string

	db          *sql.DB
	driver      string
	placeholder string
}

func New[TModel any](table string, columns *types.ColumnsStorage, opts ...Option[TModel]) (*repository[TModel], error) {
	model, fields, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	apiConnector, err := query.NewAPIConnectorFactory[test, test](columns)
	if err != nil {
		return nil, err
	}
	repo := &repository[TModel]{
		whereBuilderFactory:  query.NewWhereBuilderFabric[TModel](model, columns),
		orderBuilderFactory:  query.NewOrderBuilderFabric[TModel](model, columns),
		selectBuilderFactory: query.NewSelectBuilderFabric[TModel](model, columns),
		apiConnectorFactory:  apiConnector,
		table:                table,
		model:                model,
		fields:               fields,
		columns:              columns,
		softDelete:           make(map[types.Column]func(ctx context.Context) any),
	}
	for _, opt := range opts {
		opt.apply(repo)
	}
	// Make all nil functions as empty for easy use later
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
	return repo, nil
}

func (r *repository[TModel]) GetFirst(ctx context.Context, qFn func(m *TModel, h *query.Helper[TModel])) *TModel {
	strSQLBuilder := query.NewStringSQLBuilder(ctx)
	selectBuilder := r.selectBuilderFactory.New()
	whereBuilder := r.whereBuilderFactory.New()
	orderBuilder := r.orderBuilderFactory.New()
	apiBuilder := r.apiConnectorFactory.New()
	qFn(r.model, query.NewHelper(selectBuilder, whereBuilder, orderBuilder, apiBuilder))
	whereBuilder.Apply(strSQLBuilder.WhereBuilder())
	apiBuilder.Apply(strSQLBuilder.WhereBuilder())
	orderBuilder.Apply(strSQLBuilder.OrderBuilder())
	selectBuilder.Apply(strSQLBuilder.SelectBuilder())
	return r.model
}
