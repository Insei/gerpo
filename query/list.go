package query

import (
	"github.com/insei/gerpo/query/linq"
)

type GetListUserHelper[TModel any] interface {
	GetFirstUserHelper[TModel]
	Page(page uint64) GetListUserHelper[TModel]
	Size(size uint64) GetListUserHelper[TModel]
}

type GetListHelper[TModel any] interface {
	GetListUserHelper[TModel]
	SQLApply
	HandleFn(qFns ...func(m *TModel, h GetListUserHelper[TModel]))
}

type getListHelper[TModel any] struct {
	*getFirstHelper[TModel]
}

func (h *getListHelper[TModel]) Page(page uint64) GetListUserHelper[TModel] {
	h.paginationBuilder.Page(page)
	return h
}

func (h *getListHelper[TModel]) Size(size uint64) GetListUserHelper[TModel] {
	h.paginationBuilder.Size(size)
	return h
}

func (h *getListHelper[TModel]) HandleFn(qFns ...func(m *TModel, h GetListUserHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.core.Model().(*TModel), h)
	}
}

func NewGetListHelper[TModel any](core *linq.CoreBuilder) GetListHelper[TModel] {
	getFirstH := newGetFirstHelper[TModel](core)
	getFirstH.paginationBuilder.Size(0)
	getFirstH.paginationBuilder.Page(0)
	return &getListHelper[TModel]{
		getFirstH,
	}
}
