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

// returningModel is a Post-like fixture that exercises every RETURNING shape:
//   - ID:          DB-side default (gen_random_uuid()) — needs INSERT RETURNING.
//   - CreatedAt:   DB-side default (NOW()) — also INSERT RETURNING.
//   - Title:       app-supplied, no RETURNING.
//   - UpdatedAt:   trigger fires on UPDATE — UPDATE RETURNING.
//
// schema.sql does NOT yet have such a table. We CREATE / DROP it inside the test
// so the integration suite stays self-contained and the regular schema.sql is
// untouched. Drop runs unconditionally via t.Cleanup so a failing test does not
// leak state.
type returningModel struct {
	ID        uuid.UUID
	Title     string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

const returningSchemaUp = `
CREATE TABLE IF NOT EXISTS returning_demo (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);

CREATE OR REPLACE FUNCTION returning_demo_touch_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS returning_demo_touch ON returning_demo;
CREATE TRIGGER returning_demo_touch
    BEFORE UPDATE ON returning_demo
    FOR EACH ROW EXECUTE PROCEDURE returning_demo_touch_updated_at();
`

const returningSchemaDown = `
DROP TRIGGER IF EXISTS returning_demo_touch ON returning_demo;
DROP FUNCTION IF EXISTS returning_demo_touch_updated_at();
DROP TABLE IF EXISTS returning_demo;
`

func setupReturningSchema(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := pgx5Pool.Exec(ctx, returningSchemaUp)
	require.NoError(t, err, "create returning_demo table")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pgx5Pool.Exec(ctx, returningSchemaDown)
	})
}

func newReturningRepo(t *testing.T, ab adapterBundle) gerpo.Repository[returningModel] {
	t.Helper()
	repo, err := gerpo.New[returningModel]().
		Adapter(ab.adapter).
		Table("returning_demo").
		Columns(func(m *returningModel, c *gerpo.ColumnBuilder[returningModel]) {
			c.Field(&m.ID).ReadOnly().ReturnedOnInsert()        // PK with DB DEFAULT gen_random_uuid()
			c.Field(&m.Title)                                   // app-supplied
			c.Field(&m.CreatedAt).ReadOnly().ReturnedOnInsert() // DB DEFAULT NOW()
			c.Field(&m.UpdatedAt).OmitOnInsert().ReturnedOnUpdate()
		}).
		Build()
	require.NoError(t, err)
	return repo
}

// TestReturning_Insert_FillsServerGenerated — ReturnedOnInsert columns come
// back from RETURNING and land in the model in-place.
func TestReturning_Insert_FillsServerGenerated(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		setupReturningSchema(t)
		repo := newReturningRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		m := &returningModel{Title: "hello"}
		require.NoError(t, repo.Insert(ctx, m))

		assert.NotEqual(t, uuid.Nil, m.ID, "ID must be filled from DEFAULT gen_random_uuid()")
		assert.False(t, m.CreatedAt.IsZero(), "CreatedAt must be filled from DEFAULT NOW()")
		assert.Nil(t, m.UpdatedAt, "UpdatedAt is not RETURNING-on-insert; must remain nil")
	})
}

// TestReturning_Update_FillsTriggerColumn — UpdatedAt is set by a trigger; the
// post-update value must arrive back into the model.
func TestReturning_Update_FillsTriggerColumn(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		setupReturningSchema(t)
		repo := newReturningRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		m := &returningModel{Title: "before"}
		require.NoError(t, repo.Insert(ctx, m))
		require.Nil(t, m.UpdatedAt, "no UpdatedAt before any UPDATE")

		m.Title = "after"
		n, err := repo.Update(ctx, m, func(base *returningModel, h query.UpdateHelper[returningModel]) {
			h.Where().Field(&base.ID).EQ(m.ID)
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), n)
		require.NotNil(t, m.UpdatedAt, "trigger should have set updated_at; RETURNING brings it back")
		assert.WithinDuration(t, time.Now().UTC(), *m.UpdatedAt, 5*time.Second)
	})
}

// TestReturning_PerRequest_NarrowsList — h.Returning(&m.ID) overrides the
// repo-level returning set: only ID comes back, CreatedAt stays at zero
// despite being marked ReturnedOnInsert at the repo.
func TestReturning_PerRequest_NarrowsList(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		setupReturningSchema(t)
		repo := newReturningRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		m := &returningModel{Title: "narrow"}
		require.NoError(t, repo.Insert(ctx, m, func(base *returningModel, h query.InsertHelper[returningModel]) {
			h.Returning(&base.ID)
		}))

		assert.NotEqual(t, uuid.Nil, m.ID, "ID requested explicitly — must be filled")
		assert.True(t, m.CreatedAt.IsZero(),
			"CreatedAt was excluded by per-request Returning(); model field stays zero")
	})
}

// TestReturning_PerRequest_DisablesEntirely — h.Returning() with no args turns
// RETURNING off for the call.
func TestReturning_PerRequest_DisablesEntirely(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		setupReturningSchema(t)
		repo := newReturningRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		m := &returningModel{Title: "no-returning"}
		require.NoError(t, repo.Insert(ctx, m, func(_ *returningModel, h query.InsertHelper[returningModel]) {
			h.Returning() // disable — no pointer args, no base-model resolve needed
		}))

		assert.Equal(t, uuid.Nil, m.ID, "RETURNING disabled — ID must stay zero (server value not read back)")
		assert.True(t, m.CreatedAt.IsZero(), "CreatedAt similarly")
	})
}
