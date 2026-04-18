//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVirtual_ReadOnlyOnInsert — virtual column PostCount не попадает в INSERT.
// Устанавливаем в модели заведомо ложное значение (999), но БД принимает Insert
// без ошибки, а следующий SELECT возвращает реальное значение (0, т.к. постов нет).
func TestVirtual_ReadOnlyOnInsert(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		newUser := User{
			ID:        uuid.New(),
			Name:      "virtual-insert",
			Age:       42,
			CreatedAt: time.Now().UTC(),
			PostCount: 999, // должен быть проигнорирован при INSERT
		}
		require.NoError(t, repo.Insert(ctx, &newUser))

		got, err := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.ID).EQ(newUser.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, 0, got.PostCount, "virtual column must reflect real state, not model value")
	})
}

// TestVirtual_ReadOnlyOnUpdate — попытка обновить PostCount через Update не меняет ничего.
func TestVirtual_ReadOnlyOnUpdate(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.users[1]
		target.PostCount = 999
		target.Age = 50 // меняем и реальную колонку, чтобы Update не был no-op
		_, err := repo.Update(ctx, &target, func(m *User, h query.UpdateHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)

		got, err := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, 50, got.Age, "real column must update")
		assert.Equal(t, 3, got.PostCount, "virtual column must stay at computed value (3 posts)")
	})
}

// TestVirtual_ComputedValueReflectsState — virtual column отражает актуальное
// состояние связанных данных. После удаления поста post_count уменьшается.
func TestVirtual_ComputedValueReflectsState(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		userRepo := newUserRepo(t, ab)
		postRepo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		// Удалим один пост у user[0].
		_, err := postRepo.Delete(ctx, func(m *Post, h query.DeleteHelper[Post]) {
			h.Where().Field(&m.ID).EQ(seed.posts[0].ID)
		})
		require.NoError(t, err)

		got, err := userRepo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.ID).EQ(seed.users[0].ID)
		})
		require.NoError(t, err)
		assert.Equal(t, 2, got.PostCount, "post_count must drop to 2 after deleting one post")
	})
}
