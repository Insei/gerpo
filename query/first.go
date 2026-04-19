package query

import (
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// GetFirstHelper is the per-request helper for repo.GetFirst. It composes the
// small contracts from interfaces.go: filtering, sorting and narrowing the
// column set.
type GetFirstHelper[TModel any] interface {
	Filterable
	Sortable
	Excludable
}

// GetFirstApplier defines an interface for applying columns, filters, and ordering in a query construction process.
// It provides access to the columns storage, execution-related column operations, filtering conditions, and ordering.
type GetFirstApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
	Where() sqlpart.Where
	Order() sqlpart.Order
}

type GetFirst[TModel any] struct {
	baseModel *TModel

	whereBuilder   *linq.WhereBuilder
	orderBuilder   *linq.OrderBuilder
	excludeBuilder *linq.ExcludeBuilder
}

func (h *GetFirst[TModel]) Exclude(fieldPointers ...any) {
	h.excludeBuilder.Exclude(fieldPointers...)
}

func (h *GetFirst[TModel]) Only(fieldPointers ...any) {
	h.excludeBuilder.Only(fieldPointers...)
}

func (h *GetFirst[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *GetFirst[TModel]) OrderBy() types.OrderTarget {
	return h.orderBuilder
}

func (h *GetFirst[TModel]) HandleFn(qFns ...func(m *TModel, h GetFirstHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func (h *GetFirst[TModel]) Apply(applier GetFirstApplier) error {
	err := h.excludeBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyExcludeColumnRules, err)
	}
	err = h.orderBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyOrderByOperator, err)
	}
	err = h.whereBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyWhereClause, err)
	}
	return nil
}

func NewGetFirst[TModel any](baseModel *TModel) *GetFirst[TModel] {
	return &GetFirst[TModel]{
		baseModel: baseModel,

		whereBuilder:   linq.NewWhereBuilder(baseModel),
		excludeBuilder: linq.NewExcludeBuilder(baseModel),
		orderBuilder:   linq.NewOrderBuilder(baseModel),
	}
}
