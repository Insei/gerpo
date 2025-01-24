package linq

import (
	"github.com/insei/gerpo/types"
)

type ExcludeApplier interface {
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
		b.opts = append(b.opts, func(applier ExcludeApplier) {
			col := applier.Columns().GetByFieldPtr(b.baseModel, fieldPtr)
			applier.Columns().Exclude(col)
		})
	}
}

func (b *ExcludeBuilder) Apply(applier ExcludeApplier) {
	for _, opt := range b.opts {
		opt(applier)
	}
}
