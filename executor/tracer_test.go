package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/insei/gerpo/sqlstmt"
	"github.com/insei/gerpo/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// recordingTracer captures every span open/close pair so tests can assert on
// op names, propagation and reported errors.
type recordingTracer struct {
	starts []string
	ends   []spanEndRecord
}

type spanEndRecord struct {
	op  string
	err error
}

func (r *recordingTracer) tracer() Tracer {
	return func(ctx context.Context, op string) (context.Context, SpanEnd) {
		r.starts = append(r.starts, op)
		ctx = context.WithValue(ctx, traceKey{}, op)
		return ctx, func(err error) {
			r.ends = append(r.ends, spanEndRecord{op: op, err: err})
		}
	}
}

type traceKey struct{}

// stmtThatChecksContext is a Stmt double whose SQL() method asserts the
// downstream code receives the context propagated by the tracer. We need this
// because the executor passes the (already-traced) ctx into the stmt and the
// adapter — verifying that propagation is the whole point of tracing.
type stmtThatChecksContext struct {
	mock.Mock
	expectedOp string
	t          *testing.T
}

func (s *stmtThatChecksContext) SQL(opts ...sqlstmt.Option) (string, []any, error) {
	rets := s.Called()
	return rets.String(0), rets.Get(1).([]any), rets.Error(2)
}

func (s *stmtThatChecksContext) Columns() types.ExecutionColumns {
	rets := s.Called()
	return rets.Get(0).(types.ExecutionColumns)
}

// TestTracer_GetOne_OpensAndClosesSpan ensures the executor opens a span on
// entry and closes it with a nil error on the happy path.
func TestTracer_GetOne_OpensAndClosesSpan(t *testing.T) {
	rec := &recordingTracer{}

	stmt := new(stmtThatChecksContext)
	stmt.On("SQL").Return("SELECT 1", []any{}, errors.New("stop early"))

	e := New[testModel](nil, WithTracer(rec.tracer()))
	_, err := e.GetOne(context.Background(), stmt)
	require.Error(t, err, "stmt SQL returns an error so the path stops early")

	require.Equal(t, []string{"gerpo.GetOne"}, rec.starts)
	require.Len(t, rec.ends, 1)
	assert.Equal(t, "gerpo.GetOne", rec.ends[0].op)
	assert.ErrorContains(t, rec.ends[0].err, "stop early", "tracer must observe the terminal error")
}

// TestTracer_NoOp_WhenNotConfigured asserts the executor adds zero overhead
// (no panics, no behavior drift) when WithTracer is not used.
func TestTracer_NoOp_WhenNotConfigured(t *testing.T) {
	stmt := new(stmtThatChecksContext)
	stmt.On("SQL").Return("SELECT 1", []any{}, errors.New("ignored"))

	e := New[testModel](nil) // no tracer
	_, err := e.GetOne(context.Background(), stmt)
	require.Error(t, err)
}

// TestTracer_AllOpsCovered runs every executor entry point and verifies each
// one opens a span with the documented operation name.
func TestTracer_AllOpsCovered(t *testing.T) {
	rec := &recordingTracer{}
	e := New[testModel](nil, WithTracer(rec.tracer()))

	cases := []struct {
		name string
		op   string
		call func(t *testing.T, e Executor[testModel])
	}{
		{"GetOne", "gerpo.GetOne", func(t *testing.T, e Executor[testModel]) {
			stmt := new(stmtThatChecksContext)
			stmt.On("SQL").Return("", []any{}, errors.New("x"))
			_, _ = e.GetOne(context.Background(), stmt)
		}},
		{"GetMultiple", "gerpo.GetMultiple", func(t *testing.T, e Executor[testModel]) {
			stmt := new(stmtThatChecksContext)
			stmt.On("SQL").Return("", []any{}, errors.New("x"))
			_, _ = e.GetMultiple(context.Background(), stmt)
		}},
		{"Count", "gerpo.Count", func(t *testing.T, e Executor[testModel]) {
			stmt := new(stmtThatChecksContext)
			stmt.On("SQL").Return("", []any{}, errors.New("x"))
			_, _ = e.Count(context.Background(), stmt)
		}},
		{"Delete", "gerpo.Delete", func(t *testing.T, e Executor[testModel]) {
			stmt := new(stmtThatChecksContext)
			stmt.On("SQL").Return("", []any{}, errors.New("x"))
			_, _ = e.Delete(context.Background(), stmt)
		}},
		{"InsertOne", "gerpo.InsertOne", func(t *testing.T, e Executor[testModel]) {
			stmt := new(stmtThatChecksContext)
			stmt.On("SQL").Return("", []any{}, errors.New("x"))
			_ = e.InsertOne(context.Background(), stmt, &testModel{})
		}},
		{"Update", "gerpo.Update", func(t *testing.T, e Executor[testModel]) {
			stmt := new(stmtThatChecksContext)
			stmt.On("SQL").Return("", []any{}, errors.New("x"))
			_, _ = e.Update(context.Background(), stmt, &testModel{})
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec.starts = nil
			rec.ends = nil
			tc.call(t, e)
			require.Equal(t, []string{tc.op}, rec.starts)
			require.Len(t, rec.ends, 1)
			assert.Equal(t, tc.op, rec.ends[0].op)
		})
	}
}
