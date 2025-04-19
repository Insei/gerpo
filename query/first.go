package query

import (
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// GetFirstHelper is an interface for building query conditions to retrieve the first record matching specified criteria.
// It allows specifying WHERE conditions, excluding specific fields, and defining order-by operations for the query.
type GetFirstHelper[TModel any] interface {

	// Where defines the starting point for building conditions in a query, returning a types.WhereTarget interface.
	Where() types.WhereTarget

	// Exclude removes specified fields from requesting data from repository storage.
	Exclude(fieldsPtr ...any)

	// Only includes the specified columns in the execution context, ignoring all others in the existing collection.
	Only(fieldsPtr ...any)

	// OrderBy defines the sorting criteria for a query and returns types.OrderTarget interface for further specification.
	OrderBy() types.OrderTarget
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
