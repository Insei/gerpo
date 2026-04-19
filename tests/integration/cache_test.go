//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor"
	cachectx "github.com/insei/gerpo/executor/cache/ctx"
	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCachedPostRepo собирает Post-репозиторий, подключённый к CtxCache.
func newCachedPostRepo(t *testing.T, ab adapterBundle, cache *cachectx.CtxCache) gerpo.Repository[Post] {
	t.Helper()
	repo, err := gerpo.NewBuilder[Post]().
		DB(ab.adapter, executor.WithCacheStorage(cache)).
		Table("posts").
		Columns(func(m *Post, c *gerpo.ColumnBuilder[Post]) {
			c.Field(&m.ID).AsColumn().WithUpdateProtection()
			c.Field(&m.UserID).AsColumn()
			c.Field(&m.Title).AsColumn()
			c.Field(&m.Content).AsColumn()
			c.Field(&m.Published).AsColumn()
			c.Field(&m.PublishedAt).AsColumn()
			c.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
		}).
		Build()
	require.NoError(t, err)
	return repo
}

// TestCache_HitReturnsStaleValue — повторный GetFirst в том же контексте
// возвращает закешированное значение, даже если БД изменилась извне.
func TestCache_HitReturnsStaleValue(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		cache := cachectx.New()
		repo := newCachedPostRepo(t, ab, cache)

		baseCtx, cancel := testCtx(t)
		defer cancel()
		ctx := cachectx.NewCtxCache(baseCtx)

		target := seed.posts[0]

		first, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, target.Title, first.Title)

		// Меняем заголовок напрямую в БД — кеш не должен заметить.
		_, err = pgx5Pool.Exec(ctx, `UPDATE posts SET title = $1 WHERE id = $2`, "external-update", target.ID)
		require.NoError(t, err)

		second, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, target.Title, second.Title, "cached result must be returned, not external value")
	})
}

// TestCache_InvalidatedOnInsert — Insert через репо чистит кеш, следующий GetFirst
// видит актуальные данные.
func TestCache_InvalidatedOnInsert(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		cache := cachectx.New()
		repo := newCachedPostRepo(t, ab, cache)

		baseCtx, cancel := testCtx(t)
		defer cancel()
		ctx := cachectx.NewCtxCache(baseCtx)

		target := seed.posts[0]
		// Первый запрос — кеш заполняется.
		_, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)

		// Внешнее изменение в БД.
		_, err = pgx5Pool.Exec(ctx, `UPDATE posts SET title = $1 WHERE id = $2`, "after-insert-cleans", target.ID)
		require.NoError(t, err)

		// Insert нового поста через репо — чистит кеш.
		other := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "fresh",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		require.NoError(t, repo.Insert(ctx, &other))

		// Следующий GetFirst должен увидеть внешнее изменение, т.к. кеш очищен.
		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, "after-insert-cleans", got.Title, "cache should have been invalidated by Insert")
	})
}

// TestCache_WithoutMiddleware_WorksAsMiss — без NewCtxCache в контексте кеш
// просто не срабатывает, ошибки наружу не просачиваются.
func TestCache_WithoutMiddleware_WorksAsMiss(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		cache := cachectx.New()
		repo := newCachedPostRepo(t, ab, cache)

		ctx, cancel := testCtx(t)
		defer cancel()
		// NewCtxCache не вызываем.

		target := seed.posts[0]
		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, target.Title, got.Title)
	})
}

// TestCache_DifferentContextsDoNotShare — разные контексты имеют независимый кеш.
func TestCache_DifferentContextsDoNotShare(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		cache := cachectx.New()
		repo := newCachedPostRepo(t, ab, cache)

		base, cancel := testCtx(t)
		defer cancel()
		ctx1 := cachectx.NewCtxCache(base)
		ctx2 := cachectx.NewCtxCache(base)

		target := seed.posts[0]
		_, err := repo.GetFirst(ctx1, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)

		// Изменяем название извне.
		_, err = pgx5Pool.Exec(context.Background(), `UPDATE posts SET title = 'ctx2-sees' WHERE id = $1`, target.ID)
		require.NoError(t, err)

		// Запрос во втором контексте — кеш чистый, должен увидеть новое значение.
		got, err := repo.GetFirst(ctx2, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, "ctx2-sees", got.Title, "independent cache in the second context")
	})
}
