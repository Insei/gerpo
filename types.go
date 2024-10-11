package gerpo

import (
	"context"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/query"
)

type Repository[TModel any] interface {
	GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstUserHelper[TModel])) (model *TModel, err error)
	GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListUserHelper[TModel])) (models []*TModel, err error)
	Count(ctx context.Context, qFns ...func(m *TModel, h query.CountUserHelper[TModel])) (count uint64, err error)
	Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertUserHelper[TModel])) (err error)
	Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateUserHelper[TModel])) (err error)
	Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) (count uint64, err error)
}

func getModelAndFields[TModel any]() (*TModel, fmap.Storage, error) {
	model := new(TModel)
	mustZero(model)
	fields, err := fmap.GetFrom(model)
	if err != nil {
		return nil, nil, err
	}
	return model, fields, nil
}
