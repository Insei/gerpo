//go:build integration

package integration

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDomainPostNotFound = errors.New("domain: post not found")

// newPostRepoWithErrorTransformer — Post-репо с WithErrorTransformer, подменяющим
// gerpo.ErrNotFound на доменную ошибку.
func newPostRepoWithErrorTransformer(t *testing.T, ab adapterBundle) gerpo.Repository[Post] {
	t.Helper()
	repo, err := gerpo.NewBuilder[Post]().
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
		WithErrorTransformer(func(err error) error {
			if errors.Is(err, gerpo.ErrNotFound) {
				return errDomainPostNotFound
			}
			return err
		}).
		Build()
	require.NoError(t, err)
	return repo
}

// TestErrorTransformer_GetFirst — GetFirst на несуществующей записи возвращает доменную ошибку.
func TestErrorTransformer_GetFirst(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepoWithErrorTransformer(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		_, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(uuid.New())
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, errDomainPostNotFound)
		assert.NotErrorIs(t, err, gerpo.ErrNotFound, "raw ErrNotFound should be swallowed by the transformer")
	})
}

// TestErrorTransformer_Update — Update ничего не обновивший тоже проходит через transformer.
func TestErrorTransformer_Update(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepoWithErrorTransformer(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		ghost := Post{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Title:     "ghost",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		_, err := repo.Update(ctx, &ghost, func(m *Post, h query.UpdateHelper[Post]) {
			h.Where().Field(&m.ID).EQ(ghost.ID)
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, errDomainPostNotFound)
	})
}

// TestErrorTransformer_Delete — Delete ничего не удаливший возвращает доменную ошибку.
func TestErrorTransformer_Delete(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepoWithErrorTransformer(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		_, err := repo.Delete(ctx, func(m *Post, h query.DeleteHelper[Post]) {
			h.Where().Field(&m.ID).EQ(uuid.New())
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, errDomainPostNotFound)
	})
}

// TestErrorTransformer_PassthroughOnHappyPath — при успехе transformer не искажает результат.
func TestErrorTransformer_PassthroughOnHappyPath(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepoWithErrorTransformer(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(seed.posts[0].ID)
		})
		require.NoError(t, err)
		assert.Equal(t, seed.posts[0].Title, got.Title)
	})
}

// TestErrorTransformer_DoesNotSwallowOtherErrors — transformer получает все ошибки,
// но не трогает те, что не ErrNotFound. Проверяется через Insert с конфликтом FK.
func TestErrorTransformer_DoesNotSwallowOtherErrors(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepoWithErrorTransformer(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		// user_id не существует → FK violation, это не ErrNotFound.
		p := Post{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Title:     "orphan",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		err := repo.Insert(ctx, &p)
		require.Error(t, err)
		assert.NotErrorIs(t, err, errDomainPostNotFound, "non-NotFound errors must not be mapped to domain")
		// Сама ошибка должна упоминать FK / нарушение.
		assert.Contains(t, fmt.Sprintf("%v", err), "violates", "expected FK violation message")
	})
}
