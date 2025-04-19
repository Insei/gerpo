package linq

import (
	"fmt"

	"github.com/insei/gerpo/types"
)

type ExcludeApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
}

type ExcludeBuilder struct {
	baseModel any
	opts      []func(applier ExcludeApplier) error
}

func NewExcludeBuilder(baseModel any) *ExcludeBuilder {
	return &ExcludeBuilder{
		baseModel: baseModel,
	}
}

func (b *ExcludeBuilder) Exclude(fieldPtrs ...any) {
	for _, fieldPtr := range fieldPtrs {
		fieldSavePtr := fieldPtr
		b.opts = append(b.opts, func(applier ExcludeApplier) error {
			col, err := applier.ColumnsStorage().GetByFieldPtr(b.baseModel, fieldSavePtr)
			if err != nil {
				return err
			}
			applier.Columns().Exclude(col)
			return nil
		})
	}
}

func (b *ExcludeBuilder) Only(fieldPtrs ...any) {
	b.opts = append(b.opts, func(applier ExcludeApplier) error {
		userSpecifiedCols := make([]types.Column, 0, len(fieldPtrs))
		for _, fieldPtr := range fieldPtrs {
			col, err := applier.ColumnsStorage().GetByFieldPtr(b.baseModel, fieldPtr)
			if err != nil {
				return err
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
				return fmt.Errorf("only: specified field %s is not allowed in current operation", userCol.GetField().GetStructPath())
			}
		}
		applier.Columns().Only(userSpecifiedCols...)
		return nil
	})
}

func (b *ExcludeBuilder) Apply(applier ExcludeApplier) error {
	for _, opt := range b.opts {
		err := opt(applier)
		if err != nil {
			return err
		}
	}
	return nil
}
