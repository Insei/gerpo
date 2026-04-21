package linq

import (
	"context"
	"errors"
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
	ctx  context.Context
	join *mockJoin
}

func (a *mockJoinApplier) Join() sqlpart.Join   { return a.join }
func (a *mockJoinApplier) Ctx() context.Context { return a.ctx }

func newApplier() *mockJoinApplier {
	return &mockJoinApplier{ctx: context.Background(), join: &mockJoin{}}
}

func TestJoinBuilder_LeftJoinOn_Static(t *testing.T) {
	builder := NewJoinBuilder()
	builder.LeftJoinOn("posts", "posts.user_id = users.id")

	applier := newApplier()
	require.NoError(t, builder.Apply(applier))
	assert.Equal(t, []string{"LEFT JOIN posts ON posts.user_id = users.id"}, applier.join.calls)
	assert.Equal(t, [][]any{nil}, applier.join.args)
}

func TestJoinBuilder_LeftJoinOn_Resolver(t *testing.T) {
	type ctxKey struct{}
	builder := NewJoinBuilder()
	calls := 0
	builder.LeftJoinOn("posts", "posts.user_id = users.id AND posts.tenant = ?",
		func(ctx context.Context) ([]any, error) {
			calls++
			return []any{ctx.Value(ctxKey{})}, nil
		})

	applier := &mockJoinApplier{ctx: context.WithValue(context.Background(), ctxKey{}, "acme"), join: &mockJoin{}}
	require.NoError(t, builder.Apply(applier))
	assert.Equal(t, 1, calls)
	assert.Equal(t, [][]any{{"acme"}}, applier.join.args)

	// Second Apply with a different ctx must invoke resolver again and pick up the new value.
	applier2 := &mockJoinApplier{ctx: context.WithValue(context.Background(), ctxKey{}, "globex"), join: &mockJoin{}}
	require.NoError(t, builder.Apply(applier2))
	assert.Equal(t, 2, calls)
	assert.Equal(t, [][]any{{"globex"}}, applier2.join.args)
}

func TestJoinBuilder_InnerJoinOn_Static(t *testing.T) {
	builder := NewJoinBuilder()
	builder.InnerJoinOn("posts", "posts.user_id = users.id")

	applier := newApplier()
	require.NoError(t, builder.Apply(applier))
	assert.Equal(t, []string{"INNER JOIN posts ON posts.user_id = users.id"}, applier.join.calls)
	assert.Equal(t, [][]any{nil}, applier.join.args)
}

func TestJoinBuilder_Empty_Noop(t *testing.T) {
	builder := NewJoinBuilder()
	applier := newApplier()
	require.NoError(t, builder.Apply(applier))
	assert.Empty(t, applier.join.calls)
}

func TestJoinBuilder_Resolver_Error_Aborts(t *testing.T) {
	builder := NewJoinBuilder()
	sentinel := errors.New("tenant missing in ctx")
	builder.LeftJoinOn("posts", "posts.user_id = users.id AND posts.tenant = ?",
		func(ctx context.Context) ([]any, error) { return nil, sentinel })

	applier := newApplier()
	err := builder.Apply(applier)
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	assert.Empty(t, applier.join.calls, "JOINOn must not be called when resolver errors")
}

func TestJoinBuilder_MultipleResolvers_Panic(t *testing.T) {
	builder := NewJoinBuilder()
	r := func(ctx context.Context) ([]any, error) { return nil, nil }
	require.PanicsWithValue(t,
		"gerpo: LeftJoinOn accepts at most one resolver, got 2",
		func() { builder.LeftJoinOn("posts", "on", r, r) })
	require.PanicsWithValue(t,
		"gerpo: InnerJoinOn accepts at most one resolver, got 2",
		func() { builder.InnerJoinOn("posts", "on", r, r) })
}
