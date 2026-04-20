//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hookCounters struct {
	beforeInsert int
	afterInsert  int
	beforeUpdate int
	afterUpdate  int
	afterSelect  int
}

// newPostRepoWithHooks собирает Post-репо с подсчётом вызовов всех хуков.
// beforeInsert выставляет маркер в Content, чтобы проверить, что мутация
// из хука попадает в INSERT.
func newPostRepoWithHooks(t *testing.T, ab adapterBundle, c *hookCounters) gerpo.Repository[Post] {
	t.Helper()
	repo, err := gerpo.NewBuilder[Post]().
		DB(ab.adapter).
		Table("posts").
		Columns(func(m *Post, cb *gerpo.ColumnBuilder[Post]) {
			cb.Field(&m.ID).WithUpdateProtection()
			cb.Field(&m.UserID)
			cb.Field(&m.Title)
			cb.Field(&m.Content)
			cb.Field(&m.Published)
			cb.Field(&m.PublishedAt)
			cb.Field(&m.CreatedAt).WithUpdateProtection()
		}).
		WithBeforeInsert(func(ctx context.Context, m *Post) {
			c.beforeInsert++
			m.Content = "mutated-by-beforeInsert"
		}).
		WithAfterInsert(func(ctx context.Context, m *Post) {
			c.afterInsert++
		}).
		WithBeforeUpdate(func(ctx context.Context, m *Post) {
			c.beforeUpdate++
		}).
		WithAfterUpdate(func(ctx context.Context, m *Post) {
			c.afterUpdate++
		}).
		WithAfterSelect(func(ctx context.Context, models []*Post) {
			c.afterSelect += len(models)
		}).
		Build()
	require.NoError(t, err)
	return repo
}

// TestHooks_Insert_BeforeAfter — оба хука дёрнулись ровно один раз, мутация
// из BeforeInsert попала в базу.
func TestHooks_Insert_BeforeAfter(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		c := &hookCounters{}
		repo := newPostRepoWithHooks(t, ab, c)
		ctx, cancel := testCtx(t)
		defer cancel()

		p := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "hooked",
			Content:   "original",
			CreatedAt: time.Now().UTC(),
		}
		require.NoError(t, repo.Insert(ctx, &p))
		assert.Equal(t, 1, c.beforeInsert)
		assert.Equal(t, 1, c.afterInsert)

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(p.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, "mutated-by-beforeInsert", got.Content, "BeforeInsert mutation should persist")
	})
}

// TestHooks_Update_BeforeAfter — хуки UPDATE зовутся ровно один раз.
func TestHooks_Update_BeforeAfter(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		c := &hookCounters{}
		repo := newPostRepoWithHooks(t, ab, c)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.posts[0]
		target.Title = "updated"
		_, err := repo.Update(ctx, &target, func(m *Post, h query.UpdateHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, 1, c.beforeUpdate)
		assert.Equal(t, 1, c.afterUpdate)
	})
}

// TestHooks_AfterSelect_GetFirst — afterSelect получает срез из одной записи.
func TestHooks_AfterSelect_GetFirst(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		c := &hookCounters{}
		repo := newPostRepoWithHooks(t, ab, c)
		ctx, cancel := testCtx(t)
		defer cancel()

		_, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(seed.posts[0].ID)
		})
		require.NoError(t, err)
		assert.Equal(t, 1, c.afterSelect)
	})
}

// TestHooks_AfterSelect_GetList — afterSelect получает срез со всеми записями.
func TestHooks_AfterSelect_GetList(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		c := &hookCounters{}
		repo := newPostRepoWithHooks(t, ab, c)
		ctx, cancel := testCtx(t)
		defer cancel()

		_, err := repo.GetList(ctx)
		require.NoError(t, err)
		assert.Equal(t, len(seed.posts), c.afterSelect)
	})
}

// TestHooks_Stacking — несколько WithBeforeInsert складываются: оба хука зовутся.
func TestHooks_Stacking(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		var first, second int
		repo, err := gerpo.NewBuilder[Post]().
			DB(ab.adapter).
			Table("posts").
			Columns(func(m *Post, c *gerpo.ColumnBuilder[Post]) {
				c.Field(&m.ID).WithUpdateProtection()
				c.Field(&m.UserID)
				c.Field(&m.Title)
				c.Field(&m.Content)
				c.Field(&m.Published)
				c.Field(&m.PublishedAt)
				c.Field(&m.CreatedAt).WithUpdateProtection()
			}).
			WithBeforeInsert(func(ctx context.Context, m *Post) { first++ }).
			WithBeforeInsert(func(ctx context.Context, m *Post) { second++ }).
			Build()
		require.NoError(t, err)

		p := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "stacked",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		require.NoError(t, repo.Insert(ctx, &p))
		assert.Equal(t, 1, first)
		assert.Equal(t, 1, second)
	})
}
