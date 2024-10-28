package sql

import (
	"context"
	dbsql "database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/insei/fmap/v3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/cache"
	"github.com/insei/gerpo/types"
)

func TestNewExecutor(t *testing.T) {
	mockDB := &dbsql.DB{}

	executor := NewExecutor[string](mockDB)

	assert.NotNil(t, executor)
	assert.Equal(t, mockDB, executor.db)
	assert.NotNil(t, executor.placeholder)
}

func TestExecutorGetOne(t *testing.T) {
	testCases := []struct {
		name          string
		ctx           context.Context
		cacheValue    interface{}
		sqlQuery      string
		sqlArgs       []driver.Value
		mockRows      *sqlmock.Rows
		expectedModel *TestModel
		mockError     error
		expectedError error
	}{
		{
			name:          "Cache hit",
			ctx:           cache.NewCtxCache(context.Background()),
			cacheValue:    TestModel{Int: 456},
			sqlQuery:      "SELECT Int FROM table LIMIT 1",
			sqlArgs:       []driver.Value{},
			mockRows:      nil,
			expectedModel: &TestModel{Int: 456},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "Cache miss with result",
			ctx:           context.Background(),
			cacheValue:    nil,
			sqlQuery:      "SELECT Int FROM table LIMIT 1",
			sqlArgs:       []driver.Value{},
			mockRows:      sqlmock.NewRows([]string{"Int"}).AddRow(123),
			expectedModel: &TestModel{Int: 123},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "SQL error",
			ctx:           context.Background(),
			cacheValue:    nil,
			sqlQuery:      "SELECT Int FROM table LIMIT 1",
			sqlArgs:       []driver.Value{},
			mockRows:      nil,
			expectedModel: nil,
			mockError:     errors.New("unexpected error"),
			expectedError: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer mockDB.Close()

			executor := &Executor[TestModel]{
				db:          mockDB,
				placeholder: determinePlaceHolder(mockDB),
			}

			sb := &StringBuilder{
				ctx:   tc.ctx,
				table: "table",
				selectBuilder: &StringSelectBuilder{columns: []types.Column{
					&testColumn{sql: "Int"},
				}},
				joinBuilder:  &StringJoinBuilder{},
				groupBuilder: &StringGroupBuilder{},
				whereBuilder: &StringWhereBuilder{},
			}

			if tc.cacheValue != nil {
				cache.AppendToCtxCache[TestModel](tc.ctx, fmt.Sprintf("%s%v", tc.sqlQuery, tc.sqlArgs), tc.cacheValue)
			}

			if tc.mockRows != nil {
				mock.ExpectQuery(tc.sqlQuery).WithArgs(tc.sqlArgs...).WillReturnRows(tc.mockRows)
			} else if tc.mockError != nil {
				mock.ExpectQuery(tc.sqlQuery).WithArgs(tc.sqlArgs...).WillReturnError(tc.mockError)
			}

			result, err := executor.GetOne(tc.ctx, sb)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedModel.Int, result.Int)
			}

			if err = mock.ExpectationsWereMet(); err != nil && (tc.mockRows != nil || tc.expectedError != nil) {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestExecutorGetMultiple(t *testing.T) {
	testCases := []struct {
		name           string
		ctx            context.Context
		cacheValue     interface{}
		sqlQuery       string
		sqlArgs        []driver.Value
		mockRows       *sqlmock.Rows
		expectedModels []*TestModel
		mockError      error
		expectedError  error
	}{
		{
			name:           "Cache hit",
			ctx:            cache.NewCtxCache(context.Background()),
			cacheValue:     []*TestModel{{Int: 456}, {Int: 789}},
			sqlQuery:       "SELECT Int FROM table",
			sqlArgs:        []driver.Value{},
			mockRows:       nil,
			expectedModels: []*TestModel{{Int: 456}, {Int: 789}},
			mockError:      nil,
			expectedError:  nil,
		},
		{
			name:           "Cache miss with result",
			ctx:            context.Background(),
			cacheValue:     nil,
			sqlQuery:       "SELECT Int FROM table",
			sqlArgs:        []driver.Value{},
			mockRows:       sqlmock.NewRows([]string{"Int"}).AddRow(123).AddRow(456),
			expectedModels: []*TestModel{{Int: 123}, {Int: 456}},
			mockError:      nil,
			expectedError:  nil,
		},
		{
			name:           "Cache miss with no rows",
			ctx:            context.Background(),
			cacheValue:     nil,
			sqlQuery:       "SELECT Int FROM table",
			sqlArgs:        []driver.Value{},
			mockRows:       sqlmock.NewRows([]string{"Int"}),
			expectedModels: []*TestModel{},
			mockError:      nil,
			expectedError:  nil,
		},
		{
			name:           "SQL error",
			ctx:            context.Background(),
			cacheValue:     nil,
			sqlQuery:       "SELECT Int FROM table",
			sqlArgs:        []driver.Value{},
			mockRows:       nil,
			expectedModels: nil,
			mockError:      errors.New("unexpected error"),
			expectedError:  errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer mockDB.Close()

			executor := &Executor[TestModel]{
				db:          mockDB,
				placeholder: determinePlaceHolder(mockDB),
			}

			sb := &StringBuilder{
				ctx:   tc.ctx,
				table: "table",
				selectBuilder: &StringSelectBuilder{
					columns: []types.Column{
						&testColumn{sql: "Int"},
					}},
				joinBuilder:  &StringJoinBuilder{},
				groupBuilder: &StringGroupBuilder{},
				whereBuilder: &StringWhereBuilder{},
			}

			if tc.cacheValue != nil {
				cache.AppendToCtxCache[TestModel](tc.ctx, fmt.Sprintf("%s%v", tc.sqlQuery, tc.sqlArgs), tc.cacheValue)
			}

			if tc.mockRows != nil {
				mock.ExpectQuery(tc.sqlQuery).WithArgs(tc.sqlArgs...).WillReturnRows(tc.mockRows)
			} else if tc.mockError != nil {
				mock.ExpectQuery(tc.sqlQuery).WithArgs(tc.sqlArgs...).WillReturnError(tc.mockError)
			}

			result, err := executor.GetMultiple(tc.ctx, sb)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedModels), len(result))
				for i, model := range result {
					assert.Equal(t, tc.expectedModels[i].Int, model.Int)
				}
			}

			if err = mock.ExpectationsWereMet(); err != nil && (tc.mockRows != nil || tc.expectedError != nil) {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestExecutorInsertOne(t *testing.T) {
	testCases := []struct {
		name          string
		ctx           context.Context
		model         *TestModel
		sqlQuery      string
		sqlArgs       []driver.Value
		expectedError error
		execResult    driver.Result
		execError     error
	}{
		{
			name:          "Successful insert",
			ctx:           context.Background(),
			model:         &TestModel{Int: 123},
			sqlQuery:      "INSERT INTO table (Int) VALUES ($1)",
			sqlArgs:       []driver.Value{123},
			expectedError: nil,
			execResult:    sqlmock.NewResult(1, 1),
			execError:     nil,
		},
		{
			name:          "Insert failed with exec error",
			ctx:           context.Background(),
			model:         &TestModel{Int: 123},
			sqlQuery:      "INSERT INTO table (Int) VALUES ($1)",
			sqlArgs:       []driver.Value{123},
			expectedError: errors.New("exec error"),
			execResult:    nil,
			execError:     errors.New("exec error"),
		},
		{
			name:          "Insert affected 0 rows",
			ctx:           context.Background(),
			model:         &TestModel{Int: 123},
			sqlQuery:      "INSERT INTO table (Int) VALUES ($1)",
			sqlArgs:       []driver.Value{123},
			expectedError: fmt.Errorf("failed to insert: inserted 0 rows"),
			execResult:    sqlmock.NewResult(1, 0),
			execError:     nil,
		},
		{
			name:          "RowsAffected error",
			ctx:           context.Background(),
			model:         &TestModel{Int: 123},
			sqlQuery:      "INSERT INTO table (Int) VALUES ($1)",
			sqlArgs:       []driver.Value{123},
			expectedError: errors.New("rows affected error"),
			execResult:    sqlmock.NewErrorResult(errors.New("rows affected error")),
			execError:     nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer mockDB.Close()

			executor := &Executor[TestModel]{
				db:          mockDB,
				placeholder: determinePlaceHolder(mockDB),
			}
			fields, _ := fmap.GetFrom(tc.model)

			sb := &StringBuilder{
				ctx:   tc.ctx,
				table: "table",
				insertBuilder: &StringInsertBuilder{
					columns: []types.Column{
						&testColumn{
							sql:   "Int",
							field: fields.MustFind("Int"),
						},
					},
				},
			}

			mock.ExpectExec(regexp.QuoteMeta(tc.sqlQuery)).WithArgs(tc.sqlArgs...).WillReturnResult(tc.execResult).WillReturnError(tc.execError)

			err = executor.InsertOne(tc.ctx, tc.model, sb)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestExecutorUpdate(t *testing.T) {
	uuidExample := uuid.New()

	testCases := []struct {
		name          string
		ctx           context.Context
		model         *TestModel
		sqlQuery      string
		sqlArgs       []driver.Value
		expectedError string
		expectedRows  int64
		execResult    driver.Result
		execError     error
	}{
		{
			name:          "Successful update",
			ctx:           context.Background(),
			model:         &TestModel{Int: 123, String: "test", UUID: uuidExample},
			sqlQuery:      `UPDATE table SET Int = ?, String = ? WHERE UUID = ?`,
			sqlArgs:       []driver.Value{123, "test", uuidExample},
			expectedError: "",
			expectedRows:  1,
			execResult:    sqlmock.NewResult(1, 1),
			execError:     nil,
		},
		{
			name:          "Update failed with exec error",
			ctx:           context.Background(),
			model:         &TestModel{Int: 123, String: "test", UUID: uuidExample},
			sqlQuery:      `UPDATE table SET Int = ?, String = ? WHERE UUID = ?`,
			sqlArgs:       []driver.Value{123, "test", uuidExample},
			expectedError: "exec error",
			expectedRows:  0,
			execResult:    nil,
			execError:     errors.New("exec error"),
		},
		{
			name:          "RowsAffected error",
			ctx:           context.Background(),
			model:         &TestModel{Int: 123, String: "test", UUID: uuidExample},
			sqlQuery:      `UPDATE table SET Int = ?, String = ? WHERE UUID = ?`,
			sqlArgs:       []driver.Value{123, "test", uuidExample},
			expectedError: "rows affected error",
			expectedRows:  0,
			execResult:    sqlmock.NewErrorResult(errors.New("rows affected error")),
			execError:     nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer mockDB.Close()

			executor := &Executor[TestModel]{
				db: mockDB,
				placeholder: func(query string) string {
					return query
				},
			}

			fields, _ := fmap.GetFrom(tc.model)

			sb := &StringBuilder{
				ctx:   tc.ctx,
				table: "table",
				updateBuilder: &StringUpdateBuilder{
					columns: []types.Column{
						&testColumn{
							sql:   "Int",
							field: fields.MustFind("Int"),
						},
						&testColumn{
							sql:   "String",
							field: fields.MustFind("String"),
						},
					},
				},
				whereBuilder: &StringWhereBuilder{
					ctx:    tc.ctx,
					values: []any{tc.model.UUID},
					sql:    "UUID = ?",
				},
			}

			mock.ExpectExec(regexp.QuoteMeta(tc.sqlQuery)).WithArgs(tc.sqlArgs...).WillReturnResult(tc.execResult).WillReturnError(tc.execError)

			updatedRows, err := executor.Update(tc.ctx, tc.model, sb)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRows, updatedRows)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestExecutorCount(t *testing.T) {
	uuidExample := uuid.New()

	testCases := []struct {
		name          string
		ctx           context.Context
		sqlQuery      string
		sqlArgs       []driver.Value
		values        []any
		cachedResult  interface{}
		expectedError string
		expectedCount uint64
		queryRowError error
	}{
		{
			name:          "Count from database",
			ctx:           cache.NewCtxCache(context.Background()),
			sqlQuery:      `SELECT count(*) over() AS count FROM table WHERE UUID = ? LIMIT 1`,
			sqlArgs:       []driver.Value{uuidExample},
			values:        []any{uuidExample},
			cachedResult:  nil,
			expectedError: "",
			expectedCount: 42,
			queryRowError: nil,
		},
		{
			name:          "Count from cache",
			ctx:           cache.NewCtxCache(context.Background()),
			sqlQuery:      `SELECT count(*) over() AS count FROM table WHERE UUID = ? LIMIT 1`,
			sqlArgs:       []driver.Value{uuidExample},
			values:        []any{uuidExample},
			cachedResult:  uint64(10),
			expectedError: "",
			expectedCount: 10,
			queryRowError: nil,
		},
		{
			name:          "Error during query",
			ctx:           context.Background(),
			sqlQuery:      `SELECT count(*) over() AS count FROM table WHERE UUID = ? LIMIT 1`,
			sqlArgs:       []driver.Value{uuidExample},
			values:        []any{uuidExample},
			cachedResult:  nil,
			expectedError: "query error",
			expectedCount: 0,
			queryRowError: errors.New("query error"),
		},
		{
			name:          "No rows returned",
			ctx:           context.Background(),
			sqlQuery:      `SELECT count(*) over() AS count FROM table WHERE UUID = ? LIMIT 1`,
			sqlArgs:       []driver.Value{uuidExample},
			values:        []any{uuidExample},
			cachedResult:  nil,
			expectedError: "",
			expectedCount: 0,
			queryRowError: dbsql.ErrNoRows,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer mockDB.Close()

			if tc.cachedResult != nil {
				cache.AppendToCtxCache[TestModel](tc.ctx, fmt.Sprintf("%s%v", tc.sqlQuery, tc.sqlArgs), tc.cachedResult)
			}

			executor := &Executor[TestModel]{
				db: mockDB,
				placeholder: func(query string) string {
					return query
				},
			}

			if tc.cachedResult == nil {
				rows := sqlmock.NewRows([]string{"count"})
				if tc.queryRowError == nil {
					rows.AddRow(tc.expectedCount)
				}

				mock.ExpectQuery(regexp.QuoteMeta(tc.sqlQuery)).
					WithArgs(tc.sqlArgs...).
					WillReturnError(tc.queryRowError).
					WillReturnRows(rows)
			}

			sb := &StringBuilder{
				ctx:           tc.ctx,
				table:         "table",
				selectBuilder: &StringSelectBuilder{},
				groupBuilder:  &StringGroupBuilder{},
				joinBuilder:   &StringJoinBuilder{},
				whereBuilder: &StringWhereBuilder{
					ctx:    tc.ctx,
					values: tc.values,
					sql:    "UUID = ?",
				},
			}

			count, err := executor.Count(tc.ctx, sb)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, count)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestExecutorDelete(t *testing.T) {
	uuidExample := uuid.New()

	testCases := []struct {
		name          string
		ctx           context.Context
		sqlQuery      string
		sqlArgs       []driver.Value
		values        []any
		result        dbsql.Result
		expectedRows  int64
		expectedError string
		execError     error
		rowsAffected  int64
	}{
		{
			name:          "Delete with success",
			ctx:           context.Background(),
			sqlQuery:      `DELETE FROM table WHERE UUID = ?`,
			sqlArgs:       []driver.Value{uuidExample},
			values:        []any{uuidExample},
			expectedRows:  1,
			expectedError: "",
			execError:     nil,
			rowsAffected:  1,
		},
		{
			name:          "Delete with no rows affected",
			ctx:           context.Background(),
			sqlQuery:      `DELETE FROM table WHERE UUID = ?`,
			sqlArgs:       []driver.Value{uuidExample},
			values:        []any{uuidExample},
			expectedRows:  0,
			expectedError: "",
			execError:     nil,
			rowsAffected:  0,
		},
		{
			name:          "Error during delete",
			ctx:           context.Background(),
			sqlQuery:      `DELETE FROM table WHERE UUID = ?`,
			sqlArgs:       []driver.Value{uuidExample},
			values:        []any{uuidExample},
			expectedRows:  0,
			expectedError: "exec error",
			execError:     errors.New("exec error"),
			rowsAffected:  0,
		},
		{
			name:          "No rows error",
			ctx:           context.Background(),
			sqlQuery:      `DELETE FROM table WHERE UUID = ?`,
			sqlArgs:       []driver.Value{uuidExample},
			values:        []any{uuidExample},
			expectedRows:  0,
			expectedError: "",
			execError:     dbsql.ErrNoRows,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer mockDB.Close()

			mockResult := sqlmock.NewResult(1, tc.rowsAffected)
			if tc.execError != nil {
				mock.ExpectExec(regexp.QuoteMeta(tc.sqlQuery)).
					WithArgs(tc.sqlArgs...).
					WillReturnError(tc.execError)
			} else {
				mock.ExpectExec(regexp.QuoteMeta(tc.sqlQuery)).
					WithArgs(tc.sqlArgs...).
					WillReturnResult(mockResult)
			}

			executor := &Executor[TestModel]{
				db: mockDB,
				placeholder: func(query string) string {
					return query
				},
			}

			sb := &StringBuilder{
				ctx:         tc.ctx,
				table:       "table",
				joinBuilder: &StringJoinBuilder{},
				whereBuilder: &StringWhereBuilder{
					ctx:    tc.ctx,
					values: tc.values,
					sql:    "UUID = ?",
				},
			}

			rows, err := executor.Delete(tc.ctx, sb)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRows, rows)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}
