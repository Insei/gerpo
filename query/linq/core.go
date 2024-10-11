package linq

import "github.com/insei/gerpo/types"

type CoreBuilder struct {
	model   any
	columns *types.ColumnsStorage
}

func NewCoreBuilder(model any, columns *types.ColumnsStorage) *CoreBuilder {
	return &CoreBuilder{
		model:   model,
		columns: columns,
	}
}

func (b *CoreBuilder) GetColumn(fieldPtr any) types.Column {
	col, err := b.columns.GetByFieldPtr(b.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	return col
}

func (b *CoreBuilder) GetColumnsByAction(action types.SQLAction) []types.Column {
	return b.columns.AsSliceByAction(action)
}

func (b *CoreBuilder) Model() any {
	return b.model
}
