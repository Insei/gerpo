package query

import "github.com/insei/gerpo/types"

type Helper[TModel any] struct {
	api           *APIConnector
	whereBuilder  *WhereBuilder[TModel]
	orderBuilder  *OrderBuilder[TModel]
	selectBuilder *SelectBuilder[TModel]
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

func NewHelper[TModel any](selectBuilder *SelectBuilder[TModel], whereBuilder *WhereBuilder[TModel], orderBuilder *OrderBuilder[TModel], connector *APIConnector) *Helper[TModel] {
	return &Helper[TModel]{
		api:           connector,
		whereBuilder:  whereBuilder,
		orderBuilder:  orderBuilder,
		selectBuilder: selectBuilder,
	}
}
