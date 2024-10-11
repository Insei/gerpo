package linq

import (
	"slices"

	"github.com/insei/gerpo/types"
)

type ColumnsReader interface {
	Columns(columns ...types.Column)
}

type Exclude interface {
	Exclude(columns ...types.Column)
}

type UserExcludeBuilder interface {
	Exclude(fieldPtrs ...any)
}

type ExcludeBuilder struct {
	*CoreBuilder
	columns []types.Column
	opts    []func(e Exclude)
}

func NewExcludeBuilder(core *CoreBuilder, action types.SQLAction) *ExcludeBuilder {
	return &ExcludeBuilder{
		CoreBuilder: core,
		columns:     core.GetColumnsByAction(action),
	}
}

func (b *ExcludeBuilder) Exclude(fieldPtrs ...any) {
	excludedCols := make([]types.Column, 0, len(fieldPtrs))
	for _, fieldPtr := range fieldPtrs {
		col := b.GetColumn(fieldPtr)
		excludedCols = append(excludedCols, col)
		b.opts = append(b.opts, func(e Exclude) {
			e.Exclude(col)
		})
	}
	b.columns = slices.DeleteFunc(b.columns, func(column types.Column) bool {
		if slices.Contains(excludedCols, column) {
			return true
		}
		return false
	})
}

func (b *ExcludeBuilder) GetColumns() []types.Column {
	return b.columns
}

func (b *ExcludeBuilder) Apply(columnsReader ColumnsReader) {
	columnsReader.Columns(b.columns...)
}
