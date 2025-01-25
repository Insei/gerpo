package linq_test

import (
	"testing"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/types"
)

// mockExcludeApplier implements the ExcludeApplier interface for testing purposes.
type mockExcludeApplier struct {
	cols types.ExecutionColumns
}

func (m *mockExcludeApplier) Columns() types.ExecutionColumns {
	return m.cols
}

// mockExecColumns is a simplified implementation for testing that focuses on
// calls to GetByFieldPtr and Exclude.
type mockExecColumns struct {
	types.ExecutionColumns
	getByFieldPtrFunc func(model, fieldPtr any) types.Column
	excludeFunc       func(cols ...types.Column)
}

func (m *mockExecColumns) GetAll() []types.Column       { return nil }
func (m *mockExecColumns) Exclude(cols ...types.Column) { m.excludeFunc(cols...) }
func (m *mockExecColumns) GetByFieldPtr(model any, fieldPtr any) types.Column {
	return m.getByFieldPtrFunc(model, fieldPtr)
}
func (m *mockExecColumns) GetModelPointers(model any) []any { return nil }
func (m *mockExecColumns) GetModelValues(model any) []any   { return nil }

// mockColumn models the behavior of a column in a simplified way.
type mockColumn struct {
	types.Column
	name string
}

func TestExcludeBuilder(t *testing.T) {
	testScenarios := []struct {
		name       string
		fieldPtrs  []any
		setupMocks func() *mockExecColumns
		expectExcl int
	}{
		{
			name:      "No fields to exclude",
			fieldPtrs: []any{},
			setupMocks: func() *mockExecColumns {
				return &mockExecColumns{
					getByFieldPtrFunc: func(model, fieldPtr any) types.Column {
						return &mockColumn{name: "unused"}
					},
					excludeFunc: func(cols ...types.Column) {},
				}
			},
			expectExcl: 0,
		},
		{
			name:      "One field to exclude",
			fieldPtrs: []any{"fieldPtrA"},
			setupMocks: func() *mockExecColumns {
				return &mockExecColumns{
					getByFieldPtrFunc: func(model, fieldPtr any) types.Column {
						return &mockColumn{name: fieldPtr.(string)}
					},
					excludeFunc: func(cols ...types.Column) {},
				}
			},
			expectExcl: 1,
		},
		{
			name:      "Multiple fields to exclude",
			fieldPtrs: []any{"fieldPtrA", "fieldPtrB", "fieldPtrC"},
			setupMocks: func() *mockExecColumns {
				return &mockExecColumns{
					getByFieldPtrFunc: func(model, fieldPtr any) types.Column {
						return &mockColumn{name: fieldPtr.(string)}
					},
					excludeFunc: func(cols ...types.Column) {},
				}
			},
			expectExcl: 3,
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			excludeBuilder := linq.NewExcludeBuilder("myModel")
			testColumns := scenario.setupMocks()

			excludeCallsCount := 0
			originalExcludeFunc := testColumns.excludeFunc
			testColumns.excludeFunc = func(cols ...types.Column) {
				excludeCallsCount++
				originalExcludeFunc(cols...)
			}

			excludeBuilder.Exclude(scenario.fieldPtrs...)
			testApplier := &mockExcludeApplier{cols: testColumns}
			excludeBuilder.Apply(testApplier)

			if excludeCallsCount != scenario.expectExcl {
				t.Errorf("Expected %d call(s) to Exclude, got %d",
					scenario.expectExcl, excludeCallsCount)
			}
		})
	}
}
