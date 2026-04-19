//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPersistent_LeftJoin_VirtualColumn — persistent LeftJoin + virtual column (post_count).
// Для каждого пользователя в seed'е по 3 поста, значит post_count=3 для всех.
func TestPersistent_LeftJoin_VirtualColumn(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
			h.OrderBy().Field(&m.Age).ASC()
		})
		require.NoError(t, err)
		require.Len(t, got, len(seed.users))
		for i, u := range got {
			assert.Equal(t, 3, u.PostCount, "user %d (%s) expected 3 posts", i, u.ID)
		}
	})
}

// TestPersistent_Where_HidesSoftDeleted — persistent Where (`deleted_at IS NULL`)
// должен скрывать soft-deleted записи от GetList, GetFirst и Count.
func TestPersistent_Where_HidesSoftDeleted(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		// Помечаем user[0] как удалённого напрямую через sql, минуя репо.
		now := nowUTC()
		_, err := pgx5Pool.Exec(ctx, `UPDATE users SET deleted_at = $1 WHERE id = $2`, now, seed.users[0].ID)
		require.NoError(t, err)

		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(len(seed.users)-1), count, "soft-deleted user must be hidden from Count")

		list, err := repo.GetList(ctx)
		require.NoError(t, err)
		assert.Len(t, list, len(seed.users)-1)
		for _, u := range list {
			assert.NotEqual(t, seed.users[0].ID, u.ID, "soft-deleted user must be hidden from GetList")
		}

		_, err = repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.ID).EQ(seed.users[0].ID)
		})
		require.Error(t, err, "GetFirst for soft-deleted id must return error")
	})
}

// TestPersistent_Where_CombinesWithRequestWhere — persistent WHERE комбинируется
// с per-request WHERE через AND.
func TestPersistent_Where_CombinesWithRequestWhere(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		// Пометим user[0] (age=20) как soft-deleted. Per-request фильтр Age<23
		// в обычной ситуации дал бы 3 записи (age 20,21,22). С soft delete — 2.
		_, err := pgx5Pool.Exec(ctx, `UPDATE users SET deleted_at = NOW() WHERE id = $1`, seed.users[0].ID)
		require.NoError(t, err)

		got, err := repo.Count(ctx, func(m *User, h query.CountHelper[User]) {
			h.Where().Field(&m.Age).LT(23)
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(2), got)
	})
}

// TestPersistent_InnerJoin — отдельный репо с InnerJoin возвращает только те
// записи users, для которых есть хотя бы один post. Вставим пользователя без
// постов и убедимся, что он не попадает в выборку.
func TestPersistent_InnerJoin(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		// Пользователь без постов — должен отсечься InnerJoin'ом.
		lonelyUser := User{
			ID:        uuid.New(),
			Name:      "lonely",
			Age:       99,
			CreatedAt: nowUTC(),
		}
		_, err := pgx5Pool.Exec(ctx, `INSERT INTO users (id, name, age, created_at) VALUES ($1,$2,$3,$4)`,
			lonelyUser.ID, lonelyUser.Name, lonelyUser.Age, lonelyUser.CreatedAt)
		require.NoError(t, err)

		repo, err := gerpo.NewBuilder[User]().
			DB(ab.adapter).
			Table("users").
			Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
				c.Field(&m.ID).AsColumn().WithUpdateProtection()
				c.Field(&m.Name).AsColumn()
				c.Field(&m.Email).AsColumn()
				c.Field(&m.Age).AsColumn()
				c.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
				c.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
				c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
			}).
			WithQuery(func(m *User, h query.PersistentHelper[User]) {
				h.InnerJoin(func(ctx context.Context) string { return "posts ON posts.user_id = users.id" })
				h.GroupBy(&m.ID, &m.Name, &m.Email, &m.Age, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
				h.Where().Field(&m.DeletedAt).EQ(nil)
			}).
			Build()
		require.NoError(t, err)

		list, err := repo.GetList(ctx)
		require.NoError(t, err)
		assert.Len(t, list, len(seed.users), "InnerJoin should omit users without posts")
		for _, u := range list {
			assert.NotEqual(t, lonelyUser.ID, u.ID, "lonely user must not appear")
		}
	})
}

// TestPersistent_LeftJoinOn_BindsArgs — the bound JOIN form sends ON-clause
// values through the driver. The repo joins posts only for a specific user_id;
// post_count then reflects only that user's posts.
func TestPersistent_LeftJoinOn_BindsArgs(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		// JOIN restricts the relationship to user[3] only — every other row
		// will get a NULL right-hand side and post_count = 0.
		targetUserID := seed.users[3].ID

		repo, err := gerpo.NewBuilder[User]().
			DB(ab.adapter).
			Table("users").
			Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
				c.Field(&m.ID).AsColumn().WithUpdateProtection()
				c.Field(&m.Name).AsColumn()
				c.Field(&m.Email).AsColumn()
				c.Field(&m.Age).AsColumn()
				c.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
				c.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
				c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
				c.Field(&m.PostCount).AsVirtual().WithSQL(func(ctx context.Context) string {
					return "COALESCE(COUNT(posts.id), 0)"
				})
			}).
			WithQuery(func(m *User, h query.PersistentHelper[User]) {
				h.LeftJoinOn(
					"posts",
					"posts.user_id = users.id AND posts.user_id = ?",
					targetUserID,
				)
				h.GroupBy(&m.ID, &m.Name, &m.Email, &m.Age, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
				h.Where().Field(&m.DeletedAt).EQ(nil)
			}).
			Build()
		require.NoError(t, err)

		got, err := repo.GetList(ctx)
		require.NoError(t, err)
		require.Len(t, got, len(seed.users))

		var hits int
		for _, u := range got {
			if u.ID == targetUserID {
				assert.Equal(t, 3, u.PostCount, "target user keeps its 3 seeded posts")
				hits++
				continue
			}
			assert.Equal(t, 0, u.PostCount, "non-target user must have post_count=0 because JOIN ON filtered them out")
		}
		assert.Equal(t, 1, hits, "exactly one row matches targetUserID")
	})
}

// TestPersistent_InnerJoinOn_FiltersByBoundArg — InnerJoinOn variant: only the
// users matching the bound condition appear in the result.
func TestPersistent_InnerJoinOn_FiltersByBoundArg(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		// Restrict the inner-join to a single user — only that user appears.
		targetUserID := seed.users[2].ID

		repo, err := gerpo.NewBuilder[User]().
			DB(ab.adapter).
			Table("users").
			Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
				c.Field(&m.ID).AsColumn().WithUpdateProtection()
				c.Field(&m.Name).AsColumn()
				c.Field(&m.Email).AsColumn()
				c.Field(&m.Age).AsColumn()
				c.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
				c.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
				c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
			}).
			WithQuery(func(m *User, h query.PersistentHelper[User]) {
				h.InnerJoinOn(
					"posts",
					"posts.user_id = users.id AND posts.user_id = ?",
					targetUserID,
				)
				h.GroupBy(&m.ID, &m.Name, &m.Email, &m.Age, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
				h.Where().Field(&m.DeletedAt).EQ(nil)
			}).
			Build()
		require.NoError(t, err)

		got, err := repo.GetList(ctx)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, targetUserID, got[0].ID)
	})
}
