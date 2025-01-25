package linq

import (
	"context"
	"testing"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/stretchr/testify/assert"
)

type mockJoin struct {
	sqlpart.Join
	join string
}

func (m *mockJoin) JOIN(joinFn func(ctx context.Context) string) {
	m.join = joinFn(context.Background())
}

type mockJoinApplier struct {
	join *mockJoin
}

func (a *mockJoinApplier) Join() sqlpart.Join {
	return a.join
}

func TestJoinBuilder_LeftJoin(t *testing.T) {
	testCases := []struct {
		name         string
		leftJoinFn   func(ctx context.Context) string
		expectedJoin string
	}{
		{
			name: "Empty_LeftJoin",
			leftJoinFn: func(ctx context.Context) string {
				return ""
			},
			expectedJoin: "",
		},
		{
			name: "Simple_LeftJoin",
			leftJoinFn: func(ctx context.Context) string {
				return "users ON users.id = posts.user_id"
			},
			expectedJoin: "LEFT JOIN users ON users.id = posts.user_id",
		},
		{
			name: "LeftJoin_with_spaces",
			leftJoinFn: func(ctx context.Context) string {
				return "   users ON users.id = posts.user_id   "
			},
			expectedJoin: "LEFT JOIN users ON users.id = posts.user_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewJoinBuilder()
			builder.LeftJoin(tc.leftJoinFn)

			applier := &mockJoinApplier{
				join: &mockJoin{},
			}
			builder.Apply(applier)
			assert.Equal(t, tc.expectedJoin, applier.join.join)
		})
	}
}
