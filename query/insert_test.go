package query

import (
	"errors"
	"testing"

	"github.com/insei/gerpo/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type insertModel struct {
	ID   int
	Name string
}

func TestInsert_Apply_HappyPath(t *testing.T) {
	m := &insertModel{}
	nameCol := &mockColumn{name: "name", hasName: true}
	a := newMockApplier()
	a.storage.byPtr[&m.Name] = nameCol
	a.cols.all = []types.Column{nameCol}

	h := NewInsert(m)
	h.HandleFn(func(m *insertModel, h InsertHelper[insertModel]) {
		h.Exclude(&m.Name)
	})
	require.NoError(t, h.Apply(a))
	assert.Len(t, a.cols.excluded, 1)
}

func TestInsert_Apply_Empty_Noop(t *testing.T) {
	require.NoError(t, NewInsert(&insertModel{}).Apply(newMockApplier()))
}

func TestInsert_Apply_ExcludeMissing_Err(t *testing.T) {
	m := &insertModel{}
	h := NewInsert(m)
	h.HandleFn(func(m *insertModel, h InsertHelper[insertModel]) {
		h.Exclude(&m.ID) // нет в storage
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyExcludeColumnRules))
}
