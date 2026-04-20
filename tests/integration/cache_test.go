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

// newCachedPostRepo собирает Post-репозиторий, подключённый к Cache.
func newCachedPostRepo(t *testing.T, ab adapterBundle, cache *cachectx.Cache) gerpo.Repository[Post] {
	t.Helper()
	repo, err := gerpo.New[Post]().
		DB(ab.adapter, executor.WithCacheStorage(cache)).
		Table("posts").
		Columns(func(m *Post, c *gerpo.ColumnBuilder[Post]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.UserID)
			c.Field(&m.Title)
			c.Field(&m.Content)
			c.Field(&m.Published)
			c.Field(&m.PublishedAt)
			c.Field(&m.CreatedAt).OmitOnUpdate()
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
		ctx := cachectx.WrapContext(baseCtx)

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
		ctx := cachectx.WrapContext(baseCtx)

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

// TestCache_WithoutMiddleware_WorksAsMiss — без WrapContext в контексте кеш
// просто не срабатывает, ошибки наружу не просачиваются.
func TestCache_WithoutMiddleware_WorksAsMiss(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		cache := cachectx.New()
		repo := newCachedPostRepo(t, ab, cache)

		ctx, cancel := testCtx(t)
		defer cancel()
		// WrapContext не вызываем.

		target := seed.posts[0]
		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, target.Title, got.Title)
	})
}

// TestCache_WriteOneRepoInvalidatesOther — Update через users-репо должен
// инвалидировать кеш posts-репо, который привязан к тому же ctx. Раньше это
// было не так (per-repo Clean), см. gerpo-review.txt пункт 2.2.
func TestCache_WriteOneRepoInvalidatesOther(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		shared := cachectx.New()

		userRepo, err := gerpo.New[User]().
			DB(ab.adapter, executor.WithCacheStorage(shared)).
			Table("users").
			Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
				c.Field(&m.ID).OmitOnUpdate()
				c.Field(&m.Name)
				c.Field(&m.Email)
				c.Field(&m.Age)
				c.Field(&m.CreatedAt).OmitOnUpdate()
				c.Field(&m.UpdatedAt).OmitOnInsert()
				c.Field(&m.DeletedAt).OmitOnInsert()
			}).
			Build()
		require.NoError(t, err)

		postRepo := newCachedPostRepo(t, ab, shared)

		baseCtx, cancel := testCtx(t)
		defer cancel()
		ctx := cachectx.WrapContext(baseCtx)

		targetUser := seed.users[0]
		targetPost := seed.posts[0]

		// Прогреваем кеш обоих репо.
		_, err = postRepo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(targetPost.ID)
		})
		require.NoError(t, err)

		// Меняем пост снаружи, чтобы кеш и реальность разошлись.
		_, err = pgx5Pool.Exec(ctx, `UPDATE posts SET title = $1 WHERE id = $2`, "after-cross-invalidation", targetPost.ID)
		require.NoError(t, err)

		// Update через users-репо — должен очистить весь per-ctx кеш, в том
		// числе postRepo's.
		updated := targetUser
		updated.Age = targetUser.Age + 1
		_, err = userRepo.Update(ctx, &updated, func(m *User, h query.UpdateHelper[User]) {
			h.Where().Field(&m.ID).EQ(updated.ID)
		})
		require.NoError(t, err)

		// Следующий GetFirst по posts должен увидеть внешнее изменение —
		// кеш очищен записью через другой репо в том же контексте.
		got, err := postRepo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(targetPost.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, "after-cross-invalidation", got.Title,
			"write through userRepo must invalidate postRepo's cache in the shared context")
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
		ctx1 := cachectx.WrapContext(base)
		ctx2 := cachectx.WrapContext(base)

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
