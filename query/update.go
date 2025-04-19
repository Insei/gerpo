package query

import (
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// UpdateHelper is a generic interface for managing update operations on models in repository storage.
// It allows specifying conditions and excluding specific fields from the update process.
type UpdateHelper[TModel any] interface {
	// Exclude removes specified fields from requesting data from repository storage.
	Exclude(fieldsPtr ...any)
	// Only includes the specified columns in the execution context, ignoring all others in the existing collection.
	Only(fieldsPtr ...any)
	// Where defines the starting point for building conditions in a query, returning a types.WhereTarget interface.
	Where() types.WhereTarget
}

type UpdateApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
	Where() sqlpart.Where
}
type Update[TModel any] struct {
	baseModel *TModel

	excludeBuilder *linq.ExcludeBuilder
	whereBuilder   *linq.WhereBuilder
}

func (h *Update[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *Update[TModel]) Only(fieldPointers ...any) {
	h.excludeBuilder.Only(fieldPointers...)
}

func (h *Update[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *Update[TModel]) Apply(applier UpdateApplier) error {
	err := h.excludeBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyExcludeColumnRules, err)
	}
	err = h.whereBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyWhereClause, err)
	}
	return nil
}

func (h *Update[TModel]) HandleFn(qFns ...func(m *TModel, h UpdateHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func NewUpdate[TModel any](baseModel *TModel) *Update[TModel] {
	return &Update[TModel]{
		baseModel: baseModel,

		excludeBuilder: linq.NewExcludeBuilder(baseModel),
		whereBuilder:   linq.NewWhereBuilder(baseModel),
	}
}
