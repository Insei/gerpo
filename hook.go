package gerpo

import (
	"context"
	"github.com/insei/gerpo/query"
)

func useDeleteHook[TModel any](repo *repository[TModel]) {
	repo.getDeleteModels = func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteUserHelper[TModel])) ([]*TModel, error) {
		strSQLBuilder := repo.strSQLBuilderFactory.New(ctx)
		repo.query.ApplyDelete(strSQLBuilder, qFns...)

		return repo.executor.GetMultiple(ctx, strSQLBuilder)
	}
}
