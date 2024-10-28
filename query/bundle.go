package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type Bundle[TModel any] struct {
	core       *linq.CoreBuilder
	persistent PersistentHelper[TModel]
}

func NewBundle[TModel any](model *TModel, columns *types.ColumnsStorage) *Bundle[TModel] {
	core := linq.NewCoreBuilder(model, columns)
	return &Bundle[TModel]{
		core:       core,
		persistent: NewPersistentHelper[TModel](core),
	}
}

type handler[TModel, TUserHelper any] interface {
	HandleFn(...func(m *TModel, h TUserHelper))
	apply
}

type apply interface {
	Apply(*sql.StringBuilder)
}

func applyHelper[TModel, TUserHelper any](persistent apply, h handler[TModel, TUserHelper], sqlBuilder *sql.StringBuilder, qFns ...func(m *TModel, h TUserHelper)) {
	h.HandleFn(qFns...)
	h.Apply(sqlBuilder)
	persistent.Apply(sqlBuilder)
}

func (f *Bundle[TModel]) Persistent(qFns ...func(m *TModel, h PersistentUserHelper[TModel])) {
	for _, qFn := range qFns {
		f.persistent.HandleFn(qFn)
	}
}

func (f *Bundle[TModel]) ApplyCount(sqlBuilder *sql.StringBuilder, qFns ...func(m *TModel, h CountUserHelper[TModel])) {
	helper := NewCountHelper[TModel](f.core)
	applyHelper(f.persistent, helper, sqlBuilder, qFns...)
}

func (f *Bundle[TModel]) ApplyGetFirst(sqlBuilder *sql.StringBuilder, qFns ...func(m *TModel, h GetFirstUserHelper[TModel])) {
	helper := NewGetFirstHelper[TModel](f.core)
	applyHelper(f.persistent, helper, sqlBuilder, qFns...)
}

func (f *Bundle[TModel]) ApplyGetList(sqlBuilder *sql.StringBuilder, qFns ...func(m *TModel, h GetListUserHelper[TModel])) {
	helper := NewGetListHelper[TModel](f.core)
	applyHelper(f.persistent, helper, sqlBuilder, qFns...)
}

func (f *Bundle[TModel]) ApplyInsert(sqlBuilder *sql.StringBuilder, qFns ...func(m *TModel, h InsertUserHelper[TModel])) {
	helper := NewInsertHelper[TModel](f.core)
	applyHelper(f.persistent, helper, sqlBuilder, qFns...)
}

func (f *Bundle[TModel]) ApplyUpdate(sqlBuilder *sql.StringBuilder, qFns ...func(m *TModel, h UpdateUserHelper[TModel])) {
	helper := NewUpdateHelper[TModel](f.core)
	applyHelper(f.persistent, helper, sqlBuilder, qFns...)
}

func (f *Bundle[TModel]) ApplyDelete(sqlBuilder *sql.StringBuilder, qFns ...func(m *TModel, h DeleteUserHelper[TModel])) {
	helper := NewDeleteHelper[TModel](f.core)
	applyHelper(f.persistent, helper, sqlBuilder, qFns...)
}
