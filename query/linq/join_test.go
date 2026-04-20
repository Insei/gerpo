package linq

import (
	"testing"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockJoin struct {
	sqlpart.Join
	calls []string
	args  [][]any
}

func (m *mockJoin) JOINOn(sql string, args ...any) {
	m.calls = append(m.calls, sql)
	m.args = append(m.args, args)
}

type mockJoinApplier struct {
	join *mockJoin
}

func (a *mockJoinApplier) Join() sqlpart.Join { return a.join }

func TestJoinBuilder_LeftJoinOn(t *testing.T) {
	builder := NewJoinBuilder()
	builder.LeftJoinOn("posts", "posts.user_id = users.id AND posts.tenant = ?", "acme")

	applier := &mockJoinApplier{join: &mockJoin{}}
	require.NoError(t, builder.Apply(applier))
	assert.Equal(t, []string{"LEFT JOIN posts ON posts.user_id = users.id AND posts.tenant = ?"}, applier.join.calls)
	assert.Equal(t, [][]any{{"acme"}}, applier.join.args)
}

func TestJoinBuilder_InnerJoinOn(t *testing.T) {
	builder := NewJoinBuilder()
	builder.InnerJoinOn("posts", "posts.user_id = users.id")

	applier := &mockJoinApplier{join: &mockJoin{}}
	require.NoError(t, builder.Apply(applier))
	assert.Equal(t, []string{"INNER JOIN posts ON posts.user_id = users.id"}, applier.join.calls)
	assert.Equal(t, [][]any{nil}, applier.join.args)
}

func TestJoinBuilder_Empty_Noop(t *testing.T) {
	builder := NewJoinBuilder()
	applier := &mockJoinApplier{join: &mockJoin{}}
	require.NoError(t, builder.Apply(applier))
	assert.Empty(t, applier.join.calls)
}
