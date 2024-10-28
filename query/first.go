package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/types"
)

type GetFirstUserHelper[TModel any] interface {
	Where() types.WhereTarget
	Exclude(fieldsPtr ...any)
	OrderBy() types.OrderTarget
}

type GetFirstHelper[TModel any] interface {
	GetFirstUserHelper[TModel]
	SQLApply
	HandleFn(qFns ...func(m *TModel, h GetFirstUserHelper[TModel]))
}

type getFirstHelper[TModel any] struct {
	*countHelper[TModel]
}

func (h *getFirstHelper[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *getFirstHelper[TModel]) OrderBy() types.OrderTarget {
	return h.orderBuilder
}

func (h *getFirstHelper[TModel]) HandleFn(qFns ...func(m *TModel, h GetFirstUserHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.core.Model().(*TModel), h)
	}
}

func newGetFirstHelper[TModel any](core *linq.CoreBuilder) *getFirstHelper[TModel] {
	countH := newCountHelper[TModel](core)
	countH.paginationBuilder.Size(1)
	countH.paginationBuilder.Page(1)
	return &getFirstHelper[TModel]{
		countH,
	}
}

func NewGetFirstHelper[TModel any](core *linq.CoreBuilder) GetFirstHelper[TModel] {
	return newGetFirstHelper[TModel](core)
}
