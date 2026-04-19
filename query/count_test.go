package query

import (
	"context"
	"errors"
	"testing"

	"github.com/insei/gerpo/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countModel struct {
	ID int
}

func TestCount_Apply_HappyPath(t *testing.T) {
	m := &countModel{}
	col := &mockColumn{
		name: "id", hasName: true, sql: "id", allowed: true,
		filters: map[types.Operation]func(ctx context.Context, val any) (string, bool, error){
			types.OperationEQ: func(context.Context, any) (string, bool, error) { return "id = ?", true, nil },
		},
	}
	a := newMockApplier()
	a.storage.byPtr[&m.ID] = col

	h := NewCount(m)
	h.HandleFn(func(m *countModel, h CountHelper[countModel]) {
		h.Where().Field(&m.ID).EQ(1)
	})
	require.NoError(t, h.Apply(a))
	assert.Contains(t, a.where.fragments, "id = ?")
	assert.Equal(t, []any{1}, a.where.values)
}

func TestCount_Apply_Empty_Noop(t *testing.T) {
	m := &countModel{}
	require.NoError(t, NewCount(m).Apply(newMockApplier()))
}

func TestCount_Apply_WhereFailure(t *testing.T) {
	m := &countModel{}
	h := NewCount(m)
	h.HandleFn(func(m *countModel, h CountHelper[countModel]) {
		h.Where().Field(&m.ID).EQ(1) // column не в storage
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyWhereClause))
}
