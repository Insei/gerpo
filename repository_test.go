package gerpo

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/require"
)

type MockExecutor[TModel any] struct {
	executor.Executor[TModel]
	DeleteFunc      func(ctx context.Context, stmt executor.CountStmt) (int64, error)
	UpdateFunc      func(ctx context.Context, stmt executor.Stmt, model *TModel) (int64, error)
	InsertOneFunc   func(ctx context.Context, stmt executor.Stmt, model *TModel) error
	CountFunc       func(ctx context.Context, stmt executor.CountStmt) (uint64, error)
	GetMultipleFunc func(ctx context.Context, stmt executor.Stmt) ([]*TModel, error)
	GetOneFunc      func(ctx context.Context, stmt executor.Stmt) (*TModel, error)
}

func (m *MockExecutor[TModel]) Delete(ctx context.Context, stmt executor.CountStmt) (int64, error) {
	return m.DeleteFunc(ctx, stmt)
}

func (m *MockExecutor[TModel]) Update(ctx context.Context, stmt executor.Stmt, model *TModel) (int64, error) {
	return m.UpdateFunc(ctx, stmt, model)
}

func (m *MockExecutor[TModel]) InsertOne(ctx context.Context, stmt executor.Stmt, model *TModel) error {
	return m.InsertOneFunc(ctx, stmt, model)
}

func (m *MockExecutor[TModel]) Count(ctx context.Context, stmt executor.CountStmt) (uint64, error) {
	return m.CountFunc(ctx, stmt)
}

func (m *MockExecutor[TModel]) GetMultiple(ctx context.Context, stmt executor.Stmt) ([]*TModel, error) {
	return m.GetMultipleFunc(ctx, stmt)
}

func (m *MockExecutor[TModel]) GetOne(ctx context.Context, stmt executor.Stmt) (*TModel, error) {
	return m.GetOneFunc(ctx, stmt)
}

