package gerpo

import (
	"context"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/filter"
	"github.com/insei/gerpo/types"
)

type repository[TModel any] struct {
	table        string
	model        *TModel
	fields       fmap.Storage
	columns      *types.ColumnsStorage
	beforeInsert func(ctx context.Context, model *TModel)
	beforeUpdate func(ctx context.Context, model *TModel)
	afterSelect  func(ctx context.Context, models []*TModel)
	softDelete   map[types.Column]func(ctx context.Context) any
	leftJoins    func(ctx context.Context) string
}

func New[TModel any](table string, columns *types.ColumnsStorage, opts ...Option[TModel]) (*repository[TModel], error) {
	model, fields, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	repo := &repository[TModel]{
		table:      table,
		model:      model,
		fields:     fields,
		columns:    columns,
		softDelete: make(map[types.Column]func(ctx context.Context) any),
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

func (r *repository[TModel]) GetFirst(ctx context.Context, qFn func(m *TModel, b filter.Target[TModel])) *TModel {
	qBuilder := filter.NewQueryBuilder[TModel](r.model, r.fields, r.columns, ctx)
	qFn(r.model, qBuilder)
	whereSQL, values := qBuilder.ToSQL()
	selectSQL := "SELECT "
	for _, cl := range r.columns.AsSlice() {
		selectSQL += cl.ToSQL(ctx) + ", "
	}
	_, _ = whereSQL, values
	return r.model
}
