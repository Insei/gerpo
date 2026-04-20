package types

import (
	"context"
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type filtersModel struct {
	Name    string
	Age     int
	Email   *string
	Numbers []int
}

func getField(t *testing.T, name string) fmap.Field {
	t.Helper()
	s, err := fmap.Get[filtersModel]()
	require.NoError(t, err)
	return s.MustFind(name)
}

func TestFilterManager_AddAndGetFilterFn(t *testing.T) {
	f := getField(t, "Name")
	m := NewFilterManagerForField(f)

	m.AddFilterFn(OperationEQ, func(ctx context.Context, v any) (string, bool) {
		return "name = ?", true
	})

	fn, ok := m.GetFilterFn(OperationEQ)
	require.True(t, ok)
	sql, args, err := fn(context.Background(), "alice")
	require.NoError(t, err)
	assert.Equal(t, "name = ?", sql)
	assert.Equal(t, []any{"alice"}, args)

	_, ok = m.GetFilterFn(OperationNotEQ)
	assert.False(t, ok, "unregistered operation → not found")
}

func TestFilterManager_AvailableAndIsAvailable(t *testing.T) {
	m := NewFilterManagerForField(getField(t, "Name"))
	m.AddFilterFn(OperationEQ, func(context.Context, any) (string, bool) { return "", false })
	m.AddFilterFn(OperationNotEQ, func(context.Context, any) (string, bool) { return "", false })

	ops := m.GetAvailableFilterOperations()
	assert.ElementsMatch(t, []Operation{OperationEQ, OperationNotEQ}, ops)
	assert.True(t, m.IsAvailableFilterOperation(OperationEQ))
	assert.False(t, m.IsAvailableFilterOperation(OperationGT))
}

func TestFilterManager_TypeMismatchReturnsError(t *testing.T) {
	m := NewFilterManagerForField(getField(t, "Age"))
	m.AddFilterFn(OperationEQ, func(context.Context, any) (string, bool) { return "age = ?", true })

	fn, _ := m.GetFilterFn(OperationEQ)
	_, _, err := fn(context.Background(), "not a number")
	require.Error(t, err, "string can't be assigned to int field")
}

func TestFilterManager_AddFilterFnArgs_PassesThrough(t *testing.T) {
	m := NewFilterManagerForField(getField(t, "Name"))
	m.AddFilterFnArgs(OperationEQ, func(_ context.Context, v any) (string, []any, error) {
		return "name = ? AND tenant = ?", []any{v, "acme"}, nil
	})

	fn, ok := m.GetFilterFn(OperationEQ)
	require.True(t, ok)
	sql, args, err := fn(context.Background(), "alice")
	require.NoError(t, err)
	assert.Equal(t, "name = ? AND tenant = ?", sql)
	assert.Equal(t, []any{"alice", "acme"}, args)
}

func TestFilterManager_NilForPointerField_OK(t *testing.T) {
	m := NewFilterManagerForField(getField(t, "Email")) // *string
	m.AddFilterFn(OperationEQ, func(context.Context, any) (string, bool) {
		return "email IS NULL", false
	})

	fn, _ := m.GetFilterFn(OperationEQ)
	sql, args, err := fn(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "email IS NULL", sql)
	assert.Nil(t, args)
}

func TestFilterManager_SliceOfValues_OK(t *testing.T) {
	m := NewFilterManagerForField(getField(t, "Age"))
	m.AddFilterFn(OperationIn, func(context.Context, any) (string, bool) { return "age IN (?)", true })

	fn, _ := m.GetFilterFn(OperationIn)
	sql, args, err := fn(context.Background(), []int{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, "age IN (?)", sql)
	assert.Len(t, args, 1, "args wraps the slice; WhereBuilder expands it on append")
}

func TestFilterManager_EmptySlice_NoOp(t *testing.T) {
	m := NewFilterManagerForField(getField(t, "Age"))
	m.AddFilterFn(OperationIn, func(context.Context, any) (string, bool) { return "never", true })

	fn, _ := m.GetFilterFn(OperationIn)
	sql, args, err := fn(context.Background(), []int{})
	require.NoError(t, err)
	assert.Empty(t, sql)
	assert.Nil(t, args)
}