func TestRepository_GetFirst(t *testing.T) {
	type model struct {
		ID    int
		Name  string
		Email string
	}

	ErrTest := errors.New("test error")

	tests := []struct {
		name        string
		executor    executor.Executor[model]
		qFns        []func(m *model, h query.GetFirstHelper[model])
		expected    *model
		expectedErr error
		setupSelect func(ctx context.Context, model *model)
	}{
		{
			name: "Successful retrieval",
			executor: &MockExecutor[model]{
				GetOneFunc: func(ctx context.Context, stmt executor.Stmt) (*model, error) {
					return &model{
						ID:    1,
						Name:  "Alice",
						Email: "alice@example.com",
					}, nil
				},
			},
			expected: &model{
				ID:    1,
				Name:  "Alice",
				Email: "alice@example.com",
			},
			expectedErr: nil,
		},
		{
			name: "No matching records",
			executor: &MockExecutor[model]{
				GetOneFunc: func(ctx context.Context, stmt executor.Stmt) (*model, error) {
					return nil, sql.ErrNoRows
				},
			},
			expected:    nil,
			expectedErr: ErrNotFound,
		},
		{
			name: "Database error",
			executor: &MockExecutor[model]{
				GetOneFunc: func(ctx context.Context, stmt executor.Stmt) (*model, error) {
					return nil, ErrTest
				},
			},
			expected:    nil,
			expectedErr: ErrTest,
		},
		{
			name: "Select modifies returned model",
			executor: &MockExecutor[model]{
				GetOneFunc: func(ctx context.Context, stmt executor.Stmt) (*model, error) {
					return &model{
						ID:    2,
						Name:  "Bob",
						Email: "bob@example.com",
					}, nil
				},
			},
			expected: &model{
				ID:    2,
				Name:  "Modified Bob",
				Email: "bob@example.com",
			},
			expectedErr: nil,
			setupSelect: func(ctx context.Context, model *model) {
				model.Name = "Modified " + model.Name
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := New[model](&sql.DB{}, "test_table", func(m *model, builder *ColumnBuilder[model]) {
				builder.Field(&m.ID).Column()
				builder.Field(&m.Name).Column()
				builder.Field(&m.Email).Column()
			})
			if err != nil {
				t.Fatalf("failed to create repository: %v", err)
			}
			repoCasted := repo.(*repository[model])
			repoCasted.executor = tt.executor
			if tt.setupSelect != nil {
				repoCasted.afterSelect = func(ctx context.Context, models []*model) {
					for _, m := range models {
						tt.setupSelect(ctx, m)
					}
				}
			}

			ctx := context.Background()
			result, err := repo.GetFirst(ctx, tt.qFns...)

			require.Equal(t, tt.expected, result)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_GetList(t *testing.T) {
	type model struct {
		ID    int
		Name  string
		Email string
	}

	ErrTest := errors.New("test error")

	tests := []struct {
		name         string
		executor     executor.Executor[model]
		qFns         []func(m *model, h query.GetListHelper[model])
		expectedList []*model
		expectedErr  error
		afterSelect  func(ctx context.Context, models []*model)
	}{
		{
			name: "Successful retrieval",
			executor: &MockExecutor[model]{
				GetMultipleFunc: func(ctx context.Context, stmt executor.Stmt) ([]*model, error) {
					return []*model{
						{ID: 1, Name: "Alice", Email: "alice@example.com"},
						{ID: 2, Name: "Bob", Email: "bob@example.com"},
					}, nil
				},
			},
			expectedList: []*model{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			},
			expectedErr: nil,
		},
		{
			name: "No matching records",
			executor: &MockExecutor[model]{
				GetMultipleFunc: func(ctx context.Context, stmt executor.Stmt) ([]*model, error) {
					return []*model{}, nil
				},
			},
			expectedList: []*model{},
			expectedErr:  nil,
		},
		{
			name: "Database error",
			executor: &MockExecutor[model]{
				GetMultipleFunc: func(ctx context.Context, stmt executor.Stmt) ([]*model, error) {
					return nil, ErrTest
				},
			},
			expectedList: nil,
			expectedErr:  ErrTest,
		},
		{
			name: "AfterSelect modifies models",
			executor: &MockExecutor[model]{
				GetMultipleFunc: func(ctx context.Context, stmt executor.Stmt) ([]*model, error) {
					return []*model{
						{ID: 1, Name: "Alice", Email: "alice@example.com"},
					}, nil
				},
			},
			expectedList: []*model{
				{ID: 1, Name: "Modified Alice", Email: "alice@example.com"},
			},
			expectedErr: nil,
			afterSelect: func(ctx context.Context, models []*model) {
				for _, m := range models {
					m.Name = "Modified " + m.Name
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := New[model](&sql.DB{}, "test_table", func(m *model, builder *ColumnBuilder[model]) {
				builder.Field(&m.ID).Column()
				builder.Field(&m.Name).Column()
				builder.Field(&m.Email).Column()
			})
			if err != nil {
				t.Fatalf("failed to create repository: %v", err)
			}
			repoCasted := repo.(*repository[model])
			repoCasted.executor = tt.executor
			if tt.afterSelect != nil {
				repoCasted.afterSelect = tt.afterSelect
			}

			ctx := context.Background()
			list, err := repo.GetList(ctx, tt.qFns...)

			require.Equal(t, tt.expectedList, list)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_Count(t *testing.T) {
	type model struct {
		ID    int
		Name  string
		Email string
	}

	ErrTest := errors.New("test error")

	tests := []struct {
		name        string
		executor    executor.Executor[model]
		qFns        []func(m *model, h query.CountHelper[model])
		expectedCnt uint64
		expectedErr error
	}{
		{
			name: "Successful count",
			executor: &MockExecutor[model]{
				CountFunc: func(ctx context.Context, stmt executor.CountStmt) (uint64, error) {
					return 42, nil
				},
			},
			expectedCnt: 42,
			expectedErr: nil,
		},
		{
			name: "No matching records",
			executor: &MockExecutor[model]{
				CountFunc: func(ctx context.Context, stmt executor.CountStmt) (uint64, error) {
					return 0, nil
				},
			},
			expectedCnt: 0,
			expectedErr: nil,
		},
		{
			name: "Count error",
			executor: &MockExecutor[model]{
				CountFunc: func(ctx context.Context, stmt executor.CountStmt) (uint64, error) {
					return 0, ErrTest
				},
			},
			expectedCnt: 0,
			expectedErr: ErrTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := New[model](&sql.DB{}, "test_table", func(m *model, builder *ColumnBuilder[model]) {
				builder.Field(&m.ID).Column()
				builder.Field(&m.Name).Column()
				builder.Field(&m.Email).Column()
			})
			if err != nil {
				t.Fatalf("failed to create repository: %v", err)
			}
			repoCasted := repo.(*repository[model])
			repoCasted.executor = tt.executor

			ctx := context.Background()
			cnt, err := repo.Count(ctx, tt.qFns...)

			if cnt != tt.expectedCnt {
				t.Errorf("expected count %d, got %d", tt.expectedCnt, cnt)
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_Delete(t *testing.T) {
	type model struct {
		ID int
	}

	ErrTest := errors.New("test error")

	tests := []struct {
		name        string
		executor    executor.Executor[model]
		qFns        []func(m *model, h query.DeleteHelper[model])
		expectedCnt int64
		expectedErr error
	}{
		{
			name: "Successful delete",
			executor: &MockExecutor[model]{
				DeleteFunc: func(ctx context.Context, stmt executor.CountStmt) (int64, error) {
					return 1, nil
				},
			},
			expectedCnt: 1,
			expectedErr: nil,
		},
		{
			name: "Nothing to delete",
			executor: &MockExecutor[model]{
				DeleteFunc: func(ctx context.Context, stmt executor.CountStmt) (int64, error) {
					return 0, nil
				},
			},
			expectedCnt: 0,
			expectedErr: ErrNotFound,
		},
		{
			name: "Delete error",
			executor: &MockExecutor[model]{
				DeleteFunc: func(ctx context.Context, stmt executor.CountStmt) (int64, error) {
					return 0, ErrTest
				},
			},
			expectedCnt: 0,
			expectedErr: ErrTest,
		},
	}

	type stor struct {
		types.ColumnsStorage
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := New[model](&sql.DB{}, "test_table", func(m *model, builder *ColumnBuilder[model]) {
				builder.Field(&m.ID).Column()
			})
			repoCasted := repo.(*repository[model])
			repoCasted.executor = tt.executor

			ctx := context.Background()
			cnt, err := repo.Delete(ctx, tt.qFns...)

			if cnt != tt.expectedCnt {
				t.Errorf("expected count %d, got %d", tt.expectedCnt, cnt)
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_Insert(t *testing.T) {
	type model struct {
		ID    int
		Name  string
		Email string
	}

	ErrTest := errors.New("test error")

	tests := []struct {
		name         string
		executor     executor.Executor[model]
		model        *model
		qFns         []func(m *model, h query.InsertHelper[model])
		expectedErr  error
		beforeInsert func(ctx context.Context, model *model)
		afterInsert  func(ctx context.Context, model *model)
	}{
		{
			name: "Successful insert",
			executor: &MockExecutor[model]{
				InsertOneFunc: func(ctx context.Context, stmt executor.Stmt, model *model) error {
					model.ID = 1 // Simulate assigning an ID from DB
					return nil
				},
			},
			model:       &model{Name: "John Doe", Email: "john@example.com"},
			expectedErr: nil,
			beforeInsert: func(ctx context.Context, model *model) {
				model.Name = "BeforeInsert Name"
			},
			afterInsert: func(ctx context.Context, model *model) {
				model.Name = "AfterInsert Name"
			},
		},
		{
			name: "Insert error",
			executor: &MockExecutor[model]{
				InsertOneFunc: func(ctx context.Context, stmt executor.Stmt, model *model) error {
					return ErrTest
				},
			},
			model:       &model{Name: "Error Case", Email: "error@example.com"},
			expectedErr: ErrTest,
		},
		{
			name: "BeforeInsert modifies model",
			executor: &MockExecutor[model]{
				InsertOneFunc: func(ctx context.Context, stmt executor.Stmt, model *model) error {
					if model.Name != "Modified Name" {
						return errors.New("beforeInsert not applied")
					}
					return nil
				},
			},
			model:       &model{Name: "Original Name", Email: "test@example.com"},
			expectedErr: nil,
			beforeInsert: func(ctx context.Context, model *model) {
				model.Name = "Modified Name"
			},
		},
		{
			name: "AfterInsert modifies model",
			executor: &MockExecutor[model]{
				InsertOneFunc: func(ctx context.Context, stmt executor.Stmt, model *model) error {
					return nil
				},
			},
			model:       &model{Name: "After Insert Test", Email: "after_test@example.com"},
			expectedErr: nil,
			afterInsert: func(ctx context.Context, model *model) {
				model.ID = 42
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := New[model](&sql.DB{}, "test_table", func(m *model, builder *ColumnBuilder[model]) {
				builder.Field(&m.ID).Column()
				builder.Field(&m.Name).Column()
				builder.Field(&m.Email).Column()
			})
			if err != nil {
				t.Fatalf("failed to create repository: %v", err)
			}
			repoCasted := repo.(*repository[model])
			repoCasted.executor = tt.executor
			if tt.beforeInsert != nil {
				repoCasted.beforeInsert = tt.beforeInsert
			}
			if tt.afterInsert != nil {
				repoCasted.afterInsert = tt.afterInsert
			}

			ctx := context.Background()
			err = repo.Insert(ctx, tt.model, tt.qFns...)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_Update(t *testing.T) {
	type model struct {
		ID    int
		Name  string
		Email string
	}

	ErrTest := errors.New("test error")

	tests := []struct {
		name         string
		executor     executor.Executor[model]
		model        *model
		qFns         []func(m *model, h query.UpdateHelper[model])
		expectedErr  error
		beforeUpdate func(ctx context.Context, model *model)
		afterUpdate  func(ctx context.Context, model *model)
	}{
		{
			name: "Successful update",
			executor: &MockExecutor[model]{
				UpdateFunc: func(ctx context.Context, stmt executor.Stmt, model *model) (int64, error) {
					return 1, nil
				},
			},
			model:       &model{ID: 1, Name: "Updated Name", Email: "updated@example.com"},
			expectedErr: nil,
			beforeUpdate: func(ctx context.Context, model *model) {
				model.Name = "BeforeUpdate Name"
			},
			afterUpdate: func(ctx context.Context, model *model) {
				model.Name = "AfterUpdate Name"
			},
		},
		{
			name: "Nothing to update",
			executor: &MockExecutor[model]{
				UpdateFunc: func(ctx context.Context, stmt executor.Stmt, model *model) (int64, error) {
					return 0, nil
				},
			},
			model:       &model{ID: 2, Name: "NonExistent", Email: "noone@example.com"},
			expectedErr: ErrNotFound,
		},
		{
			name: "Update error",
			executor: &MockExecutor[model]{
				UpdateFunc: func(ctx context.Context, stmt executor.Stmt, model *model) (int64, error) {
					return 0, ErrTest
				},
			},
			model:       &model{ID: 3, Name: "Error Testing", Email: "error@example.com"},
			expectedErr: ErrTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := New[model](&sql.DB{}, "test_table", func(m *model, builder *ColumnBuilder[model]) {
				builder.Field(&m.ID).Column()
			})
			repoCasted := repo.(*repository[model])
			repoCasted.executor = tt.executor

			ctx := context.Background()
			err = repo.Update(ctx, tt.model, tt.qFns...)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
