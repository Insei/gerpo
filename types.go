package gerpo

import (
	"context"
	"errors"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
)

var ErrNotFound = errors.New("not found")

type Repository[TModel any] interface {
	GetColumns() types.ColumnsStorage
	Tx(tx *executor.Tx) (Repository[TModel], error)
	GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstHelper[TModel])) (model *TModel, err error)
	GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListHelper[TModel])) (models []*TModel, err error)
	Count(ctx context.Context, qFns ...func(m *TModel, h query.CountHelper[TModel])) (count uint64, err error)
	Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertHelper[TModel])) (err error)
	Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateHelper[TModel])) (err error)
	Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteHelper[TModel])) (count int64, err error)
}

type Builder[TModel any] interface {
	WithQuery(queryFn func(m *TModel, h query.PersistentHelper[TModel])) Builder[TModel]
	BeforeInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	BeforeUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	AfterSelect(fn func(ctx context.Context, models []*TModel)) Builder[TModel]
	AfterInsert(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	AfterUpdate(fn func(ctx context.Context, m *TModel)) Builder[TModel]
	WithErrorTransformer(fn func(err error) error) Builder[TModel]
	Build() (Repository[TModel], error)
}
