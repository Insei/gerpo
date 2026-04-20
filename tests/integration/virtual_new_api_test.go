//go:build integration

package integration

import (
	"testing"

	"github.com/insei/gerpo"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
	"github.com/insei/gerpo/virtual"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newUserRepoComputeDropIn assembles a User repository using Compute(sql) instead
// of the deprecated WithSQL(func(ctx)). The setup is otherwise identical to
// newUserRepo, so any behavioral divergence indicates a regression.
func newUserRepoComputeDropIn(t *testing.T, ab adapterBundle) gerpo.Repository[User] {
	t.Helper()
	repo, err := gerpo.New[User]().
		Adapter(ab.adapter).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
			c.Field(&m.Email)
			c.Field(&m.Age)
			c.Field(&m.CreatedAt).OmitOnUpdate()
			c.Field(&m.UpdatedAt).OmitOnInsert()
			c.Field(&m.DeletedAt).OmitOnInsert()
			c.Field(&m.PostCount).AsVirtual().Compute("COALESCE(COUNT(posts.id), 0)")
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			h.LeftJoinOn("posts", "posts.user_id = users.id")
			h.GroupBy(&m.ID, &m.Name, &m.Email, &m.Age, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
			h.Where().Field(&m.DeletedAt).EQ(nil)
		}).
		Build()
	if err != nil {
		t.Fatalf("build user repo (Compute drop-in): %v", err)
	}
	return repo
}

// newUserRepoComputeWithArgs uses Compute(sql, args...) so the SELECT clause
// contains a bound parameter that travels through the gerpo args pipeline. The
// virtual column counts only the posts whose title matches titleLike.
func newUserRepoComputeWithArgs(t *testing.T, ab adapterBundle, titleLike string) gerpo.Repository[User] {
	t.Helper()
	repo, err := gerpo.New[User]().
		Adapter(ab.adapter).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
			c.Field(&m.Email)
			c.Field(&m.Age)
			c.Field(&m.CreatedAt).OmitOnUpdate()
			c.Field(&m.UpdatedAt).OmitOnInsert()
			c.Field(&m.DeletedAt).OmitOnInsert()
			c.Field(&m.PostCount).AsVirtual().Compute(
				"SELECT count(*) FROM posts WHERE posts.user_id = users.id AND posts.title LIKE ?",
				titleLike,
			)
		}).
		Build()
	if err != nil {
		t.Fatalf("build user repo (Compute+args): %v", err)
	}
	return repo
}

// newUserRepoAggregate marks PostCount as Aggregate without registering an
// explicit Filter override. The repo can still SELECT the column, but any WHERE
// condition on it must be rejected by the WhereBuilder guard.
func newUserRepoAggregate(t *testing.T, ab adapterBundle) gerpo.Repository[User] {
	t.Helper()
	repo, err := gerpo.New[User]().
		Adapter(ab.adapter).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
			c.Field(&m.Email)
			c.Field(&m.Age)
			c.Field(&m.CreatedAt).OmitOnUpdate()
			c.Field(&m.UpdatedAt).OmitOnInsert()
			c.Field(&m.DeletedAt).OmitOnInsert()
			c.Field(&m.PostCount).AsVirtual().
				Aggregate().
				Compute("COALESCE(COUNT(posts.id), 0)")
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			h.LeftJoinOn("posts", "posts.user_id = users.id")
			h.GroupBy(&m.ID, &m.Name, &m.Email, &m.Age, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
		}).
		Build()
	if err != nil {
		t.Fatalf("build user repo (Aggregate): %v", err)
	}
	return repo
}

// TestVirtual_NewAPI_ComputeDropIn — Compute(sql) yields the same observable result
// as the deprecated WithSQL(func(ctx) string) on every adapter.
func TestVirtual_NewAPI_ComputeDropIn(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepoComputeDropIn(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.ID).EQ(seed.users[0].ID)
		})
		require.NoError(t, err)
		assert.Equal(t, 3, got.PostCount, "Compute must produce the same count as WithSQL")
	})
}

// TestVirtual_NewAPI_ComputeWithBoundArgs — bound args declared in Compute travel
// through the gerpo args pipeline and end up correctly positioned in the final
// SQL. user[1] owns posts 3..5 ("Post 3", "Post 4", "Post 5") — matching "%Post 4%"
// produces exactly one row.
func TestVirtual_NewAPI_ComputeWithBoundArgs(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepoComputeWithArgs(t, ab, "%Post 4%")
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.ID).EQ(seed.users[1].ID)
		})
		require.NoError(t, err)
		assert.Equal(t, 1, got.PostCount,
			"Compute bound arg must be applied per row; only 'Post 4' matches '%%Post 4%%' for user[1]")
	})
}

// TestVirtual_NewAPI_AggregateRejectsAutoFilter — the WhereBuilder must reject any
// WHERE condition on an aggregate-marked virtual column when no Filter override
// is registered for that operator.
func TestVirtual_NewAPI_AggregateRejectsAutoFilter(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newUserRepoAggregate(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		_, err := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.PostCount).GT(0)
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "aggregate",
			"aggregate guard must surface a clear error mentioning 'aggregate'")
	})
}

// TestVirtual_NewAPI_AggregateAcceptsExplicitFilter — once the user registers a
// Filter override for an operator, the aggregate guard lets the request through.
// The fragment `1 = ?` is intentionally trivial: we only verify that the override
// is honored by the guard, not that COUNT(...) makes sense in WHERE (it does not
// — gerpo deliberately leaves SQL validity to the override author).
func TestVirtual_NewAPI_AggregateAcceptsExplicitFilter(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo, err := gerpo.New[User]().
			Adapter(ab.adapter).
			Table("users").
			Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
				c.Field(&m.ID).OmitOnUpdate()
				c.Field(&m.Name)
				c.Field(&m.Email)
				c.Field(&m.Age)
				c.Field(&m.CreatedAt).OmitOnUpdate()
				c.Field(&m.UpdatedAt).OmitOnInsert()
				c.Field(&m.DeletedAt).OmitOnInsert()
				c.Field(&m.PostCount).AsVirtual().
					Aggregate().
					Compute("COALESCE(COUNT(posts.id), 0)").
					Filter(types.OperationEQ, virtual.Bound{SQL: "1 = ?"})
			}).
			WithQuery(func(m *User, h query.PersistentHelper[User]) {
				h.LeftJoinOn("posts", "posts.user_id = users.id")
				h.GroupBy(&m.ID, &m.Name, &m.Email, &m.Age, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
			}).
			Build()
		require.NoError(t, err)

		ctx, cancel := testCtx(t)
		defer cancel()

		// EQ(1) → "1 = 1" — always true, returns first user (deterministic ordering not
		// relied upon, we just assert no error and a non-zero row).
		got, err := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.PostCount).EQ(1)
			h.Where().Field(&m.ID).EQ(seed.users[0].ID)
		})
		require.NoError(t, err, "Filter override must satisfy the aggregate guard")
		assert.Equal(t, seed.users[0].ID, got.ID)

		// Sanity: the override did not register GT, so GT must still error.
		_, err = repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
			h.Where().Field(&m.PostCount).GT(0)
		})
		require.Error(t, err)
	})
}
