package gerpo

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/insei/gerpo/executor"
	extypes "github.com/insei/gerpo/executor/types"
	"github.com/insei/gerpo/query"
)

type softModel struct {
	ID        int
	DeletedAt *time.Time
}

// nopAdapter is the smallest possible executor.DBAdapter. It is never
// invoked from these tests — they exercise NewBuilder + SoftDeletion at
// configuration time only.
type nopAdapter struct{}

func (nopAdapter) ExecContext(context.Context, string, ...any) (extypes.Result, error) {
	return nil, nil
}
func (nopAdapter) QueryContext(context.Context, string, ...any) (extypes.Rows, error) {
	return nil, nil
}
func (nopAdapter) BeginTx(context.Context) (extypes.Tx, error) { return nil, nil }

func newSoftRepoBuilder() ColumnsAppender[softModel] {
	return NewBuilder[softModel]().DB(executor.DBAdapter(nopAdapter{})).Table("soft_users")
}

// TestWithSoftDeletion_TypeMismatch_FailsAtBuild proves that returning a value
// whose type does not match the target field is rejected by Build() instead of
// panicking later at the first Delete call.
func TestWithSoftDeletion_TypeMismatch_FailsAtBuild(t *testing.T) {
	_, err := newSoftRepoBuilder().
		Columns(func(m *softModel, c *ColumnBuilder[softModel]) {
			c.Field(&m.ID).AsColumn()
			c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
		}).
		WithSoftDeletion(func(m *softModel, b *SoftDeletionBuilder[softModel]) {
			// time.Time is NOT assignable to *time.Time — must be flagged.
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				return time.Now().UTC()
			})
		}).
		Build()
	if err == nil {
		t.Fatal("expected Build() to return an error for soft-delete type mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "soft delete") || !strings.Contains(err.Error(), "not assignable") {
		t.Fatalf("expected error to mention soft delete and type mismatch, got: %v", err)
	}
}

// TestWithSoftDeletion_NilForPointer_OK ensures that returning nil for a
// pointer-typed field stays valid (it's a legitimate way to clear the marker).
func TestWithSoftDeletion_NilForPointer_OK(t *testing.T) {
	_, err := newSoftRepoBuilder().
		Columns(func(m *softModel, c *ColumnBuilder[softModel]) {
			c.Field(&m.ID).AsColumn()
			c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
		}).
		WithSoftDeletion(func(m *softModel, b *SoftDeletionBuilder[softModel]) {
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				return (*time.Time)(nil)
			})
		}).
		Build()
	if err != nil {
		t.Fatalf("expected Build() to succeed for nil pointer value, got: %v", err)
	}
}

// TestWithSoftDeletion_PanicInProbe_BecomesError proves that a panic inside the
// user-provided function is caught and returned as an error rather than
// crashing the build.
func TestWithSoftDeletion_PanicInProbe_BecomesError(t *testing.T) {
	_, err := newSoftRepoBuilder().
		Columns(func(m *softModel, c *ColumnBuilder[softModel]) {
			c.Field(&m.ID).AsColumn()
			c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
		}).
		WithSoftDeletion(func(m *softModel, b *SoftDeletionBuilder[softModel]) {
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				panic("boom")
			})
		}).
		Build()
	if err == nil {
		t.Fatal("expected Build() to surface the panic as an error, got nil")
	}
	if !strings.Contains(err.Error(), "panicked during type probe") {
		t.Fatalf("expected error to mention the probe panic, got: %v", err)
	}
}

// TestWithSoftDeletion_HappyPath ensures the type probe does not break valid
// configurations.
func TestWithSoftDeletion_HappyPath(t *testing.T) {
	_, err := newSoftRepoBuilder().
		Columns(func(m *softModel, c *ColumnBuilder[softModel]) {
			c.Field(&m.ID).AsColumn()
			c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
		}).
		WithSoftDeletion(func(m *softModel, b *SoftDeletionBuilder[softModel]) {
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				now := time.Now().UTC()
				return &now
			})
		}).
		Build()
	if err != nil {
		t.Fatalf("expected Build() to succeed for happy-path configuration, got: %v", err)
	}
}

// mockResult / mockDB — minimal doubles used by the happy-path Delete test.
type softMockResult struct{ rows int64 }

