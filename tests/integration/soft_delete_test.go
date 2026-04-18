//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/insei/gerpo"
	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSoftDelete_UpdatesInsteadOfDeleting — Delete на репо с WithSoftDeletion
// выставляет deleted_at, а не удаляет строку физически.
func TestSoftDelete_UpdatesInsteadOfDeleting(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.users[3]
		count, err := repo.Delete(ctx, func(m *User, h query.DeleteHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Row is still in the table, but deleted_at is not NULL.
		var deletedAt *time.Time
		err = pgx5Pool.QueryRow(ctx, `SELECT deleted_at FROM users WHERE id = $1`, target.ID).Scan(&deletedAt)
		require.NoError(t, err, "row must still physically exist")
		require.NotNil(t, deletedAt, "deleted_at must be populated")
	})
}

// TestSoftDelete_HiddenFromRepo — soft-deleted запись не видна через репо
// (persistent Where исключает её).
func TestSoftDelete_HiddenFromRepo(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.users[4]
		_, err := repo.Delete(ctx, func(m *User, h query.DeleteHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)

		_, err = repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, gerpo.ErrNotFound))
	})
}

// TestSoftDelete_NothingToDelete — повторный Delete по уже soft-deleted записи
// возвращает ErrNotFound, т.к. persistent Where отсекает её от UPDATE.
func TestSoftDelete_NothingToDelete(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.users[5]
		_, err := repo.Delete(ctx, func(m *User, h query.DeleteHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)

		// Повторный Delete — уже нечего удалять.
		_, err = repo.Delete(ctx, func(m *User, h query.DeleteHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, gerpo.ErrNotFound))
	})
}

// TestSoftDelete_Restore — прямым UPDATE в БД можно «восстановить» запись,
// после чего она снова видна репо.
func TestSoftDelete_Restore(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.users[6]
		_, err := repo.Delete(ctx, func(m *User, h query.DeleteHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)

		_, err = pgx5Pool.Exec(ctx, `UPDATE users SET deleted_at = NULL WHERE id = $1`, target.ID)
		require.NoError(t, err)

		got, err := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, target.ID, got.ID)
	})
}
