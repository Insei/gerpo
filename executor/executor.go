package executor

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/sqlstmt"
	"github.com/insei/gerpo/types"
)

type executor[TModel any] struct {
	db Adapter

	options
}

func New[TModel any](db Adapter, opts ...Option) Executor[TModel] {
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

// getExecQuery returns the ExecQuery to use for the current call. When ctx
// carries a Tx (installed via executor.WithTx / gerpo.WithTx), that Tx wins;
// otherwise the repository-level adapter is used.
func (e *executor[TModel]) getExecQuery(ctx context.Context) ExecQuery {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return e.db
}

func (e *executor[TModel]) GetOne(ctx context.Context, stmt Stmt) (model *TModel, err error) {
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
	sql, values, err := stmt.SQL(sqlstmt.WithModelValues(model))
	if err != nil {
		return fmt.Errorf("failed to get sql query from stmt: %w", err)
	}
	if returning := returningColumnsOf(stmt); len(returning) > 0 {
		rows, err := e.getExecQuery(ctx).QueryContext(ctx, sql, values...)
		if err != nil {
			return err
		}
		defer rows.Close() //nolint:errcheck
		if !rows.Next() {
			// PG drivers (lib/pq, pgx via database/sql) defer execution errors —
			// e.g. unique_violation on INSERT ... RETURNING — until iteration.
			// Without this check the real error is masked by ErrNoInsertedRows.
			if err := rows.Err(); err != nil {
				return err
			}
			return ErrNoInsertedRows
		}
		ptrs := scanPointers(returning, model)
		if err = rows.Scan(ptrs...); err != nil {
			return err
		}
		clean(ctx, e.cacheSource)
		return nil
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

// InsertMany emits a multi-row INSERT ... VALUES (...) [RETURNING ...] and
// transparently chunks the input to stay under the driver's placeholder limit
// (PostgreSQL caps bound parameters at 65535 per query). When RETURNING is
// configured on the stmt, scanned rows are written back into models[i] by
// position; every chunk consumes its slice range in order.
//
// Failures leave the work already committed by prior chunks in place: the
// caller is responsible for wrapping the call in gerpo.RunInTx when atomicity
// across chunks is required.
func (e *executor[TModel]) InsertMany(ctx context.Context, stmt BatchStmt, models []*TModel) (int64, error) {
	if len(models) == 0 {
		return 0, nil
	}
	cols := stmt.Columns().GetAll()
	activeCols := 0
	for _, c := range cols {
		if _, ok := c.Name(); ok {
			activeCols++
		}
	}
	if activeCols == 0 {
		return 0, fmt.Errorf("failed to get sql query from stmt: no columns to insert")
	}

	// PostgreSQL caps bound parameters at 65535. Stay well below that so other
	// call-site contributions (RETURNING, ON CONFLICT, …) cannot trip it.
	const placeholderBudget = 65000
	chunkSize := placeholderBudget / activeCols
	if chunkSize < 1 {
		chunkSize = 1
	}

	returning := returningColumnsOfBatch(stmt)
	chunkBuf := make([]any, 0, chunkSize)

	var total int64
	for start := 0; start < len(models); start += chunkSize {
		end := start + chunkSize
		if end > len(models) {
			end = len(models)
		}
		chunkBuf = chunkBuf[:0]
		for _, m := range models[start:end] {
			chunkBuf = append(chunkBuf, m)
		}
		stmt.SetModels(chunkBuf)

		sql, values, err := stmt.SQL()
		if err != nil {
			return total, fmt.Errorf("failed to get sql query from stmt: %w", err)
		}

		if len(returning) > 0 {
			rows, err := e.getExecQuery(ctx).QueryContext(ctx, sql, values...)
			if err != nil {
				return total, err
			}
			idx := start
			for rows.Next() {
				if idx >= end {
					_ = rows.Close()
					return total, fmt.Errorf("RETURNING yielded more rows than sent (chunk [%d..%d))", start, end)
				}
				ptrs := scanPointers(returning, models[idx])
				if err := rows.Scan(ptrs...); err != nil {
					_ = rows.Close()
					return total, err
				}
				idx++
				total++
			}
			// PG drivers surface execution errors (e.g. unique_violation) only
			// after iteration; without this check they are silently swallowed.
			if err := rows.Err(); err != nil {
				_ = rows.Close()
				return total, err
			}
			_ = rows.Close()
			continue
		}
		result, err := e.getExecQuery(ctx).ExecContext(ctx, sql, values...)
		if err != nil {
			return total, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return total, err
		}
		total += n
	}
	if total > 0 {
		clean(ctx, e.cacheSource)
	}
	return total, nil
}

// returningColumnsOfBatch mirrors returningColumnsOf for the BatchStmt shape;
// BatchStmt embeds Stmt and already carries the ReturningStmt capability via
// the concrete sqlstmt.InsertBatch.
func returningColumnsOfBatch(stmt BatchStmt) []types.Column {
	rs, ok := any(stmt).(ReturningStmt)
	if !ok {
		return nil
	}
	return rs.ReturningColumns()
}

func (e *executor[TModel]) Update(ctx context.Context, stmt Stmt, model *TModel) (updatedRows int64, err error) {
	sql, values, err := stmt.SQL(sqlstmt.WithModelValues(model))
	if err != nil {
		return 0, fmt.Errorf("failed to get sql query from stmt: %w", err)
	}
	if returning := returningColumnsOf(stmt); len(returning) > 0 {
		rows, err := e.getExecQuery(ctx).QueryContext(ctx, sql, values...)
		if err != nil {
			return 0, err
		}
		defer rows.Close() //nolint:errcheck
		// Update with RETURNING: PG returns one row per affected row. We expect
		// exactly one because the user-facing Update updates a single model.
		var n int64
		ptrs := scanPointers(returning, model)
		for rows.Next() {
			if err = rows.Scan(ptrs...); err != nil {
				return n, err
			}
			n++
		}
		// PG drivers defer execution errors (e.g. unique_violation triggered by
		// the UPDATE itself) until iteration finishes — without this check the
		// caller would see (0, nil) instead of the real failure.
		if err := rows.Err(); err != nil {
			return n, err
		}
		if n > 0 {
			clean(ctx, e.cacheSource)
		}
		return n, nil
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

// returningColumnsOf extracts the RETURNING column list from stmt if it
// supports the ReturningStmt capability; returns nil otherwise.
func returningColumnsOf(stmt Stmt) []types.Column {
	rs, ok := stmt.(ReturningStmt)
	if !ok {
		return nil
	}
	return rs.ReturningColumns()
}

// scanPointers builds the slice of *T field pointers that match the order of
// the RETURNING columns, suitable for rows.Scan.
func scanPointers(cols []types.Column, model any) []any {
	ptrs := make([]any, len(cols))
	for i, c := range cols {
		ptrs[i] = c.GetPtr(model)
	}
	return ptrs
}

func (e *executor[TModel]) Count(ctx context.Context, stmt CountStmt) (count uint64, err error) {
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
