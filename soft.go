package gerpo

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/sqlstmt"
	"github.com/insei/gerpo/types"
)

type SoftDeletionBuilder[TModel any] struct {
	storage types.ColumnsStorage
	model   *TModel
	columns []types.Column
	fns     map[types.Column]func(model any, ctx context.Context)
}

type SoftDeletionValueSetter interface {
	SetValueFn(fn func(ctx context.Context) any)
}
type SoftDeletionValueFn func(fn func(ctx context.Context) any)

func (f SoftDeletionValueFn) SetValueFn(fn func(ctx context.Context) any) {
	f(fn)
}

func (b *SoftDeletionBuilder[TModel]) Field(fieldPtr any) SoftDeletionValueSetter {
	column, err := b.storage.GetByFieldPtr(b.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	if !column.IsAllowedAction(types.SQLActionUpdate) {
		panic("soft deletion can be used only on fields with allowed update action")
	}
	field := column.GetField()
	return SoftDeletionValueFn(func(fn func(ctx context.Context) any) {
		b.columns = append(b.columns, column)
		b.fns[column] = func(model any, ctx context.Context) {
			field.Set(model, fn(ctx))
		}
	})
}

func (b *SoftDeletionBuilder[TModel]) apply(repo *repository[TModel]) {
	softDeleteFn := func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteHelper[TModel])) (count int64, err error) {
		stmt := sqlstmt.NewUpdate(ctx, repo.columns, repo.table)
		// exclude all columns except soft deletion columns
		columns := repo.columns.AsSlice()
	COLUMNS:
		for _, column := range columns {
			for _, softDeleteColumn := range b.columns {
				if softDeleteColumn == column {
					continue COLUMNS
				}
			}
			stmt.Columns().Exclude(column)
		}

		// Apply persistent query
		repo.persistentQuery.Apply(stmt)

		// create new update query and apply delete query functions
		q := query.NewUpdate(repo.baseModel)
		for _, qFn := range qFns {
			qFn(repo.baseModel, q)
		}
		q.Apply(stmt)

		// Create new model and set soft deletion fields
		model := new(TModel)
		for _, column := range b.columns {
			b.fns[column](model, ctx)
		}

		// update model in repository
		updatedCount, err := repo.executor.Update(ctx, stmt, model)
		if err != nil {
			return updatedCount, repo.errorTransformer(err)
		}

		if updatedCount < 1 {
			return updatedCount, repo.errorTransformer(fmt.Errorf("nothing to delete: %w", ErrNotFound))
		}
		return updatedCount, nil
	}
	repo.deleteFn = softDeleteFn
}

// WithSoftDeletion configures soft deletion functionality for a model via a provided SoftDeletionBuilder instance.
func WithSoftDeletion[TModel any](fn func(m *TModel, builder *SoftDeletionBuilder[TModel])) Option[TModel] {
	return optionFn[TModel](func(r *repository[TModel]) {
		b := &SoftDeletionBuilder[TModel]{
			storage: r.columns,
			model:   r.baseModel,
			fns:     make(map[types.Column]func(model any, ctx context.Context)),
		}
		fn(b.model, b)
		b.apply(r)
	})
}
