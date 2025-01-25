package linq

import (
	"context"
	"testing"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
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

type mockGroupApplier struct {
	columnsStorage types.ColumnsStorage
	group          sqlpart.Group
}

func (m *mockGroupApplier) ColumnsStorage() types.ColumnsStorage {
	return m.columnsStorage
}

func (m *mockGroupApplier) Group() sqlpart.Group {
	return m.group
}

type mockColumnsStorage struct {
	types.ColumnsStorage
	columns map[any]types.Column
}

func (m *mockColumnsStorage) GetByFieldPtr(model any, fieldPtr any) (types.Column, error) {
	if col, ok := m.columns[fieldPtr]; ok {
		return col, nil
	}
	return nil, nil
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
			builder.Apply(mockApplier)

			actualGroupings := make([]string, len(mockG.groupings))
			for i, col := range mockG.groupings {
				name, _ := col.Name()
				actualGroupings[i] = name
			}

			assert.Equal(t, tc.expectedGroupings, actualGroupings)
		})
	}
}
