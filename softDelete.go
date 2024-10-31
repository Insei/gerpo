package gerpo

import (
	"context"
	"github.com/insei/gerpo/types"
)

type SoftDeleteGetValueFn func(ctx context.Context) any

type SoftDeleteBuilder[TModel any] struct {
	model        *TModel
	storage      *types.ColumnsStorage
	sdColumnsMap map[types.Column]SoftDeleteGetValueFn
}

type SoftDeleteColumnBuilder struct {
	col     types.Column
	storage map[types.Column]SoftDeleteGetValueFn
}

func newSoftDeleteColumnBuilder(column types.Column, storage map[types.Column]SoftDeleteGetValueFn) *SoftDeleteColumnBuilder {
	return &SoftDeleteColumnBuilder{
		col:     column,
		storage: storage,
	}
}

func (b *SoftDeleteColumnBuilder) WithValueFunc(fn SoftDeleteGetValueFn) {
	b.storage[b.col] = fn
}

func newSoftDeleteBuilder[TModel any](model *TModel, storage *types.ColumnsStorage) *SoftDeleteBuilder[TModel] {
	return &SoftDeleteBuilder[TModel]{
		model:   model,
		storage: storage,
	}
}

func (b *SoftDeleteBuilder[TModel]) Column(fieldPtr any) *SoftDeleteColumnBuilder {
	col, err := b.storage.GetByFieldPtr(b.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	if b.sdColumnsMap == nil {
		b.sdColumnsMap = map[types.Column]SoftDeleteGetValueFn{}
	}

	return newSoftDeleteColumnBuilder(col, b.sdColumnsMap)
}

func (b *SoftDeleteBuilder[TModel]) build() map[types.Column]SoftDeleteGetValueFn {
	return b.sdColumnsMap
}
