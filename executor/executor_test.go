package executor

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/insei/gerpo/executor/cache"
	"github.com/insei/gerpo/sqlstmt"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStmt struct {
	mock.Mock
}

type mockColumns struct {
	mock.Mock
	types.ExecutionColumns
}

func (m *mockColumns) GetModelPointers(model any) []any {
	rets := m.Called(model)
	return rets.Get(0).([]any)
}

func (m *mockColumns) GetModelValues(model any) []interface{} {
	rets := m.Called(model)
	return rets.Get(0).([]interface{})
}

func (m *mockStmt) SQL(opts ...sqlstmt.Option) (string, []interface{}) {
	optsAny := make([]any, len(opts))
	for i, opt := range opts {
		optsAny[i] = opt
	}
	rets := m.Called(optsAny...)
	return rets.String(0), rets.Get(1).([]interface{})
}

func (m *mockStmt) Columns() types.ExecutionColumns {
	rets := m.Called()
	return rets.Get(0).(types.ExecutionColumns)
}

type testModel struct {
	ID   int
	Age  int
	Name string
}

func TestGetOne(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		withStmt    *mockStmt
		setupDb     func(sqlmock.Sqlmock)
		cacheBundle func() cache.Source
		expectedErr error
	}{
		{
			name: "Return error in QueryContext",
			ctx:  context.Background(),
			withStmt: func() *mockStmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return("query", []interface{}{})
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("query").WillReturnError(dbsql.ErrTxDone)
			},
			expectedErr: dbsql.ErrTxDone,
		},
		{
			name: "Zero in rows",
			ctx:  context.Background(),
			withStmt: func() *mockStmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return("query", []interface{}{})
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("query").WillReturnRows(sqlmock.NewRows(nil)).RowsWillBeClosed()
			},
			expectedErr: dbsql.ErrNoRows,
		},
		{
			name: "Success",
			ctx:  context.Background(),
			withStmt: func() *mockStmt {
				model := new(testModel)
				stmt := new(mockStmt)
				stmt.
					On("SQL").
					Return("SELECT id, age, name FROM users LIMIT 1", []interface{}{})

				columns := new(mockColumns)
				columns.On("GetModelPointers", mock.Anything).Return([]any{&model.ID, &model.Age, &model.Name})

				stmt.On("Columns").Return(columns)
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectQuery("SELECT id, age, name FROM users LIMIT 1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "age", "name"}).AddRow(1, 2, "test")).
					RowsWillBeClosed()
			},
			expectedErr: nil,
		},
		{
			name: "Scan error not enough model pointers",
			ctx:  context.Background(),
			withStmt: func() *mockStmt {
				model := new(testModel)
				stmt := new(mockStmt)
				stmt.
					On("SQL").
					Return("SELECT id, age, name FROM users LIMIT 1", []interface{}{})
				columns := new(mockColumns)
				columns.On("GetModelPointers", mock.Anything).Return([]any{&model.ID, &model.Age})

				stmt.On("Columns").Return(columns)
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectQuery("SELECT id, age, name FROM users LIMIT 1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "age", "name"}).
						AddRow(1, 2, "test")).
					RowsWillBeClosed()
			},
			expectedErr: fmt.Errorf("sql: expected 3 destination arguments in Scan, not 2"),
		},
		{
			name: "get from cache",
			ctx:  context.Background(),
			withStmt: func() *mockStmt {
				stmt := new(mockStmt)
				stmt.
					On("SQL").
					Return("query", []interface{}{})
				return stmt
			}(),
			cacheBundle: func() cache.Source {
				b := &MockCacheSource{}
				b.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(testModel{ID: 1, Age: 2, Name: "test"}, nil)
				return b
			},
			expectedErr: nil,
		},
		// Add more tests for other scenarios here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mockDB, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if tt.setupDb != nil {
				tt.setupDb(mockDB)
			}

			e := &executor[testModel]{
				db:          db,
				placeholder: func(s string) string { return s },
			}
			if tt.cacheBundle != nil {
				e.cacheSource = tt.cacheBundle()
			}

			_, err = e.GetOne(tt.ctx, tt.withStmt)
			if (err != nil) == (tt.expectedErr != nil) && err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("executor.GetOne() error = %v, wantErr %v", err, tt.expectedErr)
			}

			if err := mockDB.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetMultiple(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		withStmt    Stmt
		setupDb     func(sqlmock.Sqlmock)
		cacheBundle func() cache.Source
		expectedErr error
	}{
		{
			name: "Failure in QueryContext",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return("query", []interface{}{})
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("query").WillReturnError(dbsql.ErrTxDone)
			},
			expectedErr: dbsql.ErrTxDone,
		},
		{
			name: "Zero rows returned it's ok",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return("query", []interface{}{})
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("query").WillReturnRows(sqlmock.NewRows(nil)).RowsWillBeClosed()
			},
			expectedErr: nil,
		},
		{
			name: "Successful query with multiple rows",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				model := []testModel{{}, {}}
				stmt := new(mockStmt)
				stmt.On("SQL").Return("SELECT id, age, name FROM users", []interface{}{})

				columns := new(mockColumns)
				columns.On("GetModelPointers", mock.Anything).Return([]any{&model[0].ID, &model[0].Age, &model[0].Name})
				stmt.On("Columns").Return(columns)

				columns = new(mockColumns)
				columns.On("GetModelPointers", mock.Anything).Return([]any{&model[1].ID, &model[1].Age, &model[1].Name})
				stmt.On("Columns").Return(columns)

				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, age, name FROM users").
					WillReturnRows(sqlmock.NewRows([]string{"id", "age", "name"}).
						AddRow(1, 2, "test").AddRow(2, 2, "test")).
					RowsWillBeClosed()
			},
			expectedErr: nil,
		},
		{
			name: "No query executed fetching from cache",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return("query", []interface{}{})
				return stmt
			}(),
			cacheBundle: func() cache.Source {
				b := &MockCacheSource{}
				b.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return([]*testModel{{ID: 1, Age: 2, Name: "test"}, {ID: 3, Age: 4, Name: "test2"}}, nil)
				return b
			},
			expectedErr: nil,
		},
		{
			name: "Rows returned but not enough model pointers",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				model := new(testModel)
				stmt := new(mockStmt)
				stmt.On("SQL").Return("SELECT id, age, name FROM users LIMIT 1", []interface{}{})

				columns := new(mockColumns)
				columns.On("GetModelPointers", mock.Anything).Return([]any{&model.ID, &model.Age})
				stmt.On("Columns").Return(columns)
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectQuery("SELECT id, age, name FROM users LIMIT 1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "age", "name"}).
						AddRow(1, 2, "test")).
					RowsWillBeClosed()
			},
			expectedErr: fmt.Errorf("sql: expected 3 destination arguments in Scan, not 2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mockDB, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if tt.setupDb != nil {
				tt.setupDb(mockDB)
			}

			e := &executor[testModel]{
				db:          db,
				placeholder: func(s string) string { return s },
			}
			if tt.cacheBundle != nil {
				e.cacheSource = tt.cacheBundle()
			}

			_, err = e.GetMultiple(tt.ctx, tt.withStmt)
			if (err != nil) != (tt.expectedErr != nil) || (err != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("executor.GetMultiple() error = %v, wantErr %v", err, tt.expectedErr)
			}

			if err := mockDB.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestInsertOne(t *testing.T) {
	tests := []struct {
		name        string
		withModel   func() *testModel
		withStmt    func() Stmt
		setupDb     func(sqlmock.Sqlmock)
		cacheBundle func() cache.Source
		expectedErr error
	}{
		{
			name: "DB error during insertion",
			withModel: func() *testModel {
				return &testModel{
					ID:   1,
					Age:  28,
					Name: "John Doe",
				}
			},
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL", mock.Anything).Return("INSERT INTO users (id, age, name) VALUES ($1, $2, $3)", []any{1, 28, "John Doe"})
				return stmt
			},
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users \(id, age, name\) VALUES \(\$1, \$2, \$3\)`).WithArgs(1, 28, "John Doe").WillReturnError(dbsql.ErrTxDone)
			},
			expectedErr: dbsql.ErrTxDone,
		},
		{
			name: "DB ",
			withModel: func() *testModel {
				return &testModel{
					ID:   1,
					Age:  28,
					Name: "John Doe",
				}
			},
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL", mock.Anything).Return("INSERT INTO users (id, age, name) VALUES ($1, $2, $3)", []any{1, 28, "John Doe"})
				return stmt
			},
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users \(id, age, name\) VALUES \(\$1, \$2, \$3\)`).WithArgs(1, 28, "John Doe").WillReturnError(dbsql.ErrTxDone)
			},
			expectedErr: dbsql.ErrTxDone,
		},
		{
			name: "Rows affected result error",
			withModel: func() *testModel {
				return &testModel{
					ID:   1,
					Age:  28,
					Name: "John Doe",
				}
			},
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL", mock.Anything).Return("INSERT INTO users (id, age, name) VALUES ($1, $2, $3)", []any{1, 28, "John Doe"})
				return stmt
			},
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users \(id, age, name\) VALUES \(\$1, \$2, \$3\)`).WithArgs(1, 28, "John Doe").WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("result error")))
			},
			expectedErr: fmt.Errorf("result error"),
		},
		{
			name: "Rows affected no rows",
			withModel: func() *testModel {
				return &testModel{
					ID:   1,
					Age:  28,
					Name: "John Doe",
				}
			},
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL", mock.Anything).Return("INSERT INTO users (id, age, name) VALUES ($1, $2, $3)", []any{1, 28, "John Doe"})
				return stmt
			},
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users \(id, age, name\) VALUES \(\$1, \$2, \$3\)`).WithArgs(1, 28, "John Doe").WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedErr: ErrNoInsertedRows,
		},
		{
			name: "Successful insertion",
			withModel: func() *testModel {
				return &testModel{
					ID:   1,
					Age:  28,
					Name: "John Doe",
				}
			},
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL", mock.Anything).Return("INSERT INTO users (id, age, name) VALUES ($1, $2, $3)", []any{1, 28, "John Doe"})
				return stmt
			},
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users \(id, age, name\) VALUES \(\$1, \$2, \$3\)`).WithArgs(1, 28, "John Doe").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			cacheBundle: func() cache.Source {
				b := &MockCacheSource{}
				b.On("Clean", mock.Anything)
				return b
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mockDB, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if tt.setupDb != nil {
				tt.setupDb(mockDB)
			}

			e := &executor[testModel]{
				db:          db,
				placeholder: func(s string) string { return s },
			}
			if tt.cacheBundle != nil {
				e.cacheSource = tt.cacheBundle()
			}

			err = e.InsertOne(context.Background(), tt.withStmt(), tt.withModel())
			if (err != nil) != (tt.expectedErr != nil) || (err != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("executor.InsertOne() error = %v, wantErr %v", err, tt.expectedErr)
			}

			if err := mockDB.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name                string
		withModel           func() *testModel
		withStmt            func() Stmt
		setupDb             func(sqlmock.Sqlmock)
		cacheBundle         func() cache.Source
		expectedErr         error
		expectedUpdatedRows int64
	}{
		{
			name: "DB error during update",
			withModel: func() *testModel {
				return &testModel{
					ID:   1,
					Age:  28,
					Name: "John Doe",
				}
			},
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL", mock.Anything).
					Return("UPDATE users SET age = $1, name = $2 WHERE id = $3", []any{28, "John Doe", 1})
				return stmt
			},
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET age = \$1, name = \$2 WHERE id = \$3`).
					WithArgs(28, "John Doe", 1).WillReturnError(dbsql.ErrTxDone)
			},
			expectedErr: dbsql.ErrTxDone,
		},
		{
			name: "Rows affected result error",
			withModel: func() *testModel {
				return &testModel{
					ID:   1,
					Age:  28,
					Name: "John Doe",
				}
			},
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL", mock.Anything).
					Return("UPDATE users SET age = $1, name = $2 WHERE id = $3", []any{28, "John Doe", 1})
				return stmt
			},
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET age = \$1, name = \$2 WHERE id = \$3`).
					WithArgs(28, "John Doe", 1).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("result error")))
			},
			expectedErr: fmt.Errorf("result error"),
		},
		{
			name: "Rows affected more than 0 rows updated",
			withModel: func() *testModel {
				return &testModel{
					ID:   1,
					Age:  28,
					Name: "John Doe",
				}
			},
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL", mock.Anything).
					Return("UPDATE users SET age = $1, name = $2 WHERE id = $3", []any{28, "John Doe", 1})
				return stmt
			},
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET age = \$1, name = \$2 WHERE id = \$3`).
					WithArgs(28, "John Doe", 1).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			cacheBundle: func() cache.Source {
				b := &MockCacheSource{}
				b.On("Clean", mock.Anything)
				return b
			},
			expectedUpdatedRows: 1,
		},
		// Define more tests here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mockDB, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if tt.setupDb != nil {
				tt.setupDb(mockDB)
			}

			e := &executor[testModel]{
				db:          db,
				placeholder: func(s string) string { return s },
			}
			if tt.cacheBundle != nil {
				e.cacheSource = tt.cacheBundle()
			}

			updatedRows, err := e.Update(context.Background(), tt.withStmt(), tt.withModel())
			if (err != nil) != (tt.expectedErr != nil) ||
				(err != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("executor.Update() error = %v, wantErr %v", err, tt.expectedErr)
			}
			if updatedRows != tt.expectedUpdatedRows {
				t.Errorf("executor.Update() updatedRows = %d, want %d", updatedRows, tt.expectedUpdatedRows)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		withStmt    Stmt
		setupDb     func(sqlmock.Sqlmock)
		cacheBundle func() cache.Source
		expectedErr error
		expectedRes uint64
	}{
		{
			name: "Error in QueryContext",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return(`SELECT COUNT(*) FROM users`, []interface{}{})
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnError(dbsql.ErrTxDone)
			},
			expectedErr: dbsql.ErrTxDone,
			expectedRes: 0,
		},
		{
			name: "Successful count",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return(`SELECT COUNT(*) FROM users`, []interface{}{})
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(10)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnRows(rows)
			},
			expectedErr: nil,
			expectedRes: 10,
		},
		{
			name: "Count from cache",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return(`SELECT COUNT(*) FROM users`, []interface{}{})
				return stmt
			}(),
			cacheBundle: func() cache.Source {
				b := &MockCacheSource{}
				b.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(uint64(20), nil)
				return b
			},
			expectedErr: nil,
			expectedRes: 20,
		},
		{
			name: "Count scan error not enough model pointers",
			ctx:  context.Background(),
			withStmt: func() Stmt {
				stmt := new(mockStmt)
				stmt.On("SQL").Return(`SELECT COUNT(*) FROM users`, []interface{}{})
				return stmt
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow("count")
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnRows(rows)
			},
			expectedErr: fmt.Errorf("sql: Scan error on column index 0, name \"count\": converting driver.Value type string (\"count\") to a uint64: invalid syntax"),
			expectedRes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mockDB, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if tt.setupDb != nil {
				tt.setupDb(mockDB)
			}

			e := &executor[testModel]{
				db:          db,
				placeholder: func(s string) string { return s },
			}
			if tt.cacheBundle != nil {
				e.cacheSource = tt.cacheBundle()
			}

			res, err := e.Count(tt.ctx, tt.withStmt)
			if (err != nil) != (tt.expectedErr != nil) || (err != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("executor.Count() error = %v, wantErr %v", err, tt.expectedErr)
			}
			if res != tt.expectedRes {
				t.Errorf("executor.Count() result = %v, expectedRes %v", res, tt.expectedRes)
			}

			if err := mockDB.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		stmt        Stmt
		setupDb     func(sqlmock.Sqlmock)
		expectedErr error
		expectedRes int64
	}{
		{
			name: "Error in ExecContext",
			ctx:  context.Background(),
			stmt: func() Stmt {
				s := &mockStmt{}
				s.On("SQL").Return("DELETE FROM users WHERE id=1", []interface{}{})
				return s
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id=1").WillReturnError(dbsql.ErrTxDone)
			},
			expectedErr: dbsql.ErrTxDone,
			expectedRes: 0,
		},
		{
			name: "Successful delete",
			ctx:  context.Background(),
			stmt: func() Stmt {
				s := &mockStmt{}
				s.On("SQL").Return("DELETE FROM users WHERE id=1", []interface{}{})
				return s
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id=1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: nil,
			expectedRes: 1,
		},
		{
			name: "Error in result",
			ctx:  context.Background(),
			stmt: func() Stmt {
				s := &mockStmt{}
				s.On("SQL").Return("DELETE FROM users WHERE id=1", []interface{}{})
				return s
			}(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id=1").
					WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("result error")))
			},
			expectedErr: fmt.Errorf("result error"),
			expectedRes: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mockDB, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if tt.setupDb != nil {
				tt.setupDb(mockDB)
			}

			e := &executor[testModel]{
				db:          db,
				placeholder: func(s string) string { return s },
			}

			res, err := e.Delete(tt.ctx, tt.stmt)
			if (err != nil) != (tt.expectedErr != nil) || (err != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("executor.Delete() error = %v, wantErr %v", err, tt.expectedErr)
			}
			if res != tt.expectedRes {
				t.Errorf("executor.Delete() result = %v, expectedRes %v", res, tt.expectedRes)
			}

			if err := mockDB.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetExecQuery(t *testing.T) {
	tests := []struct {
		name                 string
		ctx                  context.Context
		getExecQueryReplaced func(ctx context.Context) ExecQuery
		expectedExecQueryNil bool
	}{
		{
			name:                 "Return original db instance",
			ctx:                  context.Background(),
			getExecQueryReplaced: nil,
			expectedExecQueryNil: false,
		},
		{
			name: "Return replaced ExecQuery",
			ctx:  context.Background(),
			getExecQueryReplaced: func(ctx context.Context) ExecQuery {
				return nil
			},
			expectedExecQueryNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &dbsql.DB{}
			e := &executor[testModel]{
				db:                   db,
				getExecQueryReplaced: tt.getExecQueryReplaced,
				placeholder:          func(s string) string { return s },
			}
			execQuery := e.getExecQuery(tt.ctx)
			if (execQuery == nil) != tt.expectedExecQueryNil {
				t.Errorf("executor.getExecQuery() is %v, but expected %v", e.getExecQuery(tt.ctx), nil)
			}
		})
	}
}

func TestTx(t *testing.T) {
	sameDB := &dbsql.DB{}
	testCases := []struct {
		description string
		mockDB      func() *dbsql.DB
		txDB        func() *dbsql.DB
		expectedErr error
	}{
		{
			description: "success - Tx db is the same with executor db",
			mockDB: func() *dbsql.DB {
				return sameDB
			},
			txDB: func() *dbsql.DB {
				return sameDB
			},
			expectedErr: nil,
		},
		{
			description: "error - Tx db is not the same with executor db",
			mockDB: func() *dbsql.DB {
				return sameDB
			},
			txDB: func() *dbsql.DB {
				return &dbsql.DB{}
			},
			expectedErr: ErrTxDBNotSame,
		},
	}

	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			// Prepare executor
			db := test.mockDB()
			e := New[any](db)

			// Prepare Tx
			txDb := test.txDB()
			tx := &Tx{
				db: txDb,
			}

			// Executing method
			_, err := e.Tx(tx)

			// Asserting
			assert.Equal(t, test.expectedErr, err)
		})
	}
}
