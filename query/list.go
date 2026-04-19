package query

import (
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// GetListHelper is the per-request helper for repo.GetList. It composes
// the small contracts from interfaces.go: filtering, sorting, narrowing the
// column set, and pagination.
type GetListHelper[TModel any] interface {
	Filterable
	Sortable
	Excludable
	Pageable[TModel]
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

func (h *GetList[TModel]) Apply(applier GetListApplier) error {
	err := h.excludeBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyExcludeColumnRules, err)
	}
	err = h.whereBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyWhereClause, err)
	}
	err = h.orderBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyOrderByOperator, err)
	}
	err = h.paginationBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w:%w", ErrApplyLimitOffsetOperator, err)
	}
	return nil
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
