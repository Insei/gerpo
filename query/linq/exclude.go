package linq

import (
	"github.com/insei/gerpo/types"
)

type ColumnsExcluder interface {
	Exclude(columns ...types.Column)
}

type UserExcludeBuilder interface {
	Exclude(fieldPtrs ...any)
}

type ExcludeBuilder struct {
	*CoreBuilder
	opts []func(e ColumnsExcluder)
}

func NewExcludeBuilder(core *CoreBuilder) *ExcludeBuilder {
	return &ExcludeBuilder{
		CoreBuilder: core,
	}
}

func (b *ExcludeBuilder) Exclude(fieldPtrs ...any) {
	excludedCols := make([]types.Column, 0, len(fieldPtrs))
	for _, fieldPtr := range fieldPtrs {
		col := b.GetColumn(fieldPtr)
		excludedCols = append(excludedCols, col)
		b.opts = append(b.opts, func(e ColumnsExcluder) {
			e.Exclude(col)
		})
	}
}

func (b *ExcludeBuilder) Apply(columnsExcluder ColumnsExcluder) {
	for _, opt := range b.opts {
		opt(columnsExcluder)
	}
}
