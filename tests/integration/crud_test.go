//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetFirst_ByID: GetFirst с точечным фильтром возвращает ожидаемую запись.
func TestGetFirst_ByID(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.posts[5]
		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, target.ID, got.ID)
		assert.Equal(t, target.Title, got.Title)
		assert.Equal(t, target.Published, got.Published)
	})
}

// TestGetFirst_NotFound: GetFirst по несуществующему ID возвращает gerpo.ErrNotFound.
func TestGetFirst_NotFound(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		_, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(uuid.New())
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, gerpo.ErrNotFound), "expected ErrNotFound, got %v", err)
	})
}

// TestGetList_All: GetList без фильтров возвращает все записи.
func TestGetList_All(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx)
		require.NoError(t, err)
		assert.Len(t, got, len(seed.posts))
	})
}

// TestGetList_Empty: GetList на пустой таблице возвращает пустой срез без ошибки.
func TestGetList_Empty(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		// truncateAll уже вызван forEachAdapter — seed не зовём.
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

// TestCount_All: Count без фильтра возвращает общее количество.
func TestCount_All(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(len(seed.posts)), got)
	})
}

// TestCount_WithFilter: Count с WHERE возвращает только подходящие записи.
func TestCount_WithFilter(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		wantPublished := 0
		for _, p := range seed.posts {
			if p.Published {
				wantPublished++
			}
		}

		got, err := repo.Count(ctx, func(m *Post, h query.CountHelper[Post]) {
			h.Where().Field(&m.Published).EQ(true)
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(wantPublished), got)
	})
}

// TestInsert_Happy: Insert добавляет запись, которую затем можно прочитать GetFirst'ом.
func TestInsert_Happy(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		newPost := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "inserted",
			Content:   "inserted content",
			Published: false,
			CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		}
		require.NoError(t, repo.Insert(ctx, &newPost))

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(newPost.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, newPost.Title, got.Title)
		assert.Equal(t, newPost.Content, got.Content)
		assert.False(t, got.Published)
	})
}

// TestInsert_WithExclude: Insert с Exclude не включает поле в INSERT, БД использует
// колоночный DEFAULT. Проверяем на published_at — NULL при исключении.
func TestInsert_WithExclude(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		fixed := time.Date(2030, 5, 5, 0, 0, 0, 0, time.UTC)
		newPost := Post{
			ID:          uuid.New(),
			UserID:      seed.users[1].ID,
			Title:       "exclude",
			Content:     "body",
			Published:   false,
			PublishedAt: &fixed, // должно быть проигнорировано Exclude
			CreatedAt:   time.Now().UTC().Truncate(time.Microsecond),
		}
		err := repo.Insert(ctx, &newPost, func(m *Post, h query.InsertHelper[Post]) {
			h.Exclude(&m.PublishedAt)
		})
		require.NoError(t, err)

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(newPost.ID)
		})
		require.NoError(t, err)
		assert.Nil(t, got.PublishedAt, "PublishedAt should be NULL because it was excluded")
	})
}

// TestUpdate_Happy: Update меняет поле по WHERE и возвращает количество затронутых строк.
func TestUpdate_Happy(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.posts[2]
		target.Title = "renamed"
		count, err := repo.Update(ctx, &target, func(m *Post, h query.UpdateHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, "renamed", got.Title)
	})
}

// TestUpdate_NothingToUpdate: Update по несуществующему ID возвращает ErrNotFound.
func TestUpdate_NothingToUpdate(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		ghost := Post{
			ID:        uuid.New(),
			Title:     "ghost",
			Content:   "nothing",
			UserID:    uuid.New(),
			CreatedAt: time.Now().UTC(),
		}
		_, err := repo.Update(ctx, &ghost, func(m *Post, h query.UpdateHelper[Post]) {
			h.Where().Field(&m.ID).EQ(ghost.ID)
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, gerpo.ErrNotFound), "expected ErrNotFound, got %v", err)
	})
}

// TestDelete_Happy: Delete удаляет запись; последующий GetFirst возвращает ErrNotFound.
func TestDelete_Happy(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		target := seed.posts[7]
		count, err := repo.Delete(ctx, func(m *Post, h query.DeleteHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		_, err = repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(target.ID)
		})
		assert.True(t, errors.Is(err, gerpo.ErrNotFound))
	})
}

// TestDelete_NothingToDelete: Delete по несуществующему ID возвращает ErrNotFound.
func TestDelete_NothingToDelete(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		_, err := repo.Delete(ctx, func(m *Post, h query.DeleteHelper[Post]) {
			h.Where().Field(&m.ID).EQ(uuid.New())
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, gerpo.ErrNotFound), "expected ErrNotFound, got %v", err)
	})
}
