package virtual

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileFilter_SQL(t *testing.T) {
	fn := compileFilter(SQL("EXISTS (SELECT 1)"))
	sql, args, err := fn(context.Background(), "ignored")
	require.NoError(t, err)
	assert.Equal(t, "EXISTS (SELECT 1)", sql)
	assert.Nil(t, args)
}

func TestCompileFilter_Bound(t *testing.T) {
	fn := compileFilter(Bound{SQL: "SUM(amount) > ?"})
	sql, args, err := fn(context.Background(), 100)
	require.NoError(t, err)
	assert.Equal(t, "SUM(amount) > ?", sql)
	assert.Equal(t, []any{100}, args)
}

func TestCompileFilter_SQLArgs_CopiesArgs(t *testing.T) {
	original := []any{"a", "b"}
	fn := compileFilter(SQLArgs{SQL: "x BETWEEN ? AND ?", Args: original})
	original[0] = "MUTATED"

	sql, args, err := fn(context.Background(), "ignored")
	require.NoError(t, err)
	assert.Equal(t, "x BETWEEN ? AND ?", sql)
	assert.Equal(t, []any{"a", "b"}, args, "compileFilter must copy SQLArgs.Args defensively")
}

func TestCompileFilter_Match_FirstHitWins(t *testing.T) {
	fn := compileFilter(Match{
		Cases: []MatchCase{
			{Value: true, Spec: SQL("EXISTS (...)")},
			{Value: false, Spec: SQL("NOT EXISTS (...)")},
		},
		Default: SQL("FALSE"),
	})

	sql, _, err := fn(context.Background(), true)
	require.NoError(t, err)
	assert.Equal(t, "EXISTS (...)", sql)

	sql, _, err = fn(context.Background(), false)
	require.NoError(t, err)
	assert.Equal(t, "NOT EXISTS (...)", sql)

	sql, _, err = fn(context.Background(), "anything")
	require.NoError(t, err)
	assert.Equal(t, "FALSE", sql)
}

func TestCompileFilter_Match_NoDefaultErrors(t *testing.T) {
	fn := compileFilter(Match{
		Cases: []MatchCase{{Value: 1, Spec: SQL("X")}},
	})

	_, _, err := fn(context.Background(), 999)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no case matched"), "got: %v", err)
}

func TestCompileFilter_Match_NestedSpec(t *testing.T) {
	fn := compileFilter(Match{
		Cases: []MatchCase{
			{Value: "vip", Spec: Bound{SQL: "tier = ? AND priority"}},
		},
		Default: SQLArgs{SQL: "tier = ? AND base", Args: []any{"std"}},
	})

	sql, args, err := fn(context.Background(), "vip")
	require.NoError(t, err)
	assert.Equal(t, "tier = ? AND priority", sql)
	assert.Equal(t, []any{"vip"}, args)

	sql, args, err = fn(context.Background(), "anonymous")
	require.NoError(t, err)
	assert.Equal(t, "tier = ? AND base", sql)
	assert.Equal(t, []any{"std"}, args)
}

func TestCompileFilter_Func_PassesContext(t *testing.T) {
	type ctxKey struct{}
	fn := compileFilter(Func(func(ctx context.Context, value any) (string, []any, error) {
		tid, _ := ctx.Value(ctxKey{}).(string)
		if tid == "" {
			return "", nil, errors.New("missing tenant")
		}
		return "tenant = ? AND v = ?", []any{tid, value}, nil
	}))

	ctx := context.WithValue(context.Background(), ctxKey{}, "acme")
	sql, args, err := fn(ctx, 42)
	require.NoError(t, err)
	assert.Equal(t, "tenant = ? AND v = ?", sql)
	assert.Equal(t, []any{"acme", 42}, args)

	_, _, err = fn(context.Background(), 42)
	require.Error(t, err)
}
