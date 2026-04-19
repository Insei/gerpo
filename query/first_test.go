package query

import (
	"errors"
	"testing"

	"github.com/insei/gerpo/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type firstModel struct {
	ID   int
	Name string
}

func TestGetFirst_Apply_HappyPath(t *testing.T) {
	m := &firstModel{}
	nameCol := &mockColumn{name: "name", hasName: true, sql: "name", allowed: true}
	a := newMockApplier()
	a.storage.byPtr[&m.Name] = nameCol
	a.cols.all = []types.Column{nameCol}

	h := NewGetFirst(m)
	h.HandleFn(func(m *firstModel, h GetFirstHelper[firstModel]) {
		h.OrderBy().Field(&m.Name).ASC()
		h.Exclude(&m.Name)
	})

	require.NoError(t, h.Apply(a))
	assert.Equal(t, []string{"name ASC"}, a.order.calls)
	assert.Len(t, a.cols.excluded, 1, "Exclude should have flagged the column")
}

func TestGetFirst_Apply_EmptyBuilders_Noop(t *testing.T) {
	m := &firstModel{}
	h := NewGetFirst(m)
	require.NoError(t, h.Apply(newMockApplier()))
}

func TestGetFirst_Apply_ExcludeMissingColumn_ErrApplyExclude(t *testing.T) {
	m := &firstModel{}
	h := NewGetFirst(m)
	h.HandleFn(func(m *firstModel, h GetFirstHelper[firstModel]) {
		h.Exclude(&m.ID) // нет в storage.byPtr → GetByFieldPtr вернёт ошибку
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyExcludeColumnRules), "got: %v", err)
}

func TestGetFirst_Apply_OrderMissingColumn_ErrApplyOrderBy(t *testing.T) {
	m := &firstModel{}
	h := NewGetFirst(m)
	h.HandleFn(func(m *firstModel, h GetFirstHelper[firstModel]) {
		h.OrderBy().Field(&m.ID).ASC() // не настроено → ошибка
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyOrderByOperator), "got: %v", err)
}
