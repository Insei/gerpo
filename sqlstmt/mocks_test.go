package sqlstmt

import (
	"context"

	"github.com/insei/gerpo/types"
)

type mockColumn struct {
	types.Column
	name          string
	allowedAction bool
	hasName       bool
}

func (m *mockColumn) IsAllowedAction(action types.SQLAction) bool {
	return m.allowedAction
}

func (m *mockColumn) ToSQL(ctx context.Context) string {
	return m.name
}

func (m *mockColumn) Name() (string, bool) {
	return m.name, m.hasName
}

type mockExecutionColumns struct {
	types.ExecutionColumns
	columns     []types.Column
	modelValues []any
}

func newMockExecutionColumns(columns []types.Column) *mockExecutionColumns {
	return &mockExecutionColumns{
		columns: columns,
	}
}

// GetAll возвращает все смоделированные колонки
func (m *mockExecutionColumns) GetAll() []types.Column {
	return m.columns
}

func (m *mockExecutionColumns) GetModelValues(model any) []any {
	return m.modelValues
}

type mockStorage struct {
	types.ColumnsStorage
	executionColumns []types.Column
}

func newMockStorage(executionColumns []types.Column) *mockStorage {
	return &mockStorage{
		executionColumns: executionColumns,
	}
}

func (m *mockStorage) NewExecutionColumns(ctx context.Context, action types.SQLAction) types.ExecutionColumns {
	return newMockExecutionColumns(m.executionColumns)
}
