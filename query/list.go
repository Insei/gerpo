package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// GetListHelper is a generic interface for building complex queries to retrieve lists of data models.
type GetListHelper[TModel any] interface {
	// Exclude removes specified fields from requesting data from repository storage.
	Exclude(fieldsPtr ...any)
	// Only includes the specified columns in the execution context, ignoring all others in the existing collection.
	Only(fieldsPtr ...any)
	// Where defines the starting point for building conditions in a query, returning a types.WhereTarget interface.
	Where() types.WhereTarget
	// OrderBy defines the sorting criteria for a query and returns types.OrderTarget interface for further specification.
	OrderBy() types.OrderTarget

	// Page sets the page number for pagination in a query and returns the same GetListHelper instance.
	Page(page uint64) GetListHelper[TModel]

	// Size sets the maximum number of items to retrieve per page and returns the same GetListHelper instance.
	Size(size uint64) GetListHelper[TModel]
}

type GetListApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
	Where() sqlpart.Where
	Order() sqlpart.Order
	LimitOffset() sqlpart.LimitOffset
}

type GetList[TModel any] struct {
	baseModel *TModel

	whereBuilder      *linq.WhereBuilder
	orderBuilder      *linq.OrderBuilder
	excludeBuilder    *linq.ExcludeBuilder
	paginationBuilder *linq.PaginationBuilder
}

func (h *GetList[TModel]) Exclude(fieldPointers ...any) {
	h.excludeBuilder.Exclude(fieldPointers...)
}

func (h *GetList[TModel]) Only(fieldPointers ...any) {
	h.excludeBuilder.Only(fieldPointers...)
}

func (h *GetList[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *GetList[TModel]) OrderBy() types.OrderTarget {
	return h.orderBuilder
}

func (h *GetList[TModel]) Page(page uint64) GetListHelper[TModel] {
	h.paginationBuilder.Page(page)
	return h
}

func (h *GetList[TModel]) Size(size uint64) GetListHelper[TModel] {
	h.paginationBuilder.Size(size)
	return h
}

func (h *GetList[TModel]) Apply(applier GetListApplier) {
	h.excludeBuilder.Apply(applier)
	h.whereBuilder.Apply(applier)
	h.orderBuilder.Apply(applier)
	h.paginationBuilder.Apply(applier)
}

func (h *GetList[TModel]) HandleFn(qFns ...func(m *TModel, h GetListHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func NewGetList[TModel any](baseModel *TModel) *GetList[TModel] {
	return &GetList[TModel]{
		baseModel: baseModel,

		whereBuilder:      linq.NewWhereBuilder(baseModel),
		excludeBuilder:    linq.NewExcludeBuilder(baseModel),
		orderBuilder:      linq.NewOrderBuilder(baseModel),
		paginationBuilder: linq.NewPaginationBuilder(),
	}
}
