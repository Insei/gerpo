package gerpo

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/sqlstmt"
	"github.com/insei/gerpo/types"
)

type repository[TModel any] struct {
	// Callbacks and Hooks
	beforeInsert     func(ctx context.Context, model *TModel)
	beforeUpdate     func(ctx context.Context, model *TModel)
	afterInsert      func(ctx context.Context, model *TModel)
	afterUpdate      func(ctx context.Context, model *TModel)
	afterSelect      func(ctx context.Context, models []*TModel)
	errorTransformer func(err error) error

	// Tracing — opt-in via WithTracer; nil means "no spans".
	tracer Tracer

	// Columns and fields
	baseModel *TModel
	table     string
	columns   types.ColumnsStorage

	// SQL Query, execution and dependency
	executor        executor.Executor[TModel]
	persistentQuery *query.Persistent[TModel]

	deleteFn func(ctx context.Context, qFns ...func(m *TModel, h query.DeleteHelper[TModel])) (count int64, err error)
}

// startSpan opens a tracing span around a repository operation. When no Tracer
// is configured, returns the original context and a no-op end function so the
// callers stay branch-free.
func (r *repository[TModel]) startSpan(ctx context.Context, op string) (context.Context, SpanEnd) {
	if r.tracer == nil {
		return ctx, noopSpanEnd
	}
	return r.tracer(ctx, SpanInfo{Op: op, Table: r.table})
}

func replaceNilCallbacks[TModel any](repo *repository[TModel]) {
	if repo.beforeInsert == nil {
		repo.beforeInsert = func(_ context.Context, _ *TModel) {}
	}
	if repo.beforeUpdate == nil {
		repo.beforeUpdate = func(_ context.Context, _ *TModel) {}
	}
	if repo.afterInsert == nil {
		repo.afterInsert = func(_ context.Context, _ *TModel) {}
	}
	if repo.afterUpdate == nil {
		repo.afterUpdate = func(_ context.Context, _ *TModel) {}
	}
	if repo.afterSelect == nil {
		repo.afterSelect = func(_ context.Context, _ []*TModel) {}
	}
	if repo.errorTransformer == nil {
		repo.errorTransformer = func(err error) error { return err }
	}
}

// newRepository is the low-level constructor used by the fluent builder
// (gerpo.New[TModel]().…Build()). It assumes a fully prepared executor and a
// columnsFn — kept unexported so that external callers always go through the
// builder, which validates db / table / columns and applies executor options.
func newRepository[TModel any](exec executor.Executor[TModel], table string, columnsFn func(m *TModel, builder *ColumnBuilder[TModel]), opts ...Option[TModel]) (Repository[TModel], error) {
	model, fields, err := getModelAndFields[TModel]()
	if err != nil {
		return nil, err
	}
	columnsBuilder := newColumnBuilder(table, model, fields)
	columnsFn(model, columnsBuilder)
	columns, err := columnsBuilder.build()
	if err != nil {
		return nil, err
	}

	if len(columns.AsSlice()) < 1 {
		return nil, fmt.Errorf("failed to create repository with empty columns")
	}
	repo := &repository[TModel]{
		columns:         columns,
		executor:        exec,
		table:           table,
		baseModel:       model,
		persistentQuery: query.NewPersistent(model),
	}
	repo.deleteFn = repo.delete

	for _, opt := range opts {
		err := opt.apply(repo)
		if err != nil {
			return nil, err
		}
	}

	replaceNilCallbacks(repo)
	return repo, nil
}

func (r *repository[TModel]) GetColumns() types.ColumnsStorage {
	return r.columns
}

func (r *repository[TModel]) GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstHelper[TModel])) (model *TModel, err error) {
	ctx, end := r.startSpan(ctx, "gerpo.GetFirst")
	defer func() { end(err) }()

	stmt := sqlstmt.NewGetFirst(ctx, r.table, r.columns)
	defer stmt.Release()
	err = r.persistentQuery.Apply(stmt)
	if err != nil {
		return nil, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyPersistentQuery, err))
	}

	q := query.NewGetFirst(r.baseModel)
	q.HandleFn(qFns...)
	err = q.Apply(stmt)
	if err != nil {
		return nil, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyQuery, err))
	}

	model, err = r.executor.GetOne(ctx, stmt)
	if err != nil {
		return nil, r.errorTransformer(err)
	}

	r.afterSelect(ctx, []*TModel{model})
	return model, nil
}

