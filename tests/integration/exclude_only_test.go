//go:build integration

package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExclude_GetFirst — Exclude исключает поле из SELECT, оно приходит в zero-state.
func TestExclude_GetFirst(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.posts[4]
		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
			h.Exclude(&m.Content, &m.PublishedAt)
		})
		require.NoError(t, err)
		assert.Equal(t, target.Title, got.Title)
		assert.Empty(t, got.Content, "Content excluded from SELECT should be empty")
		assert.Nil(t, got.PublishedAt, "PublishedAt excluded from SELECT should stay nil")
	})
}

// TestOnly_GetFirst — Only оставляет в SELECT только указанные поля.
func TestOnly_GetFirst(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.posts[2]
		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
			h.Only(&m.Title)
		})
		require.NoError(t, err)
		assert.Equal(t, target.Title, got.Title)
		assert.Empty(t, got.Content, "Content not in Only")
		assert.Equal(t, uuid.Nil, got.ID, "ID not in Only")
	})
}

// TestExclude_GetList_ZeroCreatedAt — поле, исключённое из SELECT, остаётся нулевым.
func TestExclude_GetList(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx, func(m *Post, h query.GetListHelper[Post]) {
			h.Exclude(&m.Content, &m.CreatedAt)
		})
		require.NoError(t, err)
		require.Len(t, got, len(seed.posts))
		for _, p := range got {
			assert.Empty(t, p.Content)
			assert.True(t, p.CreatedAt.IsZero(), "excluded CreatedAt must be zero-value")
		}
	})
}

// TestUpdate_Only — Update с Only обновляет только указанное поле, остальные не трогаются.
func TestUpdate_Only(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.posts[6]
		original := target
		target.Title = "only-title-changed"
		target.Content = "should-be-ignored"
		target.Published = !original.Published

		count, err := repo.Update(ctx, &target, func(m *Post, h query.UpdateHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
			h.Only(&m.Title)
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, "only-title-changed", got.Title)
		assert.Equal(t, original.Content, got.Content, "Content outside Only must not change")
		assert.Equal(t, original.Published, got.Published, "Published outside Only must not change")
	})
}

// TestUpdate_Exclude — Update с Exclude не трогает указанное поле.
func TestUpdate_Exclude(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.posts[8]
		original := target
		target.Title = "new-title"
		target.Content = "excluded-ignored"
		_, err := repo.Update(ctx, &target, func(m *Post, h query.UpdateHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
			h.Exclude(&m.Content)
		})
		require.NoError(t, err)

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, "new-title", got.Title)
		assert.Equal(t, original.Content, got.Content, "excluded Content must not change")
	})
}
