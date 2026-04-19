package query

import (
	"context"
	"errors"
	"testing"

	"github.com/insei/gerpo/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type persistentModel struct {
	ID        int
	DeletedAt *int
}

func TestPersistent_Apply_AllOpsForwarded(t *testing.T) {
	m := &persistentModel{}
	delCol := &mockColumn{
		name: "deleted_at", hasName: true, sql: "deleted_at", allowed: true,
		filters: map[types.Operation]func(ctx context.Context, val any) (string, bool, error){
			types.OperationEQ: func(context.Context, any) (string, bool, error) { return "deleted_at IS NULL", false, nil },
		},
	}
	idCol := &mockColumn{name: "id", hasName: true, sql: "id", allowed: true}
	a := newMockApplier()
	a.storage.byPtr[&m.DeletedAt] = delCol
	a.storage.byPtr[&m.ID] = idCol
	a.cols.all = []types.Column{idCol, delCol}

	h := NewPersistent(m)
	h.HandleFn(func(m *persistentModel, h PersistentHelper[persistentModel]) {
		h.Where().Field(&m.DeletedAt).EQ(nil)
		h.GroupBy(&m.ID)
		h.Exclude(&m.DeletedAt)
		h.LeftJoinOn("posts", "posts.user_id = users.id AND posts.tenant = ?", "T")
	})

	require.NoError(t, h.Apply(a))
	assert.Contains(t, a.where.fragments, "deleted_at IS NULL")
	assert.Equal(t, []types.Column{idCol}, a.group.cols)
	assert.Len(t, a.cols.excluded, 1)
	assert.Equal(t, []string{"LEFT JOIN posts ON posts.user_id = users.id AND posts.tenant = ?"}, a.join.bound)
	assert.Equal(t, [][]any{{"T"}}, a.join.boundArgs)
}

func TestPersistent_Apply_NilApplier_Err(t *testing.T) {
	err := NewPersistent(&persistentModel{}).Apply(nil)
	require.Error(t, err)
}

func TestPersistent_Apply_GroupMissingColumn_Err(t *testing.T) {
	m := &persistentModel{}
	h := NewPersistent(m)
	h.HandleFn(func(m *persistentModel, h PersistentHelper[persistentModel]) {
		h.GroupBy(&m.ID) // нет в storage
	})
	err := h.Apply(newMockApplier())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApplyGroupByClause))
}
