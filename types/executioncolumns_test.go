package types

import (
	"context"
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type execModel struct {
	ID   int
	Name string
}

func newExecFixture(t *testing.T) (*execModel, []Column, *columnsStorage) {
	t.Helper()
	storage, err := fmap.Get[execModel]()
	require.NoError(t, err)
	m := &execModel{ID: 7, Name: "alice"}

	idField := storage.MustFind("ID")
	nameField := storage.MustFind("Name")

	idCol := &mockColumn{name: "id", hasName: true, allowedAction: true, field: idField}
	nameCol := &mockColumn{name: "name", hasName: true, allowedAction: true, field: nameField}
	cs := NewEmptyColumnsStorage(storage).(*columnsStorage)
	cs.Add(idCol)
	cs.Add(nameCol)
	return m, []Column{idCol, nameCol}, cs
}

func TestExecutionColumns_GetAllAndExcludeOnly(t *testing.T) {
	m, cols, cs := newExecFixture(t)
	_ = m

	ec := cs.NewExecutionColumns(context.Background(), SQLActionSelect)
	assert.Len(t, ec.GetAll(), len(cols))

	ec.Exclude(cols[0])
	all := ec.GetAll()
	assert.Len(t, all, 1)
	assert.Equal(t, cols[1], all[0])

	// Exclude removes only the existing column; excluding an unknown leaves the slice unchanged.
	ec.Exclude(&mockColumn{name: "ghost"})
	assert.Len(t, ec.GetAll(), 1)

	// Only replaces the entire column set.
	ec.Only(cols[0])
	assert.Equal(t, []Column{cols[0]}, ec.GetAll())
}

func TestExecutionColumns_GetByFieldPtr(t *testing.T) {
	m, cols, cs := newExecFixture(t)
	ec := cs.NewExecutionColumns(context.Background(), SQLActionSelect)

	got, err := ec.GetByFieldPtr(m, &m.ID)
	require.NoError(t, err)
	assert.Equal(t, cols[0], got)

	// After Exclude(cols[0]) the lookup must fail even though the column is in storage.
	ec.Exclude(cols[0])
	_, err = ec.GetByFieldPtr(m, &m.ID)
	require.Error(t, err)

	// Unknown pointer → error out of storage.
	type other struct{ Age int }
	var o other
	_, err = ec.GetByFieldPtr(&o, &o.Age)
	require.Error(t, err)
}

func TestExecutionColumns_GetModelPointersAndValues(t *testing.T) {
	m, _, cs := newExecFixture(t)
	ec := cs.NewExecutionColumns(context.Background(), SQLActionSelect)

	ptrs := ec.GetModelPointers(m)
	require.Len(t, ptrs, 2)
	// First is *int for ID.
	if p, ok := ptrs[0].(*int); ok {
		*p = 42
		assert.Equal(t, 42, m.ID, "GetModelPointers returns real pointers")
	} else {
		t.Fatalf("expected *int, got %T", ptrs[0])
	}

	vals := ec.GetModelValues(m)
	require.Len(t, vals, 2)
	assert.Equal(t, 42, vals[0])
	assert.Equal(t, "alice", vals[1])
}
