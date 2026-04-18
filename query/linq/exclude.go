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
	baseModel     any
	excluded      []any
	onlyFieldPtrs []any
	hasOnly       bool
}

func NewExcludeBuilder(baseModel any) *ExcludeBuilder {
	return &ExcludeBuilder{
		baseModel: baseModel,
	}
}

func (b *ExcludeBuilder) Exclude(fieldPtrs ...any) {
	b.excluded = append(b.excluded, fieldPtrs...)
}

func (b *ExcludeBuilder) Only(fieldPtrs ...any) {
	b.hasOnly = true
	b.onlyFieldPtrs = append(b.onlyFieldPtrs[:0], fieldPtrs...)
}

func (b *ExcludeBuilder) Apply(applier ExcludeApplier) error {
	storage := applier.ColumnsStorage()
	cols := applier.Columns()
	for _, fieldPtr := range b.excluded {
		col, err := storage.GetByFieldPtr(b.baseModel, fieldPtr)
		if err != nil {
			return err
		}
		cols.Exclude(col)
	}
	if !b.hasOnly {
		return nil
	}
	userSpecifiedCols := make([]types.Column, 0, len(b.onlyFieldPtrs))
	for _, fieldPtr := range b.onlyFieldPtrs {
		col, err := storage.GetByFieldPtr(b.baseModel, fieldPtr)
		if err != nil {
			return err
		}
		userSpecifiedCols = append(userSpecifiedCols, col)
	}
	executionAllowedCols := cols.GetAll()
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
	cols.Only(userSpecifiedCols...)
	return nil
}
