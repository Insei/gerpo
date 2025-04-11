package linq

import (
	"github.com/insei/gerpo/types"
)

type ExcludeApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
}

type ExcludeBuilder struct {
	baseModel any
	opts      []func(applier ExcludeApplier)
}

func NewExcludeBuilder(baseModel any) *ExcludeBuilder {
	return &ExcludeBuilder{
		baseModel: baseModel,
	}
}

func (b *ExcludeBuilder) Exclude(fieldPtrs ...any) {
	for _, fieldPtr := range fieldPtrs {
		fieldSavePtr := fieldPtr
		b.opts = append(b.opts, func(applier ExcludeApplier) {
			col, err := applier.ColumnsStorage().GetByFieldPtr(b.baseModel, fieldSavePtr)
			if err != nil {
				panic(err)
			}
			applier.Columns().Exclude(col)
		})
	}
}

func (b *ExcludeBuilder) Only(fieldPtrs ...any) {
	b.opts = append(b.opts, func(applier ExcludeApplier) {
		userSpecifiedCols := make([]types.Column, 0, len(fieldPtrs))
		for _, fieldPtr := range fieldPtrs {
			col, err := applier.ColumnsStorage().GetByFieldPtr(b.baseModel, fieldPtr)
			if err != nil {
				panic(err)
			}
			userSpecifiedCols = append(userSpecifiedCols, col)
		}
		executionAllowedCols := applier.Columns().GetAll()
		// check that specified user fields can be used in current operation
		for _, userCol := range userSpecifiedCols {
			allowed := false
			for _, execCol := range executionAllowedCols {
				if userCol == execCol {
					allowed = true
					break
				}
			}
			if !allowed {
				panic("only: specified field is not allowed in current operation")
			}
		}
		applier.Columns().Only(userSpecifiedCols...)
	})
}

func (b *ExcludeBuilder) Apply(applier ExcludeApplier) {
	for _, opt := range b.opts {
		opt(applier)
	}
}
