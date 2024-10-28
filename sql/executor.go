package sql

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"

	"github.com/insei/gerpo/cache"
)

type Executor[TModel any] struct {
	db          *dbsql.DB
	placeholder Placeholder
}

func NewExecutor[TModel any](db *dbsql.DB) *Executor[TModel] {
	return &Executor[TModel]{
		db:          db,
		placeholder: determinePlaceHolder(db),
	}
}

func (e *Executor[TModel]) GetOne(ctx context.Context, sql *StringBuilder) (*TModel, error) {
	sql.SelectBuilder().Limit(1)
	sqlQuery, args := sql.SelectSQL()
	if cached, ok := cache.GetFromCtxCache[TModel](ctx, fmt.Sprintf("%s%v", sqlQuery, args)); ok {
		cachedTyped, ok := cached.(TModel)
		if ok {
			fmt.Println("read from cache")
			return &cachedTyped, nil
		}
	}
	model := new(TModel)
	columns := sql.SelectBuilder().GetColumns()
	pointers := make([]interface{}, len(columns))
	for i, cl := range columns {
		pointers[i] = cl.GetPtr(model)
	}
	sqlQuery = e.placeholder(sqlQuery)
	err := e.db.QueryRowContext(ctx, sqlQuery, args...).Scan(pointers...)
	if err != nil {
		return nil, err
	}
	cache.AppendToCtxCache[TModel](ctx, fmt.Sprintf("%s%v", sqlQuery, args), *model)
	return model, nil
}

func (e *Executor[TModel]) GetMultiple(ctx context.Context, sql *StringBuilder) ([]*TModel, error) {
	sqlQuery, args := sql.SelectSQL()
	if cached, ok := cache.GetFromCtxCache[TModel](ctx, fmt.Sprintf("%s%v", sqlQuery, args)); ok {
		cachedTyped, ok := cached.([]*TModel)
		if ok {
			fmt.Println("read from cache")
			return cachedTyped, nil
		}
	}
	columns := sql.SelectBuilder().GetColumns()
	rows, err := e.db.QueryContext(ctx, e.placeholder(sqlQuery), args...)
	if err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var models []*TModel
	for rows.Next() {
		model := new(TModel)
		pointers := make([]interface{}, len(columns))
		for i, cl := range columns {
			pointers[i] = cl.GetPtr(model)
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	cache.AppendToCtxCache[TModel](ctx, fmt.Sprintf("%s%v", sqlQuery, args), models)
	return models, nil
}

func (e *Executor[TModel]) InsertOne(ctx context.Context, model *TModel, sql *StringBuilder) error {
	columns := sql.InsertBuilder().GetColumns()
	values := make([]interface{}, len(columns))
	for i, cl := range columns {
		values[i] = cl.GetField().Get(model)
	}
	result, err := e.db.ExecContext(ctx, e.placeholder(sql.InsertSQL()), values...)
	if err != nil {
		return err
	}
	insertedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if insertedRows == 0 {
		return fmt.Errorf("failed to insert: inserted 0 rows")
	}
	cache.CleanupCtxCache[TModel](ctx)
	return nil
}

func (e *Executor[TModel]) Update(ctx context.Context, model *TModel, sql *StringBuilder) (int64, error) {
	columns := sql.UpdateBuilder().GetColumns()
	values := make([]interface{}, len(columns))
	for i, cl := range columns {
		values[i] = cl.GetField().Get(model)
	}
	values = append(values, sql.WhereBuilder().Values()...)
	result, err := e.db.ExecContext(ctx, e.placeholder(sql.UpdateSQL()), values...)
	if err != nil {
		return 0, err
	}
	updatedRows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if updatedRows > 0 {
		cache.CleanupCtxCache[TModel](ctx)
	}
	return updatedRows, nil
}

func (e *Executor[TModel]) Count(ctx context.Context, sql *StringBuilder) (uint64, error) {
	sqlQuery, args := sql.CountSQL()
	if cached, ok := cache.GetFromCtxCache[TModel](ctx, fmt.Sprintf("%s%v", sqlQuery, args)); ok {
		cachedTyped, ok := cached.(uint64)
		if ok {
			return cachedTyped, nil
		}
	}
	count := uint64(0)
	err := e.db.QueryRowContext(ctx, e.placeholder(sqlQuery), args...).Scan(&count)
	if err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	cache.AppendToCtxCache[TModel](ctx, fmt.Sprintf("%s%v", sqlQuery, args), count)
	return count, nil
}

func (e *Executor[TModel]) Delete(ctx context.Context, sql *StringBuilder) (int64, error) {
	sqlQuery := sql.DeleteSQL()
	result, err := e.db.ExecContext(ctx, e.placeholder(sqlQuery), sql.WhereBuilder().values...)
	if err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	deletedRows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if deletedRows > 0 {
		cache.CleanupCtxCache[TModel](ctx)
	}
	return deletedRows, nil
}
