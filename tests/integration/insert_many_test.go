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

// TestInsertMany_BasicRoundTrip — многострочный INSERT проходит на всех адаптерах,
// count равен длине среза, данные попадают в таблицу.
func TestInsertMany_BasicRoundTrip(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		author := seed.users[0].ID
		batch := []*Post{
			{ID: uuid.New(), UserID: author, Title: "bulk-1", Content: "c1", Published: true, CreatedAt: nowUTC()},
			{ID: uuid.New(), UserID: author, Title: "bulk-2", Content: "c2", Published: false, CreatedAt: nowUTC()},
			{ID: uuid.New(), UserID: author, Title: "bulk-3", Content: "c3", Published: true, CreatedAt: nowUTC()},
		}

		n, err := repo.InsertMany(ctx, batch)
		require.NoError(t, err)
		assert.Equal(t, int64(len(batch)), n)

		for _, m := range batch {
			got, err := repo.GetFirst(ctx, func(p *Post, h query.GetFirstHelper[Post]) {
				h.Where().Field(&p.ID).EQ(m.ID)
			})
			require.NoError(t, err)
			assert.Equal(t, m.Title, got.Title)
			assert.Equal(t, m.Published, got.Published)
		}
	})
}

// TestInsertMany_Empty — пустой срез не делает запроса и не вызывает хуки.
func TestInsertMany_Empty(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		var beforeCalls, afterCalls int
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
			WithBeforeInsertMany(func(_ context.Context, _ []*Post) error {
				beforeCalls++
				return nil
			}).
			WithAfterInsertMany(func(_ context.Context, _ []*Post) error {
				afterCalls++
				return nil
			}).
			Build()
		require.NoError(t, err)

		ctx, cancel := testCtx(t)
		defer cancel()

		n, err := repo.InsertMany(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(0), n)
		assert.Zero(t, beforeCalls)
		assert.Zero(t, afterCalls)
	})
}

// TestInsertMany_Returning — модели помеченные ReturnedOnInsert получают
// серверные значения обратно по позиции.
func TestInsertMany_Returning(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		setupReturningSchema(t)
		repo := newReturningRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		batch := []*returningModel{
			{Title: "ret-1"},
			{Title: "ret-2"},
			{Title: "ret-3"},
		}
		n, err := repo.InsertMany(ctx, batch)
		require.NoError(t, err)
		assert.Equal(t, int64(len(batch)), n)

		seen := make(map[uuid.UUID]struct{}, len(batch))
		for _, m := range batch {
			assert.NotEqual(t, uuid.Nil, m.ID, "ID must arrive via RETURNING")
			assert.False(t, m.CreatedAt.IsZero(), "CreatedAt must arrive via RETURNING")
			_, dup := seen[m.ID]
			assert.False(t, dup, "each row must get its own ID")
			seen[m.ID] = struct{}{}
		}
	})
}

// TestInsertMany_Hooks — before/after получают весь срез одним вызовом; ошибка
// из Before отменяет SQL, ошибка из After возвращается после записи.
func TestInsertMany_Hooks(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		ctx, cancel := testCtx(t)
		defer cancel()

		var beforeSeen, afterSeen int
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
			WithBeforeInsertMany(func(_ context.Context, models []*Post) error {
				beforeSeen = len(models)
				return nil
			}).
			WithAfterInsertMany(func(_ context.Context, models []*Post) error {
				afterSeen = len(models)
				return nil
			}).
			Build()
		require.NoError(t, err)

		author := seed.users[0].ID
		batch := []*Post{
			{ID: uuid.New(), UserID: author, Title: "h1", Content: "c", CreatedAt: nowUTC()},
			{ID: uuid.New(), UserID: author, Title: "h2", Content: "c", CreatedAt: nowUTC()},
		}
		_, err = repo.InsertMany(ctx, batch)
		require.NoError(t, err)
		assert.Equal(t, 2, beforeSeen)
		assert.Equal(t, 2, afterSeen)
	})
}

// TestInsertMany_LargeBatch_ChunksTransparently — батч превышает PG limit 65535
// placeholder'ов; executor разбивает на чанки, результат — все строки в базе.
// Используем returning_demo (3 колонки = 3 placeholder'а на строку), чтобы набрать
// достаточно строк при разумном весе.
func TestInsertMany_LargeBatch_ChunksTransparently(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		setupReturningSchema(t)
		// На returning_demo в INSERT идут только (id, title, created_at) —
		// active-cols ≤ 3, значит chunkSize ≥ 21000. Достаточно прислать
		// чуть больше, чтобы стабильно уйти на второй чанк.
		repo, err := gerpo.New[returningModel]().
			DB(ab.adapter).
			Table("returning_demo").
			Columns(func(m *returningModel, c *gerpo.ColumnBuilder[returningModel]) {
				c.Field(&m.ID).ReadOnly().ReturnedOnInsert()
				c.Field(&m.Title)
				c.Field(&m.CreatedAt).ReadOnly().ReturnedOnInsert()
				c.Field(&m.UpdatedAt).OmitOnInsert()
			}).
			Build()
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		const total = 25000 // пересекает один chunk-boundary
		batch := make([]*returningModel, total)
		for i := range batch {
			batch[i] = &returningModel{Title: "bulk"}
		}

		n, err := repo.InsertMany(ctx, batch)
		require.NoError(t, err)
		assert.Equal(t, int64(total), n)

		// RETURNING должен был раздать уникальные UUID.
		seen := make(map[uuid.UUID]struct{}, total)
		for _, m := range batch {
			require.NotEqual(t, uuid.Nil, m.ID)
			_, dup := seen[m.ID]
			require.False(t, dup)
			seen[m.ID] = struct{}{}
		}
	})
}
