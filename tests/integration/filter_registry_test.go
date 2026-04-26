//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/filters"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// statusKind is a string-alias type. Before the registry-based filter resolver
// the runtime would reject string-alias values that don't match the field's
// reflect.Type exactly: pre-registry, type-equality lived in
// types/filters.go::AddFilterFnArgs and broke any non-builtin string type.
type statusKind string

const (
	statusActive   statusKind = "active"
	statusArchived statusKind = "archived"
)

type statusModel struct {
	ID     uuid.UUID
	Status statusKind
	Note   string
}

const statusSchemaUp = `
CREATE TABLE IF NOT EXISTS filter_registry_demo (
    id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status TEXT NOT NULL,
    note   TEXT NOT NULL DEFAULT ''
);
`

const statusSchemaDown = `DROP TABLE IF EXISTS filter_registry_demo;`

func setupStatusSchema(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := pgx5Pool.Exec(ctx, statusSchemaUp)
	require.NoError(t, err, "create filter_registry_demo table")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pgx5Pool.Exec(ctx, statusSchemaDown)
	})
}

// TestFilterRegistry_StringAliasFilters_RoundTrip — registering a string-alias
// type (statusKind) lets WHERE accept literal-string values without runtime
// type-equality complaints; the registry path goes through
// SQLFilterManager.AddFilterFnArgsRaw which omits the strict guard.
func TestFilterRegistry_StringAliasFilters_RoundTrip(t *testing.T) {
	restore := filters.Snapshot()
	t.Cleanup(restore)

	// Register the alias with EQ + In. Without this it would fall through to
	// the String KindBucket and the legacy gard would still reject the value.
	filters.Registry.Register(statusKind("")).
		Allow(types.OperationEQ, types.OperationIn)

	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		setupStatusSchema(t)

		repo, err := gerpo.New[statusModel]().
			Adapter(ab.adapter).
			Table("filter_registry_demo").
			Columns(func(m *statusModel, c *gerpo.ColumnBuilder[statusModel]) {
				c.Field(&m.ID).ReadOnly().ReturnedOnInsert()
				c.Field(&m.Status)
				c.Field(&m.Note)
			}).
			Build()
		require.NoError(t, err, "build status repo")

		ctx, cancel := testCtx(t)
		defer cancel()

		require.NoError(t, repo.Insert(ctx, &statusModel{Status: statusActive, Note: "first"}))
		require.NoError(t, repo.Insert(ctx, &statusModel{Status: statusArchived, Note: "second"}))
		require.NoError(t, repo.Insert(ctx, &statusModel{Status: statusActive, Note: "third"}))

		// EQ with the alias-typed value.
		gotEQ, err := repo.GetList(ctx, func(m *statusModel, h query.GetListHelper[statusModel]) {
			h.Where().Field(&m.Status).EQ(statusActive)
		})
		require.NoError(t, err, "GetList EQ")
		assert.Len(t, gotEQ, 2, "EQ should match two rows")

		// EQ with a literal string — used to be rejected by the type-equality
		// guard. With the registry path it works.
		gotLit, err := repo.GetList(ctx, func(m *statusModel, h query.GetListHelper[statusModel]) {
			h.Where().Field(&m.Status).EQ("archived")
		})
		require.NoError(t, err, "GetList EQ with literal string")
		assert.Len(t, gotLit, 1, "EQ literal should match one row")

		// In with a slice of aliases.
		gotIn, err := repo.GetList(ctx, func(m *statusModel, h query.GetListHelper[statusModel]) {
			h.Where().Field(&m.Status).In(statusActive, statusArchived)
		})
		require.NoError(t, err, "GetList In")
		assert.Len(t, gotIn, 3, "In should match all rows")
	})
}

// TestFilterRegistry_TimeOverride_AppliesCustomSQL — overriding Time.EQ with a
// Bound spec that compares only the date part demonstrates that user SQL gets
// installed and routed through the column without disturbing other operators.
func TestFilterRegistry_TimeOverride_AppliesCustomSQL(t *testing.T) {
	restore := filters.Snapshot()
	t.Cleanup(restore)

	// Compare only the date — ignore the time-of-day portion.
	filters.Registry.Time.Override(types.OperationEQ, filters.Bound{
		SQL: "DATE_TRUNC('day', created_at) = DATE_TRUNC('day', CAST(? AS timestamptz))",
	})

	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		// All seeded posts share the same calendar day in CreatedAt; pick any
		// CreatedAt and shift the time-of-day forward — the EQ override should
		// still find the rows.
		base := seed.posts[0].CreatedAt
		shifted := time.Date(base.Year(), base.Month(), base.Day(), 23, 59, 59, 0, base.Location())

		gotShifted, err := repo.GetList(ctx, func(m *Post, h query.GetListHelper[Post]) {
			h.Where().Field(&m.CreatedAt).EQ(shifted)
		})
		require.NoError(t, err, "GetList CreatedAt EQ with override")
		assert.NotEmpty(t, gotShifted,
			"override compares date-only — rows from the same day must match despite time shift")

		// LT/LTE/GT/GTE remain stock so the regular range query still works.
		gotRange, err := repo.GetList(ctx, func(m *Post, h query.GetListHelper[Post]) {
			h.Where().Field(&m.CreatedAt).GT(base.Add(-time.Hour))
		})
		require.NoError(t, err, "GetList CreatedAt GT")
		assert.NotEmpty(t, gotRange, "stock GT still works after overriding EQ")
	})
}
