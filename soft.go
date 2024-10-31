package gerpo

import (
	"context"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
)

type softDeleteGetValueFn func(ctx context.Context) any
type deleteFn[TModel any] func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error)

type SoftDeleteBuilder[TModel any] struct {
	model        *TModel
	storage      *types.ColumnsStorage
	sdColumnsMap map[types.Column]softDeleteGetValueFn
}

type SoftDeleteColumnBuilder struct {
	col     types.Column
	storage map[types.Column]softDeleteGetValueFn
}

func (b *SoftDeleteColumnBuilder) WithValueFunc(fn softDeleteGetValueFn) {
	b.storage[b.col] = fn
}

func (b *SoftDeleteBuilder[TModel]) Column(fieldPtr any) *SoftDeleteColumnBuilder {
	col, err := b.storage.GetByFieldPtr(b.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	if b.sdColumnsMap == nil {
		b.sdColumnsMap = map[types.Column]softDeleteGetValueFn{}
	}

	return newSoftDeleteColumnBuilder(col, b.sdColumnsMap)
}

func newSoftDeleteColumnBuilder(column types.Column, storage map[types.Column]softDeleteGetValueFn) *SoftDeleteColumnBuilder {
	return &SoftDeleteColumnBuilder{
		col:     column,
		storage: storage,
	}
}

func newSoftDeleteBuilder[TModel any](model *TModel, storage *types.ColumnsStorage) *SoftDeleteBuilder[TModel] {
	return &SoftDeleteBuilder[TModel]{
		model:   model,
		storage: storage,
	}
}

func (b *SoftDeleteBuilder[TModel]) build() map[types.Column]softDeleteGetValueFn {
	return b.sdColumnsMap
}

func getDeleteFn[TModel any](r *repository[TModel], model *TModel, sdColumnsFn func(m *TModel, builder *SoftDeleteBuilder[TModel])) deleteFn[TModel] {
	var sdColumnsMap map[types.Column]softDeleteGetValueFn
	if sdColumnsFn != nil {
		sdColumnsBuilder := newSoftDeleteBuilder(model, r.columns)
		sdColumnsFn(model, sdColumnsBuilder)
		sdColumnsMap = sdColumnsBuilder.build()
	}

	if sdColumnsMap != nil && len(sdColumnsMap) > 0 {
		return func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error) {
			strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
			m := new(TModel)

			for col, getValFn := range sdColumnsMap {
				col.GetField().Set(m, getValFn(ctx))
			}

			r.query.ApplyUpdate(strSQLBuilder, func(m *TModel, h query.UpdateUserHelper[TModel]) {
				excludePointers := getExcludePointers(m, r.columns.AsSlice(), sdColumnsMap)
				h.Exclude(excludePointers...)
			})
			// Apply WHERE filters
			r.query.ApplyDelete(strSQLBuilder, qFns...)

			return r.executor.Update(ctx, m, strSQLBuilder)
		}
	}

	return func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count int64, err error) {
		strSQLBuilder := r.strSQLBuilderFactory.New(ctx)
		r.query.ApplyDelete(strSQLBuilder, qFns...)
		return r.executor.Delete(ctx, strSQLBuilder)
	}
}

func getExcludePointers[TModel any](m *TModel, allCols []types.Column, sdCols map[types.Column]softDeleteGetValueFn) []any {
	pointers := make([]any, 0, len(allCols)-len(sdCols))
	for _, col := range allCols {
		if _, ok := sdCols[col]; !ok {
			pointers = append(pointers, col.GetField().GetPtr(m))
		}
	}

	return pointers
}
