package linq

import (
	"context"
	"testing"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
)

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

func (m *mockStorage) GetByFieldPtr(model any, fieldPtr any) (types.Column, error) {
	return m.executionColumns[0], nil
}

func (m *mockStorage) NewExecutionColumns(ctx context.Context, action types.SQLAction) types.ExecutionColumns {
	return newMockExecutionColumns(m.executionColumns)
}

type mockOrder struct {
	sqlpart.Order
	order []string
}

func (m *mockOrder) OrderByColumn(column types.Column, direction types.OrderDirection) {
	m.order = append(m.order, column.ToSQL(context.Background())+" "+string(direction))
}

type mockOrderApplier struct {
	storage types.ColumnsStorage
	order   sqlpart.Order
}

func (m *mockOrderApplier) ColumnsStorage() types.ColumnsStorage {
	return m.storage
}

func (m *mockOrderApplier) Order() sqlpart.Order {
	return m.order
}

func TestOrderBuilder_Column(t *testing.T) {
	testCases := []struct {
		name          string
		column        types.Column
		direction     types.OrderDirection
		expectedOrder []string
	}{
		{
			name: "Order_by_column_ASC",
			column: &mockColumn{
				name: "created_at",
			},
			direction: types.OrderDirectionASC,
			expectedOrder: []string{
				"created_at ASC",
			},
		},
		{
			name: "Order_by_column_DESC",
			column: &mockColumn{
				name: "updated_at",
			},
			direction: types.OrderDirectionDESC,

			expectedOrder: []string{
				"updated_at DESC",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			builder := NewOrderBuilder(nil)
			switch tc.direction {
			case types.OrderDirectionASC:
				builder.Column(tc.column).ASC()
			case types.OrderDirectionDESC:
				builder.Column(tc.column).DESC()
			}

			mockOrder := &mockOrder{}
			applier := &mockOrderApplier{
				order: mockOrder,
				storage: &mockStorage{
					executionColumns: []types.Column{tc.column},
				},
			}
			builder.Apply(applier)

			assert.Equal(t, tc.expectedOrder, mockOrder.order)
		})
	}
}

func TestOrderBuilder_Field(t *testing.T) {

	type model struct {
		CreatedAt string
		UpdatedAt string
	}
	modelInstance := model{}

	testCases := []struct {
		name          string
		fieldPtr      any
		direction     types.OrderDirection
		columns       []types.Column
		expectedOrder []string
	}{
		{
			name:     "Order_by_field_ASC",
			fieldPtr: &modelInstance.CreatedAt,
			columns: []types.Column{
				&mockColumn{
					name:    "created_at",
					hasName: true,
				},
			},
			direction: types.OrderDirectionASC,
			expectedOrder: []string{
				"created_at ASC",
			},
		},
		{
			name:     "Order_by_field_DESC",
			fieldPtr: &modelInstance.UpdatedAt,
			columns: []types.Column{
				&mockColumn{
					name:    "updated_at",
					hasName: true,
				},
			},
			direction: types.OrderDirectionDESC,
			expectedOrder: []string{
				"updated_at DESC",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewOrderBuilder(&modelInstance)

			mockStorage := newMockStorage(tc.columns)

			switch tc.direction {
			case types.OrderDirectionASC:
				builder.Field(tc.fieldPtr).ASC()
			case types.OrderDirectionDESC:
				builder.Field(tc.fieldPtr).DESC()
			}
			mockOrder := &mockOrder{}

			applier := &mockOrderApplier{
				storage: mockStorage,
				order:   mockOrder,
			}

			builder.Apply(applier)
			assert.Equal(t, tc.expectedOrder, mockOrder.order)

		})
	}
}
