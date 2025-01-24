package sqlpart

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringJoinBuilder(t *testing.T) {
	testCases := []struct {
		name         string
		initialJoins []func(ctx context.Context) string
		newJoin      func(ctx context.Context) string
	}{
		{
			name:         "Add single join",
			initialJoins: []func(ctx context.Context) string{},
			newJoin: func(ctx context.Context) string {
				return "INNER JOIN table1 ON table1.id = table2.table1_id"
			},
		},
		{
			name: "Add multiple joins",
			initialJoins: []func(ctx context.Context) string{
				func(ctx context.Context) string {
					return "INNER JOIN table1 ON table1.id = table2.table1_id"
				},
			},
			newJoin: func(ctx context.Context) string {
				return "LEFT JOIN table3 ON table3.id = table2.table3_id"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := &JoinBuilder{
				joins: tc.initialJoins,
			}

			builder.JOIN(tc.newJoin)
		})
	}
}

func TestStringJoinBuilderSQL(t *testing.T) {
	testCases := []struct {
		name         string
		initialJoins []func(ctx context.Context) string
		expectedSQL  string
	}{
		{
			name:         "No joins",
			initialJoins: []func(ctx context.Context) string{},
			expectedSQL:  "",
		},
		{
			name: "Single join",
			initialJoins: []func(ctx context.Context) string{
				func(ctx context.Context) string {
					return "INNER JOIN table1 ON table1.id = table2.table1_id"
				},
			},
			expectedSQL: " INNER JOIN table1 ON table1.id = table2.table1_id",
		},
		{
			name: "Multiple joins",
			initialJoins: []func(ctx context.Context) string{
				func(ctx context.Context) string {
					return "INNER JOIN table1 ON table1.id = table2.table1_id"
				},
				func(ctx context.Context) string {
					return "LEFT JOIN table3 ON table3.id = table2.table3_id"
				},
			},
			expectedSQL: " INNER JOIN table1 ON table1.id = table2.table1_id LEFT JOIN table3 ON table3.id = table2.table3_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := &JoinBuilder{
				joins: tc.initialJoins,
			}

			assert.Equal(t, tc.expectedSQL, builder.SQL())
		})
	}
}
