package executor

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/sqlstmt"
)

type executor[TModel any] struct {
	db                   DBAdapter
	getExecQueryReplaced func(ctx context.Context) ExecQuery

	options
}

func New[TModel any](db DBAdapter, opts ...Option) Executor[TModel] {
	o := &options{}
	for _, opt := range opts {
		opt.apply(o)
	}
	e := &executor[TModel]{
		options: *o,
		db:      db,
	}
	return e
}

func (e *executor[TModel]) Tx(tx Tx) Executor[TModel] {
	ecp := *e
	ecp.getExecQueryReplaced = func(ctx context.Context) ExecQuery {
		return tx
	}
	return &ecp
}

func (e *executor[TModel]) getExecQuery(ctx context.Context) ExecQuery {
	if e.getExecQueryReplaced != nil {
		return e.getExecQueryReplaced(ctx)
	}
	return e.db
}

// startSpan opens a tracing span around an executor operation. When no Tracer
// is configured, returns the original context and a no-op end function so the
// callers can stay branch-free.
func (e *executor[TModel]) startSpan(ctx context.Context, op string) (context.Context, SpanEnd) {
	if e.tracer == nil {
		return ctx, noopSpanEnd
	}
	return e.tracer(ctx, op)
}

func noopSpanEnd(error) {}

func (e *executor[TModel]) GetOne(ctx context.Context, stmt Stmt) (model *TModel, err error) {
	ctx, end := e.startSpan(ctx, "gerpo.GetOne")
	defer func() { end(err) }()

	sql, args, err := stmt.SQL()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql query from stmt: %w", err)
	}
	if cached, ok := get[TModel](ctx, e.cacheSource, sql, args...); ok {
		return cached, nil
	}
	rows, err := e.getExecQuery(ctx).QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	if rows.Next() {
		model = new(TModel)
		pointers := stmt.Columns().GetModelPointers(model)
		if err = rows.Scan(pointers...); err != nil {
			return nil, err
		}
		set(ctx, e.cacheSource, *model, sql, args...)
	}
	if model == nil {
		return nil, ErrNoRows
	}
	return model, nil
}

func (e *executor[TModel]) GetMultiple(ctx context.Context, stmt Stmt) (models []*TModel, err error) {
	ctx, end := e.startSpan(ctx, "gerpo.GetMultiple")
	defer func() { end(err) }()

	sql, args, err := stmt.SQL()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql query from stmt: %w", err)
	}
	if cached, ok := get[[]*TModel](ctx, e.cacheSource, sql, args...); ok {
		return *cached, nil
	}
	rows, err := e.getExecQuery(ctx).QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	for rows.Next() {
		model := new(TModel)
		if err = rows.Scan(stmt.Columns().GetModelPointers(model)...); err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	set(ctx, e.cacheSource, models, sql, args...)
	return models, nil
}

func (e *executor[TModel]) InsertOne(ctx context.Context, stmt Stmt, model *TModel) (err error) {
	ctx, end := e.startSpan(ctx, "gerpo.InsertOne")
	defer func() { end(err) }()

	sql, values, err := stmt.SQL(sqlstmt.WithModelValues(model))
	if err != nil {
		return fmt.Errorf("failed to get sql query from stmt: %w", err)
	}
	result, err := e.getExecQuery(ctx).ExecContext(ctx, sql, values...)
	if err != nil {
		return err
	}
	insertedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if insertedRows == 0 {
		return ErrNoInsertedRows
	}
	clean(ctx, e.cacheSource)
	return nil
}

func (e *executor[TModel]) Update(ctx context.Context, stmt Stmt, model *TModel) (updatedRows int64, err error) {
	ctx, end := e.startSpan(ctx, "gerpo.Update")
	defer func() { end(err) }()

	sql, values, err := stmt.SQL(sqlstmt.WithModelValues(model))
	if err != nil {
		return 0, fmt.Errorf("failed to get sql query from stmt: %w", err)
	}
	result, err := e.getExecQuery(ctx).ExecContext(ctx, sql, values...)
	if err != nil {
		return 0, err
	}
	updatedRows, err = result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if updatedRows > 0 {
		clean(ctx, e.cacheSource)
	}
	return updatedRows, nil
}

func (e *executor[TModel]) Count(ctx context.Context, stmt CountStmt) (count uint64, err error) {
	ctx, end := e.startSpan(ctx, "gerpo.Count")
	defer func() { end(err) }()

	sql, args, err := stmt.SQL()
	if err != nil {
		return 0, fmt.Errorf("failed to get sql query from stmt: %w", err)
	}
	if cached, ok := get[uint64](ctx, e.cacheSource, sql, args...); ok {
		return *cached, nil
	}
	rows, err := e.getExecQuery(ctx).QueryContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close() //nolint:errcheck
	if rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	set(ctx, e.cacheSource, count, sql, args...)
	return count, nil
}

func (e *executor[TModel]) Delete(ctx context.Context, stmt CountStmt) (deletedRows int64, err error) {
	ctx, end := e.startSpan(ctx, "gerpo.Delete")
	defer func() { end(err) }()

	sql, args, err := stmt.SQL()
	if err != nil {
		return 0, fmt.Errorf("failed to get sql query from stmt: %w", err)
	}
	result, err := e.getExecQuery(ctx).ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	deletedRows, err = result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if deletedRows > 0 {
		clean(ctx, e.cacheSource)
	}
	return deletedRows, nil
}
