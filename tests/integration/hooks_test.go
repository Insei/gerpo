//go:build integration

package integration

import (
	"context"
	"errors"
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
	repo, err := gerpo.New[Post]().
		DB(ab.adapter).
		Table("posts").
		Columns(func(m *Post, cb *gerpo.ColumnBuilder[Post]) {
			cb.Field(&m.ID).OmitOnUpdate()
			cb.Field(&m.UserID)
			cb.Field(&m.Title)
			cb.Field(&m.Content)
			cb.Field(&m.Published)
			cb.Field(&m.PublishedAt)
			cb.Field(&m.CreatedAt).OmitOnUpdate()
		}).
		WithBeforeInsert(func(ctx context.Context, m *Post) error {
			c.beforeInsert++
			m.Content = "mutated-by-beforeInsert"
			return nil
		}).
		WithAfterInsert(func(ctx context.Context, m *Post) error {
			c.afterInsert++
			return nil
		}).
		WithBeforeUpdate(func(ctx context.Context, m *Post) error {
			c.beforeUpdate++
			return nil
		}).
		WithAfterUpdate(func(ctx context.Context, m *Post) error {
			c.afterUpdate++
			return nil
		}).
		WithAfterSelect(func(ctx context.Context, models []*Post) error {
			c.afterSelect += len(models)
			return nil
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

// TestHooks_BeforeInsert_ErrorAbortsSQL — если BeforeInsert вернул error,
// INSERT не выполняется, ошибка выходит из repo.Insert.
func TestHooks_BeforeInsert_ErrorAbortsSQL(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		wantErr := errors.New("reject this insert")
		repo, err := gerpo.New[Post]().
			DB(ab.adapter).
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
			WithBeforeInsert(func(_ context.Context, _ *Post) error { return wantErr }).
			Build()
		require.NoError(t, err)

		id := uuid.New()
		p := Post{
			ID: id, UserID: seed.users[0].ID, Title: "should-never-land",
			Content: "c", CreatedAt: time.Now().UTC(),
		}
		err = repo.Insert(ctx, &p)
		require.ErrorIs(t, err, wantErr, "BeforeInsert error must propagate")

		var exists bool
		require.NoError(t,
			pgx5Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`, id).Scan(&exists))
		assert.False(t, exists, "aborted before SQL — row must not exist")
	})
}

// TestHooks_AfterInsert_ErrorSurfaces — если AfterInsert вернул error, SQL
// уже прошёл (запись в БД есть), но ошибка возвращается пользователю.
func TestHooks_AfterInsert_ErrorSurfaces(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		wantErr := errors.New("after-insert hook failed")
		repo, err := gerpo.New[Post]().
			DB(ab.adapter).
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
			WithAfterInsert(func(_ context.Context, _ *Post) error { return wantErr }).
			Build()
		require.NoError(t, err)

		id := uuid.New()
		p := Post{
			ID: id, UserID: seed.users[0].ID, Title: "after-hook-err",
			Content: "c", CreatedAt: time.Now().UTC(),
		}
		err = repo.Insert(ctx, &p)
		require.ErrorIs(t, err, wantErr)

		var exists bool
		require.NoError(t,
			pgx5Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`, id).Scan(&exists))
		assert.True(t, exists, "SQL runs before AfterInsert; without tx rollback row persists")
	})
}

// TestHooks_AfterInsert_CascadeInTx — канонический user-land one-to-many:
// вставка родителя + каскад детей в AfterInsert через тот же ctx-tx.
func TestHooks_AfterInsert_CascadeInTx(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		commentRepo := newCommentRepo(t, ab)
		postWithComments, err := gerpo.New[Post]().
			DB(ab.adapter).
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
			WithAfterInsert(func(ctx context.Context, p *Post) error {
				for i := 0; i < 3; i++ {
					if err := commentRepo.Insert(ctx, &Comment{
						ID: uuid.New(), PostID: p.ID, UserID: p.UserID,
						Body: "cascade-" + p.Title, CreatedAt: time.Now().UTC(),
					}); err != nil {
						return err
					}
				}
				return nil
			}).
			Build()
		require.NoError(t, err)

		postID := uuid.New()
		err = gerpo.RunInTx(ctx, ab.adapter, func(ctx context.Context) error {
			return postWithComments.Insert(ctx, &Post{
				ID: postID, UserID: seed.users[0].ID, Title: "cascade-parent",
				Content: "c", CreatedAt: time.Now().UTC(),
			})
		})
		require.NoError(t, err)

		var postExists bool
		require.NoError(t,
			pgx5Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`, postID).Scan(&postExists))
		assert.True(t, postExists)

		var commentCount int
		require.NoError(t,
			pgx5Pool.QueryRow(ctx, `SELECT count(*) FROM comments WHERE post_id = $1`, postID).Scan(&commentCount))
		assert.Equal(t, 3, commentCount, "cascade must have persisted 3 comments")
	})
}

// TestHooks_AfterInsert_CascadeRollback — если каскад в AfterInsert упал, вся
// tx откатывается: родитель и уже сохранённый ребёнок не остаются в БД.
func TestHooks_AfterInsert_CascadeRollback(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		commentRepo := newCommentRepo(t, ab)
		wantErr := errors.New("cascade child failed")

		postWithBadCascade, err := gerpo.New[Post]().
			DB(ab.adapter).
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
			WithAfterInsert(func(ctx context.Context, p *Post) error {
				if err := commentRepo.Insert(ctx, &Comment{
					ID: uuid.New(), PostID: p.ID, UserID: p.UserID,
					Body: "ok-cascade", CreatedAt: time.Now().UTC(),
				}); err != nil {
					return err
				}
				return wantErr
			}).
			Build()
		require.NoError(t, err)

		postID := uuid.New()
		err = gerpo.RunInTx(ctx, ab.adapter, func(ctx context.Context) error {
			return postWithBadCascade.Insert(ctx, &Post{
				ID: postID, UserID: seed.users[0].ID, Title: "rolled-back",
				Content: "c", CreatedAt: time.Now().UTC(),
			})
		})
		require.ErrorIs(t, err, wantErr)

		var anything bool
		require.NoError(t,
			pgx5Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`, postID).Scan(&anything))
		assert.False(t, anything, "cascade failure must roll back the parent row")

		var orphanComments int
		require.NoError(t,
			pgx5Pool.QueryRow(ctx, `SELECT count(*) FROM comments WHERE post_id = $1`, postID).Scan(&orphanComments))
		assert.Zero(t, orphanComments, "successful child row must have rolled back too")
	})
}

// TestHooks_Stacking — несколько WithBeforeInsert складываются: оба хука зовутся.
func TestHooks_Stacking(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		var first, second int
		repo, err := gerpo.New[Post]().
			DB(ab.adapter).
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
			WithBeforeInsert(func(ctx context.Context, m *Post) error { first++; return nil }).
			WithBeforeInsert(func(ctx context.Context, m *Post) error { second++; return nil }).
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
