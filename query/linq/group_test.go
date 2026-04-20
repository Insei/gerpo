package linq

import (
	"context"
	"fmt"
	"testing"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockColumn struct {
	types.Column
	name          string
	allowedAction bool
	hasName       bool
	aggregate     bool
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

func (m *mockColumn) IsAggregate() bool { return m.aggregate }

func (m *mockColumn) HasFilterOverride(_ types.Operation) bool { return false }

type mockGroupApplier struct {
	columnsStorage types.ColumnsStorage
	group          sqlpart.Group
	cols           types.ExecutionColumns // optional, set when the applier needs to look like a SELECT statement
}

func (m *mockGroupApplier) ColumnsStorage() types.ColumnsStorage {
	return m.columnsStorage
}

func (m *mockGroupApplier) Group() sqlpart.Group {
	return m.group
}

// Columns is satisfied conditionally: nil cols means this applier does NOT
// look like a SELECT-style statement and the auto GROUP BY logic must skip it.
// Tests assign cols when they want to exercise the auto path.
func (m *mockGroupApplier) Columns() types.ExecutionColumns { return m.cols }

type mockExecCols struct {
	types.ExecutionColumns
	all []types.Column
}

func (m *mockExecCols) GetAll() []types.Column { return m.all }

type mockColumnsStorage struct {
	types.ColumnsStorage
	columns map[any]types.Column
}

func (m *mockColumnsStorage) GetByFieldPtr(model any, fieldPtr any) (types.Column, error) {
	if col, ok := m.columns[fieldPtr]; ok {
		return col, nil
	}
	return nil, fmt.Errorf("column not found")
}

type mockGroup struct {
	sqlpart.Group
	groupings []types.Column
}

func (m *mockGroup) GroupBy(cols ...types.Column) {
	m.groupings = append(m.groupings, cols...)
}

func TestGroupBuilder_GroupBy(t *testing.T) {
	type testModel struct {
		ID   int
		Name string
	}
	tm := &testModel{}
	testCases := []struct {
		name              string
		model             *testModel
		fields            []any
		expectedGroupings []string
	}{
		{
			name:              "GroupBy one field",
			model:             tm,
			fields:            []any{&tm.ID},
			expectedGroupings: []string{"id"},
		},
		{
			name:              "GroupBy multiple fields",
			model:             tm,
			fields:            []any{&tm.ID, &tm.Name},
			expectedGroupings: []string{"id", "name"},
		},
		{
			name:              "GroupBy no fields",
			model:             tm,
			fields:            []any{},
			expectedGroupings: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			columns := make(map[any]types.Column)
			for _, field := range tc.fields {
				fieldName := ""
				switch field.(type) {
				case *int:
					fieldName = "id"
				case *string:
					fieldName = "name"
				}
				columns[field] = &mockColumn{
					name: fieldName,
				}
			}

			mockCS := &mockColumnsStorage{columns: columns}
			mockG := &mockGroup{}
			mockApplier := &mockGroupApplier{columnsStorage: mockCS, group: mockG}

			builder := NewGroupBuilder(tc.model)
			builder.GroupBy(tc.fields...)
			err := builder.Apply(mockApplier)
			assert.NoError(t, err)

			actualGroupings := make([]string, len(mockG.groupings))
			for i, col := range mockG.groupings {
				name, _ := col.Name()
				actualGroupings[i] = name
			}

			assert.Equal(t, tc.expectedGroupings, actualGroupings)
		})
	}
}

// TestGroupBuilder_AutoFill_FromAggregate — when at least one SELECT column is
// an aggregate and the user didn't configure GroupBy, every non-aggregate SELECT
// column is added to GROUP BY automatically.
func TestGroupBuilder_AutoFill_FromAggregate(t *testing.T) {
	id := &mockColumn{name: "id"}
	name := &mockColumn{name: "name"}
	postCount := &mockColumn{name: "COALESCE(COUNT(posts.id), 0)", aggregate: true}

	mockG := &mockGroup{}
	applier := &mockGroupApplier{
		group: mockG,
		cols:  &mockExecCols{all: []types.Column{id, name, postCount}},
	}

	type m struct{}
	require.NoError(t, NewGroupBuilder(&m{}).Apply(applier))

	got := make([]string, len(mockG.groupings))
	for i, c := range mockG.groupings {
		got[i] = c.(*mockColumn).name
	}
	assert.Equal(t, []string{"id", "name"}, got,
		"every non-aggregate SELECT column must be auto-added; aggregate is skipped")
}

// TestGroupBuilder_AutoFill_NoAggregate_Noop — without any aggregate column the
// builder must not add a GROUP BY (regression: empty-fieldPtrs + no aggregate
// used to be a no-op and that contract still holds).
func TestGroupBuilder_AutoFill_NoAggregate_Noop(t *testing.T) {
	id := &mockColumn{name: "id"}
	name := &mockColumn{name: "name"}

	mockG := &mockGroup{}
	applier := &mockGroupApplier{
		group: mockG,
		cols:  &mockExecCols{all: []types.Column{id, name}},
	}

	type m struct{}
	require.NoError(t, NewGroupBuilder(&m{}).Apply(applier))
	assert.Empty(t, mockG.groupings,
		"no aggregate present → no GROUP BY")
}

// TestGroupBuilder_ManualGroupBy_OverridesAuto — if the user already called
// GroupBy(...), the auto logic must stay out of the way: only the user's
// fields end up in GROUP BY, even when an aggregate is present in SELECT.
func TestGroupBuilder_ManualGroupBy_OverridesAuto(t *testing.T) {
	type model struct {
		ID   int
		Name string
	}
	tm := &model{}

	idCol := &mockColumn{name: "id"}
	nameCol := &mockColumn{name: "name"}
	postCount := &mockColumn{name: "count", aggregate: true}

	storage := &mockColumnsStorage{columns: map[any]types.Column{&tm.ID: idCol}}
	mockG := &mockGroup{}
	applier := &mockGroupApplier{
		columnsStorage: storage,
		group:          mockG,
		cols:           &mockExecCols{all: []types.Column{idCol, nameCol, postCount}},
	}

	b := NewGroupBuilder(tm)
	b.GroupBy(&tm.ID) // explicit user choice — auto must not augment.
	require.NoError(t, b.Apply(applier))

	got := make([]string, len(mockG.groupings))
	for i, c := range mockG.groupings {
		got[i] = c.(*mockColumn).name
	}
	assert.Equal(t, []string{"id"}, got,
		"manual GroupBy must not be augmented by auto-fill")
}

// TestGroupBuilder_NotSelectableApplier_AutoSkipped — Update / Delete style
// statements implement neither Group() nor Columns(); the auto logic must
// handle the case where the applier exposes Group() but not Columns()
// gracefully (Persistent.Apply already guards Update/Delete via type-assert
// on GroupApplier, but the contract should be defensive).
func TestGroupBuilder_NotSelectableApplier_AutoSkipped(t *testing.T) {
	mockG := &mockGroup{}
	// applier without cols — Columns() returns nil, auto-fill must skip.
	applier := &mockGroupApplier{group: mockG}

	type m struct{}
	require.NoError(t, NewGroupBuilder(&m{}).Apply(applier))
	assert.Empty(t, mockG.groupings)
}
