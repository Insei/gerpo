package sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

func TestWhereBuilder(t *testing.T) {
	builder := &StringBuilder{
		whereBuilder: &StringWhereBuilder{},
	}

	whereBuilder := builder.WhereBuilder()

	assert.NotNil(t, whereBuilder)
	assert.IsType(t, &StringWhereBuilder{}, whereBuilder)
}

func TestGroupBuilder(t *testing.T) {
	builder := &StringBuilder{
		groupBuilder: &StringGroupBuilder{},
	}

	groupBuilder := builder.GroupBuilder()

	assert.NotNil(t, groupBuilder)
	assert.IsType(t, &StringGroupBuilder{}, groupBuilder)
}

func TestSelectBuilder(t *testing.T) {
	builder := &StringBuilder{
		selectBuilder: &StringSelectBuilder{},
	}

	selectBuilder := builder.SelectBuilder()

	assert.NotNil(t, selectBuilder)
	assert.IsType(t, &StringSelectBuilder{}, selectBuilder)
}

func TestInsertBuilder(t *testing.T) {
	builder := &StringBuilder{
		insertBuilder: &StringInsertBuilder{},
	}

	insertBuilder := builder.InsertBuilder()

	assert.NotNil(t, insertBuilder)
	assert.IsType(t, &StringInsertBuilder{}, insertBuilder)
}

func TestUpdateBuilder(t *testing.T) {
	builder := &StringBuilder{
		updateBuilder: &StringUpdateBuilder{},
	}

	updateBuilder := builder.UpdateBuilder()

	assert.NotNil(t, updateBuilder)
	assert.IsType(t, &StringUpdateBuilder{}, updateBuilder)
}

func TestJoinBuilder(t *testing.T) {
	builder := &StringBuilder{
		joinBuilder: &StringJoinBuilder{},
	}

	joinBuilder := builder.JoinBuilder()

	assert.NotNil(t, joinBuilder)
	assert.IsType(t, &StringJoinBuilder{}, joinBuilder)
}

func TestCountSQL(t *testing.T) {
	testCases := []struct {
		name           string
		whereBuilder   *StringWhereBuilder
		selectBuilder  *StringSelectBuilder
		groupBuilder   *StringGroupBuilder
		joinBuilder    *StringJoinBuilder
		table          string
		expectedSQL    string
		expectedValues []any
	}{
		{
			name: "Count with where clause",
			whereBuilder: &StringWhereBuilder{
				sql: "col1 = ?",
			},
			selectBuilder: &StringSelectBuilder{},
			groupBuilder:  &StringGroupBuilder{},
			joinBuilder:   &StringJoinBuilder{},
			table:         "test_table",
			expectedSQL:   "SELECT count(*) over() AS count FROM test_table WHERE col1 = ? LIMIT 1",
		},
		{
			name: "Count with no where clause",
			whereBuilder: &StringWhereBuilder{
				sql:    "",
				values: []any{},
			},
			selectBuilder: &StringSelectBuilder{},
			groupBuilder:  &StringGroupBuilder{},
			joinBuilder:   &StringJoinBuilder{},
			table:         "test_table",
			expectedSQL:   "SELECT count(*) over() AS count FROM test_table LIMIT 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := &StringBuilder{
				ctx:           context.Background(),
				table:         tc.table,
				whereBuilder:  tc.whereBuilder,
				selectBuilder: tc.selectBuilder,
				groupBuilder:  tc.groupBuilder,
				joinBuilder:   tc.joinBuilder,
			}

			actualSQL := builder.countSQL()
			assert.Equal(t, tc.expectedSQL, actualSQL)
		})
	}
}

func TestSelectSQLFunction(t *testing.T) {
	testCases := []struct {
		name          string
		whereBuilder  *StringWhereBuilder
		selectBuilder *StringSelectBuilder
		groupBuilder  *StringGroupBuilder
		joinBuilder   *StringJoinBuilder
		table         string
		expectedSQL   string
	}{
		{
			name: "Basic select with where and limit",
			whereBuilder: &StringWhereBuilder{
				sql:    "col1 = ?",
				values: []any{"value1"},
			},
			selectBuilder: &StringSelectBuilder{
				limit: 10,
			},
			groupBuilder: &StringGroupBuilder{},
			joinBuilder:  &StringJoinBuilder{},
			table:        "test_table",
			expectedSQL:  "SELECT  FROM test_table WHERE col1 = ? LIMIT 10",
		},
		{
			name: "Minimal select",
			whereBuilder: &StringWhereBuilder{
				sql:    "",
				values: []any{},
			},
			selectBuilder: &StringSelectBuilder{},
			groupBuilder:  &StringGroupBuilder{},
			joinBuilder:   &StringJoinBuilder{},
			table:         "test_table",
			expectedSQL:   "SELECT  FROM test_table",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := &StringBuilder{
				ctx:           context.Background(),
				table:         tc.table,
				whereBuilder:  tc.whereBuilder,
				selectBuilder: tc.selectBuilder,
				groupBuilder:  tc.groupBuilder,
				joinBuilder:   tc.joinBuilder,
			}

			actualSQL := builder.selectSQL()
			assert.Equal(t, tc.expectedSQL, actualSQL)
		})
	}
}