func (r *repository[TModel]) GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListHelper[TModel])) (models []*TModel, err error) {
	ctx, end := r.startSpan(ctx, "gerpo.GetList")
	defer func() { end(err) }()

	stmt := sqlstmt.NewGetList(ctx, r.table, r.columns)
	defer stmt.Release()
	err = r.persistentQuery.Apply(stmt)
	if err != nil {
		return nil, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyPersistentQuery, err))
	}

	q := query.NewGetList(r.baseModel)
	q.HandleFn(qFns...)
	err = q.Apply(stmt)
	if err != nil {
		return nil, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyQuery, err))
	}

	models, err = r.executor.GetMultiple(ctx, stmt)
	if err != nil {
		return nil, r.errorTransformer(err)
	}

	r.afterSelect(ctx, models)
	return models, nil
}

func (r *repository[TModel]) Count(ctx context.Context, qFns ...func(m *TModel, h query.CountHelper[TModel])) (count uint64, err error) {
	ctx, end := r.startSpan(ctx, "gerpo.Count")
	defer func() { end(err) }()

	stmt := sqlstmt.NewCount(ctx, r.table, r.columns)
	defer stmt.Release()
	err = r.persistentQuery.Apply(stmt)
	if err != nil {
		return 0, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyPersistentQuery, err))
	}

	q := query.NewCount(r.baseModel)
	q.HandleFn(qFns...)
	err = q.Apply(stmt)
	if err != nil {
		return 0, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyQuery, err))
	}

	count, err = r.executor.Count(ctx, stmt)
	if err != nil {
		return 0, r.errorTransformer(err)
	}

	return count, nil
}

func (r *repository[TModel]) Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertHelper[TModel])) (err error) {
	ctx, end := r.startSpan(ctx, "gerpo.Insert")
	defer func() { end(err) }()

	r.beforeInsert(ctx, model)
	stmt := sqlstmt.NewInsert(ctx, r.table, r.columns)
	err = r.persistentQuery.Apply(stmt)
	if err != nil {
		return r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyPersistentQuery, err))
	}

	q := query.NewInsert(r.baseModel)
	q.HandleFn(qFns...)
	err = q.Apply(stmt)
	if err != nil {
		return r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyQuery, err))
	}

	err = r.executor.InsertOne(ctx, stmt, model)
	if err != nil {
		return r.errorTransformer(err)
	}

	r.afterInsert(ctx, model)
	return nil
}

func (r *repository[TModel]) Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateHelper[TModel])) (count int64, err error) {
	ctx, end := r.startSpan(ctx, "gerpo.Update")
	defer func() { end(err) }()

	r.beforeUpdate(ctx, model)
	stmt := sqlstmt.NewUpdate(ctx, r.columns, r.table)
	err = r.persistentQuery.Apply(stmt)
	if err != nil {
		return 0, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyPersistentQuery, err))
	}

	q := query.NewUpdate(r.baseModel)
	q.HandleFn(qFns...)
	err = q.Apply(stmt)
	if err != nil {
		return 0, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyQuery, err))
	}

	updatedCount, err := r.executor.Update(ctx, stmt, model)
	if err != nil {
		return updatedCount, r.errorTransformer(err)
	}

	if updatedCount < 1 {
		return updatedCount, r.errorTransformer(fmt.Errorf("nothing to update: %w", ErrNotFound))
	}
	r.afterUpdate(ctx, model)
	return updatedCount, nil
}

func (r *repository[TModel]) delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteHelper[TModel])) (count int64, err error) {
	stmt := sqlstmt.NewDelete(ctx, r.table, r.columns)
	err = r.persistentQuery.Apply(stmt)
	if err != nil {
		return 0, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyPersistentQuery, err))
	}

	q := query.NewDelete(r.baseModel)
	q.HandleFn(qFns...)
	err = q.Apply(stmt)
	if err != nil {
		return 0, r.errorTransformer(fmt.Errorf("%w: %w", ErrApplyQuery, err))
	}

	count, err = r.executor.Delete(ctx, stmt)
	if err != nil {
		return count, r.errorTransformer(err)
	}

	if count < 1 {
		return 0, r.errorTransformer(fmt.Errorf("nothing to delete: %w", ErrNotFound))
	}
	return count, r.errorTransformer(err)
}

func (r *repository[TModel]) Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteHelper[TModel])) (count int64, err error) {
	ctx, end := r.startSpan(ctx, "gerpo.Delete")
	defer func() { end(err) }()
	count, err = r.deleteFn(ctx, qFns...)
	return count, err
}
