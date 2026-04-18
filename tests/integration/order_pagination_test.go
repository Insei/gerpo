//go:build integration

package integration

import (
	"testing"

	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrder_ASC сортирует по возрастанию: наименьший age → пользователь с index 0.
func TestOrder_ASC(t *testing.T) {
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
		assert.Equal(t, seed.users[0].ID, got[0].ID, "ASC by Age: youngest first")
		assert.Equal(t, seed.users[9].ID, got[9].ID, "ASC by Age: oldest last")
	})
}

// TestOrder_DESC — обратный порядок.
func TestOrder_DESC(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
			h.OrderBy().Field(&m.Age).DESC()
		})
		require.NoError(t, err)
		require.Len(t, got, len(seed.users))
		assert.Equal(t, seed.users[9].ID, got[0].ID, "DESC by Age: oldest first")
		assert.Equal(t, seed.users[0].ID, got[9].ID, "DESC by Age: youngest last")
	})
}

// TestOrder_MultipleFields — несколько OrderBy комбинируются в одном ORDER BY.
// Сортируем сначала по Published DESC (true до false), затем по CreatedAt ASC.
// В seed'е Published=true у постов с чётным индексом; CreatedAt растёт с индексом.
// Первая запись должна быть Post 0 (published, самая ранняя).
func TestOrder_MultipleFields(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx, func(m *Post, h query.GetListHelper[Post]) {
			h.OrderBy().Field(&m.Published).DESC()
			h.OrderBy().Field(&m.CreatedAt).ASC()
		})
		require.NoError(t, err)
		require.Len(t, got, len(seed.posts))

		// Первый — published и самая ранняя дата → Post 0.
		assert.Equal(t, seed.posts[0].ID, got[0].ID)
		// Среди published (индексы 0,2,4,...,28) последний в ASC по CreatedAt — Post 28.
		assert.Equal(t, seed.posts[28].ID, got[14].ID)
		// Первая unpublished запись — самая ранняя unpublished (индекс 1).
		assert.Equal(t, seed.posts[1].ID, got[15].ID)
	})
}

// TestPagination_PageSize — разбиение по страницам.
func TestPagination_PageSize(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		page1, err := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
			h.OrderBy().Field(&m.Age).ASC()
			h.Page(1).Size(5)
		})
		require.NoError(t, err)
		require.Len(t, page1, 5)
		assert.Equal(t, seed.users[0].ID, page1[0].ID)
		assert.Equal(t, seed.users[4].ID, page1[4].ID)

		page2, err := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
			h.OrderBy().Field(&m.Age).ASC()
			h.Page(2).Size(5)
		})
		require.NoError(t, err)
		require.Len(t, page2, 5)
		assert.Equal(t, seed.users[5].ID, page2[0].ID)
		assert.Equal(t, seed.users[9].ID, page2[4].ID)
	})
}

// TestPagination_EmptyPage — страница за пределами набора пуста.
func TestPagination_EmptyPage(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
			h.Page(3).Size(5) // всего 10 пользователей — третья страница пуста
		})
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

// TestPagination_SizeOnly — только LIMIT, без OFFSET.
func TestPagination_SizeOnly(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
			h.OrderBy().Field(&m.Age).ASC()
			h.Size(3)
		})
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, seed.users[0].ID, got[0].ID)
	})
}

// TestPagination_PageWithoutSize — Page без Size должен вернуть ошибку.
func TestPagination_PageWithoutSize(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		_, err := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
			h.Page(2)
		})
		require.Error(t, err, "Page without Size is invalid")
	})
}
