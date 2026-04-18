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

// TestTx_Commit — изменения, сделанные в транзакции и закоммиченные, видны извне.
func TestTx_Commit(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		tx, err := ab.adapter.BeginTx(ctx)
		require.NoError(t, err)

		txRepo := repo.Tx(tx)
		p := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "tx-commit",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		require.NoError(t, txRepo.Insert(ctx, &p))
		require.NoError(t, tx.Commit())

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(p.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, "tx-commit", got.Title)
	})
}

// TestTx_Rollback — после Rollback изменения не применяются.
func TestTx_Rollback(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		tx, err := ab.adapter.BeginTx(ctx)
		require.NoError(t, err)

		txRepo := repo.Tx(tx)
		p := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "tx-rollback",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		require.NoError(t, txRepo.Insert(ctx, &p))
		require.NoError(t, tx.Rollback())

		_, err = repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(p.ID)
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, gerpo.ErrNotFound), "after rollback the row must not exist")
	})
}

// TestTx_Isolation — до Commit другие коннекты не видят вставленную запись
// (PostgreSQL default isolation = Read Committed).
func TestTx_Isolation(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		tx, err := ab.adapter.BeginTx(ctx)
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()

		txRepo := repo.Tx(tx)
		p := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "tx-isolation",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		require.NoError(t, txRepo.Insert(ctx, &p))

		// Чтение через отдельный pool (pgx5Pool) — другой коннект, не в транзакции.
		var exists bool
		err = pgx5Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`, p.ID).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "uncommitted row must not be visible to other connections")

		// Внутри транзакции — видна.
		got, err := txRepo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(p.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, p.ID, got.ID, "row must be visible inside its own transaction")
	})
}

// TestTx_RollbackUnlessCommitted_AfterCommit — после Commit вызов
// RollbackUnlessCommitted должен быть no-op (не возвращать ошибку).
func TestTx_RollbackUnlessCommitted_AfterCommit(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		tx, err := ab.adapter.BeginTx(ctx)
		require.NoError(t, err)

		txRepo := repo.Tx(tx)
		p := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "tx-roll-unless",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		require.NoError(t, txRepo.Insert(ctx, &p))
		require.NoError(t, tx.Commit())
		// После Commit вызов должен быть безопасным.
		require.NoError(t, tx.RollbackUnlessCommitted(), "RollbackUnlessCommitted after Commit must be a no-op")

		got, err := repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(p.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, p.ID, got.ID)
	})
}

// TestTx_RollbackUnlessCommitted_WithoutCommit — без Commit() метод откатывает.
func TestTx_RollbackUnlessCommitted_WithoutCommit(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		tx, err := ab.adapter.BeginTx(ctx)
		require.NoError(t, err)

		txRepo := repo.Tx(tx)
		p := Post{
			ID:        uuid.New(),
			UserID:    seed.users[0].ID,
			Title:     "tx-roll-unless-no-commit",
			Content:   "c",
			CreatedAt: time.Now().UTC(),
		}
		require.NoError(t, txRepo.Insert(ctx, &p))
		require.NoError(t, tx.RollbackUnlessCommitted())

		_, err = repo.GetFirst(ctx, func(m *Post, h query.GetFirstHelper[Post]) {
			h.Where().Field(&m.ID).EQ(p.ID)
		})
		assert.True(t, errors.Is(err, gerpo.ErrNotFound))
	})
}
