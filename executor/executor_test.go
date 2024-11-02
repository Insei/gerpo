package executor

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/insei/gerpo/cache"
	"github.com/insei/gerpo/logger"
	"github.com/insei/gerpo/sql"
	"github.com/stretchr/testify/mock"
)

type SQLDBMock struct {
	mock.Mock
}

func (m *SQLDBMock) ExecContext(ctx context.Context, query string, args ...any) (dbsql.Result, error) {
	argsPassed := m.Called(ctx, query, args)
	return argsPassed.Get(0).(dbsql.Result), argsPassed.Error(1)
}

func (m *SQLDBMock) QueryContext(ctx context.Context, query string, args ...any) (*dbsql.Rows, error) {
	argsPassed := m.Called(ctx, query, args)
	return argsPassed.Get(0).(*dbsql.Rows), argsPassed.Error(1)
}

func TestGetExecQuery(t *testing.T) {
	// setup test cases
	tests := []struct {
		name  string
		setup func() (*executor[int], context.Context, ExecQuery)
	}{
		{
			name: "With TxContextKey returns tx",
			setup: func() (*executor[int], context.Context, ExecQuery) {
				db := &dbsql.DB{}
				tx := &dbsql.Tx{}
				executor := &executor[int]{db: db, options: options{}}
				data := newTxData()
				data.setTx(db, tx)
				ctx := context.WithValue(context.Background(), txContextKey, data)
				return executor, ctx, tx
			},
		},
		{
			name: "Without TxContextKey returns db",
			setup: func() (*executor[int], context.Context, ExecQuery) {
				db := &dbsql.DB{}
				executor := &executor[int]{db: db}
				return executor, context.Background(), db
			},
		},
		{
			name: "With TxContext contains different db returns executor db",
			setup: func() (*executor[int], context.Context, ExecQuery) {
				db := &dbsql.DB{}
				tx := &dbsql.Tx{}
				executor := &executor[int]{db: db, options: options{log: logger.NoopLogger}}
				data := newTxData()
				data.setTx(&dbsql.DB{}, tx)
				ctx := context.WithValue(context.Background(), txContextKey, data)
				return executor, ctx, executor.db
			},
		},
		{
			name: "Context is nil returns executor db",
			setup: func() (*executor[int], context.Context, ExecQuery) {
				db := &dbsql.DB{}
				executor := &executor[int]{db: db}
				return executor, nil, db
			},
		},
	}

	// run tests
	for _, tt := range tests {
		executor, ctx, expectedDb := tt.setup()
		output := executor.getExecQuery(ctx)

		if output != expectedDb {
			t.Errorf("Test: %s \n Expected: %#v but got: %#v", tt.name, expectedDb, output)
		}
	}
}

func TestGetOne(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		withStmt    *mockStmt
		setupDb     func(sqlmock.Sqlmock)
		cacheBundle func() cache.ModelBundle
		expectedErr error
	}{
		{
			name: "Return error in QueryContext",
			ctx:  context.Background(),
			withStmt: func() *mockStmt {
				stmt := new(mockStmt)
				stmt.On("GetStmtWithArgs", sql.SelectOne).Return("query", []interface{}{})
				stmt.On("GetModelPointers", sql.SelectOne, mock.Anything).Return([]interface{}{})
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
				stmt.On("GetStmtWithArgs", sql.SelectOne).Return("query", []interface{}{})
				stmt.On("GetModelPointers", sql.SelectOne, mock.Anything).Return([]interface{}{})
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
					On("GetStmtWithArgs", sql.SelectOne).
					Return("SELECT id, age, name FROM users LIMIT 1", []interface{}{})
				stmt.On("GetModelPointers", sql.SelectOne, mock.Anything).Return([]any{&model.ID, &model.Age, &model.Name})
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
					On("GetStmtWithArgs", sql.SelectOne).
					Return("SELECT id, age, name FROM users LIMIT 1", []interface{}{})
				stmt.On("GetModelPointers", sql.SelectOne, mock.Anything).Return([]any{&model.ID, &model.Age})
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
					On("GetStmtWithArgs", sql.SelectOne).
					Return("query", []interface{}{})
				return stmt
			}(),
			cacheBundle: func() cache.ModelBundle {
				b := &MockModelBundle{}
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
				e.cacheBundle = tt.cacheBundle()
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

func TestTx(t *testing.T) {
	tests := []struct {
		name        string
		withCtx     context.Context
		txOptions   *dbsql.TxOptions
		setupDb     func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name:    "When TxContextKey is not present in context",
			withCtx: context.Background(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
			},
			expectedErr: nil,
		},
		{
			name:    "When TxContext is present in context, but not have a transaction",
			withCtx: context.WithValue(context.Background(), txContextKey, newTxData()),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
			},
			expectedErr: nil,
		},
		{
			name:    "When db.BeginTx returns error",
			withCtx: context.Background(),
			setupDb: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(dbsql.ErrTxDone)
			},
			expectedErr: dbsql.ErrTxDone,
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

			executor := &executor[int]{db: db, options: options{log: logger.NoopLogger}}

			_, _, err = executor.Tx(tt.withCtx, tt.txOptions)
			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Fatalf("Tx() error = %v, wantErr %v", err, tt.expectedErr)
			} else if err == nil && tt.expectedErr != nil {
				t.Fatalf("Tx() expected error %v but got no error.", tt.expectedErr)
			}

			if err := mockDB.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