func TestInsertSQL(t *testing.T) {
	testCases := []struct {
		name           string
		builder        *StringBuilder
		expectedResult string
	}{
		{
			name: "Insert with table name",
			builder: &StringBuilder{
				table:         "test_table",
				insertBuilder: &StringInsertBuilder{},
			},
			expectedResult: "INSERT INTO test_table ",
		},
		{
			name: "Insert with different table name",
			builder: &StringBuilder{
				table:         "another_table",
				insertBuilder: &StringInsertBuilder{},
			},
			expectedResult: "INSERT INTO another_table ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.builder.insertSQL()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestUpdateSQL(t *testing.T) {
	testCases := []struct {
		name        string
		builder     *StringBuilder
		expectedSQL string
	}{
		{
			name: "Update with where clause",
			builder: &StringBuilder{
				table: "test_table",
				updateBuilder: &StringUpdateBuilder{
					columns: []types.Column{
						&testColumn{sql: "col1"},
					},
				},
				whereBuilder: &StringWhereBuilder{
					sql: "col1 = ?",
				},
			},
			expectedSQL: "UPDATE test_table SET col1 = ? WHERE col1 = ?",
		},
		{
			name: "Update without where clause",
			builder: &StringBuilder{
				table: "test_table",
				updateBuilder: &StringUpdateBuilder{
					columns: []types.Column{
						&testColumn{sql: "col1"},
					},
				},
				whereBuilder: &StringWhereBuilder{
					sql: "",
				},
			},
			expectedSQL: "UPDATE test_table SET col1 = ?",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualSQL := tc.builder.updateSQL()
			assert.Equal(t, tc.expectedSQL, actualSQL)
		})
	}
}

func TestDeleteSQL(t *testing.T) {
	testCases := []struct {
		name        string
		builder     *StringBuilder
		expectedSQL string
		expectPanic bool
	}{
		{
			name: "Delete with where clause and join",
			builder: &StringBuilder{
				ctx:   context.Background(),
				table: "test_table",
				joinBuilder: &StringJoinBuilder{
					joins: []func(ctx context.Context) string{
						func(ctx context.Context) string {
							return "JOIN other_table ON test_table.id = other_table.id"
						},
					},
				},
				whereBuilder: &StringWhereBuilder{sql: "col1 = ?"},
			},
			expectedSQL: "DELETE FROM test_table  JOIN other_table ON test_table.id = other_table.id WHERE col1 = ?",
			expectPanic: false,
		},
		{
			name: "Delete with where clause without join",
			builder: &StringBuilder{
				ctx:          context.Background(),
				table:        "test_table",
				joinBuilder:  &StringJoinBuilder{},
				whereBuilder: &StringWhereBuilder{sql: "col1 = ?"},
			},
			expectedSQL: "DELETE FROM test_table WHERE col1 = ?",
			expectPanic: false,
		},
		{
			name: "Delete without where clause",
			builder: &StringBuilder{
				ctx:         context.Background(),
				table:       "test_table",
				joinBuilder: &StringJoinBuilder{},
				whereBuilder: &StringWhereBuilder{
					sql: "",
				},
			},
			expectedSQL: "",
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				assert.Panics(t, func() { tc.builder.deleteSQL() })
			} else {
				actualSQL := tc.builder.deleteSQL()
				assert.Equal(t, tc.expectedSQL, actualSQL)
			}
		})
	}
}

func TestNewStringBuilder(t *testing.T) {
	ctx := context.Background()
	table := "test_table"
	columns := &types.ColumnsStorage{}

	builder := NewStringBuilder(ctx, table, columns)

	assert.Equal(t, table, builder.table)
	assert.Equal(t, ctx, builder.ctx)
	assert.NotNil(t, builder.whereBuilder)
	assert.NotNil(t, builder.selectBuilder)
	assert.NotNil(t, builder.groupBuilder)
	assert.NotNil(t, builder.insertBuilder)
	assert.NotNil(t, builder.updateBuilder)
	assert.NotNil(t, builder.joinBuilder)

	assert.Equal(t, []types.Column(nil), builder.selectBuilder.columns)
	assert.Equal(t, []types.Column(nil), builder.insertBuilder.columns)
	assert.Equal(t, []types.Column(nil), builder.updateBuilder.columns)
}

func TestStringBuilderFactory_New(t *testing.T) {
	ctx := context.Background()
	table := "test_table"
	columns := &types.ColumnsStorage{}

	factory := StringBuilderFactory(func(ctx context.Context) *StringBuilder {
		return NewStringBuilder(ctx, table, columns)
	})

	builder := factory.New(ctx)

	assert.NotNil(t, builder)
	assert.Equal(t, table, builder.table)
	assert.Equal(t, ctx, builder.ctx)
}

func TestNewStringBuilderFactory(t *testing.T) {
	table := "test_table"
	columns := &types.ColumnsStorage{}

	builder := NewStringBuilderFactory(table, columns)

	assert.NotNil(t, builder)
}
