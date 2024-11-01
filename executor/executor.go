package executor

import (
	"context"
	dbsql "database/sql"

	"github.com/insei/gerpo/sql"
)

type executor[TModel any] struct {
	db          SQLDB
	placeholder Placeholder

	options
}

func New[TModel any](db *dbsql.DB, opts ...Option) Executor[TModel] {
	o := &options{}
	for _, opt := range opts {
		opt.apply(o)
	}
	e := &executor[TModel]{
		options:     *o,
		db:          db,
		placeholder: determinePlaceHolder(db),
	}
	return e
}

func (e *executor[TModel]) GetOne(ctx context.Context, selectStmt sql.StmtSelect) (*TModel, error) {
	sqlStmt, args := selectStmt.GetStmtWithArgs(sql.SelectOne)
	if cached, ok := get[TModel](ctx, e.cacheBundle, sqlStmt, args...); ok {
		return cached, nil
	}

	rows, err := e.db.QueryContext(ctx, e.placeholder(sqlStmt), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var model *TModel
	if rows.Next() {
		model = new(TModel)
		pointers := selectStmt.GetModelPointers(sql.SelectOne, model)
		if err = rows.Scan(pointers...); err != nil {
			return nil, err
		}
		set(ctx, e.cacheBundle, *model, sqlStmt, args...)
	}
	if model == nil {
		return nil, dbsql.ErrNoRows
	}
	return model, nil
}

func (e *executor[TModel]) GetMultiple(ctx context.Context, selectStmt sql.StmtSelect) ([]*TModel, error) {
	sqlStmt, args := selectStmt.GetStmtWithArgs(sql.Select)
	if cached, ok := get[[]*TModel](ctx, e.cacheBundle, sqlStmt, args...); ok {
		return *cached, nil
	}
	rows, err := e.db.QueryContext(ctx, e.placeholder(sqlStmt), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var models []*TModel
	for rows.Next() {
		model := new(TModel)
		if err = rows.Scan(selectStmt.GetModelPointers(sql.Select, model)...); err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	set(ctx, e.cacheBundle, models, sqlStmt, args...)
	return models, nil
}

func (e *executor[TModel]) InsertOne(ctx context.Context, model *TModel, stmtModel sql.StmtModel) error {
	sqlStmt, values := stmtModel.GetStmtWithArgsForModel(sql.Insert, model)
	result, err := e.db.ExecContext(ctx, e.placeholder(sqlStmt), values...)
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
	clean(ctx, e.cacheBundle)
	return nil
}

func (e *executor[TModel]) Update(ctx context.Context, model *TModel, stmtModel sql.StmtModel) (int64, error) {
	sqlStmt, values := stmtModel.GetStmtWithArgsForModel(sql.Update, model)
	result, err := e.db.ExecContext(ctx, e.placeholder(sqlStmt), values...)
	if err != nil {
		return 0, err
	}
	updatedRows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if updatedRows > 0 {
		clean(ctx, e.cacheBundle)
	}
	return updatedRows, nil
}

func (e *executor[TModel]) Count(ctx context.Context, stmt sql.Stmt) (uint64, error) {
	sqlStmt, args := stmt.GetStmtWithArgs(sql.Count)
	if cached, ok := get[uint64](ctx, e.cacheBundle, sqlStmt, args...); ok {
		return *cached, nil
	}
	count := uint64(0)
	rows, err := e.db.QueryContext(ctx, e.placeholder(sqlStmt), args...)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	set(ctx, e.cacheBundle, count, sqlStmt, args...)
	return count, nil
}

func (e *executor[TModel]) Delete(ctx context.Context, stmt sql.Stmt) (int64, error) {
	sqlStmt, args := stmt.GetStmtWithArgs(sql.Delete)
	result, err := e.db.ExecContext(ctx, e.placeholder(sqlStmt), args...)
	if err != nil {
		return 0, err
	}
	deletedRows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if deletedRows > 0 {
		clean(ctx, e.cacheBundle)
	}
	return deletedRows, nil
}
