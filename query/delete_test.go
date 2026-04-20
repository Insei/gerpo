package query

import (
	"context"
	"errors"
	"testing"

	"github.com/insei/gerpo/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type deleteModel struct {
	ID int
}

func TestDelete_Apply_HappyPath(t *testing.T) {
	m := &deleteModel{}
	col := &mockColumn{
		name: "id", hasName: true, sql: "id", allowed: true,
		filters: map[types.Operation]func(ctx context.Context, val any) (string, []any, error){
			types.OperationEQ: func(_ context.Context, v any) (string, []any, error) { return "id = ?", []any{v}, nil },
		},
	}
	a := newMockApplier()
	a.storage.byPtr[&m.ID] = col

	h := NewDelete(m)
	h.HandleFn(func(m *deleteModel, h DeleteHelper[deleteModel]) {
		h.Where().Field(&m.ID).EQ(1)
	})
	require.NoError(t, h.Apply(a))
	assert.Contains(t, a.where.fragments, "id = ?")
}

func TestDelete_Apply_Empty_Noop(t *testing.T) {
	require.NoError(t, NewDelete(&deleteModel{}).Apply(newMockApplier()))
}

func TestDelete_Apply_WhereFailure(t *testing.T) {
	m := &deleteModel{}
	h := NewDelete(m)
	h.HandleFn(func(m *deleteModel, h DeleteHelper[deleteModel]) {
		h.Where().Field(&m.ID).EQ(1)
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyWhereClause))
}
