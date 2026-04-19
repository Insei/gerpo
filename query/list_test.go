package query

import (
	"errors"
	"testing"

	"github.com/insei/gerpo/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type listModel struct {
	ID   int
	Name string
}

func TestGetList_Apply_HappyPath(t *testing.T) {
	m := &listModel{}
	nameCol := &mockColumn{name: "name", hasName: true, sql: "name", allowed: true}
	a := newMockApplier()
	a.storage.byPtr[&m.Name] = nameCol
	a.cols.all = []types.Column{nameCol}

	h := NewGetList(m)
	h.HandleFn(func(m *listModel, h GetListHelper[listModel]) {
		h.Exclude(&m.Name)
		h.OrderBy().Field(&m.Name).DESC()
		h.Page(2).Size(10)
	})

	require.NoError(t, h.Apply(a))
	assert.Equal(t, []string{"name DESC"}, a.order.calls)
	assert.Len(t, a.cols.excluded, 1)
	assert.Equal(t, uint64(10), a.limit.limit)
	assert.Equal(t, uint64(10), a.limit.offset, "page 2 size 10 → offset 10")
}

func TestGetList_Apply_EmptyBuilders_Noop(t *testing.T) {
	m := &listModel{}
	require.NoError(t, NewGetList(m).Apply(newMockApplier()))
}

func TestGetList_Apply_PageWithoutSize_ErrApplyLimitOffset(t *testing.T) {
	m := &listModel{}
	h := NewGetList(m)
	h.HandleFn(func(m *listModel, h GetListHelper[listModel]) {
		h.Page(2)
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyLimitOffsetOperator), "got: %v", err)
}

func TestGetList_Apply_WhereFailure_ErrApplyWhere(t *testing.T) {
	m := &listModel{}
	h := NewGetList(m)
	// Column без filter для EQ → Apply упадёт с WhereClause error.
	col := &mockColumn{name: "name", hasName: true, sql: "name", allowed: true, filters: nil}
	a := newMockApplier()
	a.storage.byPtr[&m.Name] = col
	h.HandleFn(func(m *listModel, h GetListHelper[listModel]) {
		h.Where().Field(&m.Name).EQ("x")
	})
	err := h.Apply(a)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyWhereClause), "got: %v", err)
}