func (r softMockResult) RowsAffected() (int64, error) { return r.rows, nil }

type softMockAdapter struct {
	execSQL  []string
	execArgs [][]any
	rows     int64
}

func (a *softMockAdapter) ExecContext(_ context.Context, sql string, args ...any) (extypes.Result, error) {
	a.execSQL = append(a.execSQL, sql)
	a.execArgs = append(a.execArgs, args)
	return softMockResult{rows: a.rows}, nil
}
func (a *softMockAdapter) QueryContext(context.Context, string, ...any) (extypes.Rows, error) {
	return nil, nil
}
func (a *softMockAdapter) BeginTx(context.Context) (extypes.Tx, error) { return nil, nil }

// TestWithSoftDeletion_Delete_ExecutesUpdate — exercises the Delete path on a
// repo with soft deletion configured. The mock adapter records the SQL, so we
// can prove it was UPDATE ... SET deleted_at rather than DELETE FROM.
func TestWithSoftDeletion_Delete_ExecutesUpdate(t *testing.T) {
	adapter := &softMockAdapter{rows: 1}
	repo, err := NewBuilder[softModel]().
		DB(executor.DBAdapter(adapter)).
		Table("soft_users").
		Columns(func(m *softModel, c *ColumnBuilder[softModel]) {
			c.Field(&m.ID).AsColumn()
			c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
		}).
		WithSoftDeletion(func(m *softModel, b *SoftDeletionBuilder[softModel]) {
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				now := time.Now().UTC()
				return &now
			})
		}).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	n, err := repo.Delete(context.Background(), func(m *softModel, h query.DeleteHelper[softModel]) {
		h.Where().Field(&m.ID).EQ(7)
	})
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 row affected, got %d", n)
	}
	if len(adapter.execSQL) != 1 {
		t.Fatalf("expected one exec, got %d", len(adapter.execSQL))
	}
	if !strings.HasPrefix(adapter.execSQL[0], "UPDATE ") {
		t.Fatalf("soft delete must UPDATE, got: %s", adapter.execSQL[0])
	}
	if !strings.Contains(adapter.execSQL[0], "deleted_at") {
		t.Fatalf("expected deleted_at in SET clause, got: %s", adapter.execSQL[0])
	}
}

// TestWithSoftDeletion_Delete_NoRows_ReturnsNotFound — UPDATE touches zero rows
// → the repo returns ErrNotFound wrapped with the context.
func TestWithSoftDeletion_Delete_NoRows_ReturnsNotFound(t *testing.T) {
	adapter := &softMockAdapter{rows: 0}
	repo, err := NewBuilder[softModel]().
		DB(executor.DBAdapter(adapter)).
		Table("soft_users").
		Columns(func(m *softModel, c *ColumnBuilder[softModel]) {
			c.Field(&m.ID).AsColumn()
			c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
		}).
		WithSoftDeletion(func(m *softModel, b *SoftDeletionBuilder[softModel]) {
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				now := time.Now().UTC()
				return &now
			})
		}).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	_, err = repo.Delete(context.Background(), func(m *softModel, h query.DeleteHelper[softModel]) {
		h.Where().Field(&m.ID).EQ(7)
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestWithSoftDeletion_DisallowedField_FailsAtBuild keeps the prior contract
// alive: marking a column with WithUpdateProtection makes it ineligible.
func TestWithSoftDeletion_DisallowedField_FailsAtBuild(t *testing.T) {
	_, err := newSoftRepoBuilder().
		Columns(func(m *softModel, c *ColumnBuilder[softModel]) {
			c.Field(&m.ID).AsColumn()
			c.Field(&m.DeletedAt).AsColumn().WithUpdateProtection()
		}).
		WithSoftDeletion(func(m *softModel, b *SoftDeletionBuilder[softModel]) {
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				now := time.Now().UTC()
				return &now
			})
		}).
		Build()
	if err == nil || !errors.Is(err, err) /* keep linter happy */ {
		t.Fatal("expected Build() to reject soft-delete on update-protected column")
	}
	if !strings.Contains(err.Error(), "allowed update action") {
		t.Fatalf("expected error to mention update action, got: %v", err)
	}
}
