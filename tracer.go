package gerpo

import "context"

// SpanEnd is returned by a Tracer to signal the end of a span. The wrapped
// repository operation reports its terminal error (or nil on success) through it.
type SpanEnd func(err error)

// SpanInfo carries the metadata available when a Repository opens a span:
// the operation name (already prefixed with "gerpo.") and the table the
// repository is bound to. The struct is open for additive extension —
// fields may be added later without breaking existing Tracer implementations.
type SpanInfo struct {
	// Op is the prefixed operation name, e.g. "gerpo.GetFirst".
	Op string

	// Table is the table the Repository was built with. Empty only if
	// the repository is misconfigured (Build() guards against that).
	Table string
}

// Tracer is gerpo's tracing hook. Repository wraps every public operation —
// GetFirst, GetList, Count, Insert, Update, Delete — in a span by calling
// the configured Tracer at entry and the returned SpanEnd at exit.
//
// gerpo deliberately does not import any tracing library — write a thin
// adapter against your tracer of choice (OpenTelemetry, Datadog, ...).
//
// Decide per implementation whether to put SpanInfo.Table into the span
// name (e.g. "gerpo.GetFirst users") or into an attribute (e.g.
// db.sql.table = "users"). The OpenTelemetry semantic conventions prefer
// the attribute form.
type Tracer = func(ctx context.Context, span SpanInfo) (context.Context, SpanEnd)

// noopSpanEnd is used when no tracer is configured so call sites can stay branch-free.
func noopSpanEnd(error) {}
