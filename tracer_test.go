package gerpo

import (
	"context"
	"errors"
	"testing"

	"github.com/insei/gerpo/executor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingTracer captures every SpanInfo + terminal error pair so the tests
// can verify span open/close pairing and table propagation.
type recordingTracer struct {
	starts []SpanInfo
	ends   []spanRecord
}

type spanRecord struct {
	op  string
	err error
}

func (r *recordingTracer) tracer() Tracer {
	return func(ctx context.Context, span SpanInfo) (context.Context, SpanEnd) {
		r.starts = append(r.starts, span)
		op := span.Op
		return ctx, func(err error) {
			r.ends = append(r.ends, spanRecord{op: op, err: err})
		}
	}
}

// TestRepository_Tracer_GetFirst — happy path: GetFirst opens a "gerpo.GetFirst"
// span carrying the bound table, and closes it with nil on success.
func TestRepository_Tracer_GetFirst(t *testing.T) {
	type model struct {
		ID int
	}
	rec := &recordingTracer{}
	exec := &MockExecutor[model]{
		GetOneFunc: func(_ context.Context, _ executor.Stmt) (*model, error) {
			return &model{ID: 7}, nil
		},
	}
	repo, err := New[model](exec, "users", func(m *model, c *ColumnBuilder[model]) {
		c.Field(&m.ID)
	}, WithTracer[model](rec.tracer()))
	require.NoError(t, err)

	got, err := repo.GetFirst(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 7, got.ID)

	require.Len(t, rec.starts, 1)
	assert.Equal(t, SpanInfo{Op: "gerpo.GetFirst", Table: "users"}, rec.starts[0])
	require.Len(t, rec.ends, 1)
	assert.Equal(t, "gerpo.GetFirst", rec.ends[0].op)
	assert.NoError(t, rec.ends[0].err, "happy path closes the span with nil")
}

// TestRepository_Tracer_PropagatesError — terminal error reaches SpanEnd so the
// tracer can mark the span as failed.
func TestRepository_Tracer_PropagatesError(t *testing.T) {
	type model struct {
		ID int
	}
	wantErr := errors.New("boom")
	rec := &recordingTracer{}
	exec := &MockExecutor[model]{
		GetOneFunc: func(_ context.Context, _ executor.Stmt) (*model, error) {
			return nil, wantErr
		},
	}
	repo, err := New[model](exec, "users", func(m *model, c *ColumnBuilder[model]) {
		c.Field(&m.ID)
	}, WithTracer[model](rec.tracer()))
	require.NoError(t, err)

	_, err = repo.GetFirst(context.Background())
	require.ErrorIs(t, err, wantErr)
	require.Len(t, rec.ends, 1)
	assert.ErrorIs(t, rec.ends[0].err, wantErr,
		"SpanEnd must observe the terminal error returned to the caller")
}

// TestRepository_Tracer_AllOps — every public Repository method opens a span
// with the documented name. The mock executor returns trivial values; we only
// assert on the recorded SpanInfo.
func TestRepository_Tracer_AllOps(t *testing.T) {
	type model struct {
		ID int
	}
	rec := &recordingTracer{}
	exec := &MockExecutor[model]{
		GetOneFunc:      func(context.Context, executor.Stmt) (*model, error) { return &model{}, nil },
		GetMultipleFunc: func(context.Context, executor.Stmt) ([]*model, error) { return nil, nil },
		CountFunc:       func(context.Context, executor.CountStmt) (uint64, error) { return 0, nil },
		InsertOneFunc:   func(context.Context, executor.Stmt, *model) error { return nil },
		UpdateFunc:      func(context.Context, executor.Stmt, *model) (int64, error) { return 1, nil },
		DeleteFunc:      func(context.Context, executor.CountStmt) (int64, error) { return 1, nil },
	}
	repo, err := New[model](exec, "users", func(m *model, c *ColumnBuilder[model]) {
		c.Field(&m.ID)
	}, WithTracer[model](rec.tracer()))
	require.NoError(t, err)

	ctx := context.Background()
	_, _ = repo.GetFirst(ctx)
	_, _ = repo.GetList(ctx)
	_, _ = repo.Count(ctx)
	_ = repo.Insert(ctx, &model{ID: 1})
	_, _ = repo.Update(ctx, &model{ID: 1})
	_, _ = repo.Delete(ctx)

	want := []string{
		"gerpo.GetFirst",
		"gerpo.GetList",
		"gerpo.Count",
		"gerpo.Insert",
		"gerpo.Update",
		"gerpo.Delete",
	}
	got := make([]string, len(rec.starts))
	for i, s := range rec.starts {
		got[i] = s.Op
		assert.Equal(t, "users", s.Table, "every span must carry the bound table")
	}
	assert.Equal(t, want, got)
	assert.Len(t, rec.ends, len(want), "every opened span must be closed")
}

// TestRepository_Tracer_NoTracerNoOverhead — when WithTracer is not configured
// the repository must still work without panics or behavior drift.
func TestRepository_Tracer_NoTracerNoOverhead(t *testing.T) {
	type model struct {
		ID int
	}
	exec := &MockExecutor[model]{
		GetOneFunc: func(context.Context, executor.Stmt) (*model, error) { return &model{ID: 1}, nil },
	}
	repo, err := New[model](exec, "users", func(m *model, c *ColumnBuilder[model]) {
		c.Field(&m.ID)
	})
	require.NoError(t, err)

	got, err := repo.GetFirst(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, got.ID)
}
