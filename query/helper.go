package query

import (
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type Helper[TModel any] struct {
	factory       *HelperFactory[TModel]
	api           *APIConnector
	whereBuilder  *WhereBuilder[TModel]
	orderBuilder  *OrderBuilder[TModel]
	selectBuilder *SelectBuilder[TModel]
	insertBuilder *InsertBuilder[TModel]
	updateBuilder *UpdateBuilder[TModel]
}

func (h *Helper[TModel]) Select() *SelectBuilder[TModel] {
	return h.selectBuilder
}

func (h *Helper[TModel]) Where() types.WhereTarget[TModel] {
	return h.whereBuilder
}

func (h *Helper[TModel]) OrderBy() types.OrderTarget[TModel] {
	return h.orderBuilder
}

func (h *Helper[TModel]) RESTAPI() *APIConnector {
	return h.api
}

func (h *Helper[TModel]) Handle(fn func(m *TModel, h *Helper[TModel])) {
	if fn == nil {
		return
	}
	fn(h.factory.model, h)
}

func (h *Helper[TModel]) Apply(strSQLBuilder *sql.StringBuilder) {
	h.selectBuilder.Apply(strSQLBuilder.SelectBuilder())
	h.whereBuilder.Apply(strSQLBuilder.WhereBuilder())
	h.orderBuilder.Apply(strSQLBuilder.OrderBuilder())
	h.insertBuilder.Apply(strSQLBuilder.InsertBuilder())
	h.updateBuilder.Apply(strSQLBuilder.UpdateBuilder())
	if h.api != nil {
		h.api.ApplyWhere(strSQLBuilder.WhereBuilder())
	}
}

type HelperFactory[TModel any] struct {
	model         *TModel
	api           *APIConnectorFactory
	whereBuilder  *WhereBuilderFactory[TModel]
	orderBuilder  *OrderBuilderFactory[TModel]
	selectBuilder *SelectBuilderFactory[TModel]
	insertBuilder *InsertBuilderFactory[TModel]
	updateBuilder *UpdateBuilderFactory[TModel]
}

func (f *HelperFactory[TModel]) New() *Helper[TModel] {
	return &Helper[TModel]{
		factory:       f,
		api:           f.api.New(),
		whereBuilder:  f.whereBuilder.New(),
		orderBuilder:  f.orderBuilder.New(),
		selectBuilder: f.selectBuilder.New(),
		insertBuilder: f.insertBuilder.New(),
		updateBuilder: f.updateBuilder.New(),
	}
}

func NewHelperFactory[TModel any](model *TModel, columns *types.ColumnsStorage, apiConnFactory *APIConnectorFactory) *HelperFactory[TModel] {
	return &HelperFactory[TModel]{
		model:         model,
		api:           apiConnFactory,
		whereBuilder:  NewWhereBuilderFabric(model, columns),
		orderBuilder:  NewOrderBuilderFabric(model, columns),
		selectBuilder: NewSelectBuilderFactory(model, columns),
		insertBuilder: NewInsertBuilderFactory(model, columns),
		updateBuilder: NewUpdateBuilderFactory(model, columns),
	}
}
