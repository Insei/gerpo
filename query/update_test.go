package query

import (
	"context"
	"errors"
	"testing"

	"github.com/insei/gerpo/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type updateModel struct {
	ID   int
	Name string
}

func TestUpdate_Apply_HappyPath(t *testing.T) {
	m := &updateModel{}
	nameCol := &mockColumn{name: "name", hasName: true}
	idCol := &mockColumn{
		name: "id", hasName: true, sql: "id", allowed: true,
		filters: map[types.Operation]func(ctx context.Context, val any) (string, []any, error){
			types.OperationEQ: func(_ context.Context, v any) (string, []any, error) { return "id = ?", []any{v}, nil },
		},
	}
	a := newMockApplier()
	a.storage.byPtr[&m.Name] = nameCol
	a.storage.byPtr[&m.ID] = idCol

	h := NewUpdate(m)
	h.HandleFn(func(m *updateModel, h UpdateHelper[updateModel]) {
		h.Exclude(&m.Name)
		h.Where().Field(&m.ID).EQ(7)
	})
	require.NoError(t, h.Apply(a))
	assert.Len(t, a.cols.excluded, 1)
	assert.Contains(t, a.where.fragments, "id = ?")
}

func TestUpdate_Apply_Empty_Noop(t *testing.T) {
	require.NoError(t, NewUpdate(&updateModel{}).Apply(newMockApplier()))
}

func TestUpdate_Apply_ExcludeMissing_Err(t *testing.T) {
	m := &updateModel{}
	h := NewUpdate(m)
	h.HandleFn(func(m *updateModel, h UpdateHelper[updateModel]) {
		h.Exclude(&m.Name)
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyExcludeColumnRules))
}

func TestUpdate_Apply_WhereFailure(t *testing.T) {
	m := &updateModel{}
	h := NewUpdate(m)
	h.HandleFn(func(m *updateModel, h UpdateHelper[updateModel]) {
		h.Where().Field(&m.ID).EQ(7) // не в storage
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyWhereClause))
}
