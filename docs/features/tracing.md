# Tracing

`gerpo.WithTracer(fn)` installs a hook that the **Repository** opens around every public operation (`GetFirst`, `GetList`, `Count`, `Insert`, `Update`, `Delete`). gerpo deliberately does not pull a tracing library into its dependencies — adapt the hook to whatever your stack uses.

The span is opened at the Repository layer (not the Executor), so the hook sees both the repository method name and the bound table — exactly what you need to identify *which* repository's request you are looking at in a trace explorer.

## The hook

```go
type SpanEnd func(err error)

type SpanInfo struct {
    Op    string // e.g. "gerpo.GetFirst"
    Table string // table the Repository was built with
}

type Tracer = func(ctx context.Context, span SpanInfo) (context.Context, SpanEnd)
```

- `Op` is the operation name, prefixed with `gerpo.` — one of `gerpo.GetFirst`, `gerpo.GetList`, `gerpo.Count`, `gerpo.Insert`, `gerpo.Update`, `gerpo.Delete`.
- `Table` is the table the Repository is bound to. Decide per implementation whether to put it into the span name (`"gerpo.GetFirst users"`) or into a span attribute (`db.sql.table = "users"`). The OpenTelemetry semantic conventions prefer the attribute form.
- The returned `context.Context` is propagated downstream — child operations (driver calls, cache reads) land inside the same span.
- `SpanEnd` is called once on the operation's terminal error (or `nil` on success).

## Wiring with OpenTelemetry

```go
import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"

    "github.com/insei/gerpo"
    "github.com/insei/gerpo/executor/adapters/pgx5"
)

func otelTracer() gerpo.Tracer {
    tr := otel.Tracer("gerpo")
    return func(ctx context.Context, span gerpo.SpanInfo) (context.Context, gerpo.SpanEnd) {
        ctx, otelSpan := tr.Start(ctx, span.Op,
            trace.WithAttributes(
                attribute.String("db.system", "postgresql"),
                attribute.String("db.sql.table", span.Table),
            ),
        )
        return ctx, func(err error) {
            if err != nil {
                otelSpan.RecordError(err)
                otelSpan.SetStatus(codes.Error, err.Error())
            }
            otelSpan.End()
        }
    }
}

repo, _ := gerpo.New[User]().
    DB(pgx5.NewPoolAdapter(pool)).
    Table("users").
    Columns(...).
    WithTracer(otelTracer()).
    Build()
```

## Wiring with Datadog (`dd-trace-go`)

```go
import (
    "context"

    "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

    "github.com/insei/gerpo"
)

func datadogTracer() gerpo.Tracer {
    return func(ctx context.Context, span gerpo.SpanInfo) (context.Context, gerpo.SpanEnd) {
        ddSpan, ctx := tracer.StartSpanFromContext(ctx, span.Op,
            tracer.ResourceName(span.Op+" "+span.Table),
            tracer.SpanType("sql"),
            tracer.Tag("db.sql.table", span.Table),
        )
        return ctx, func(err error) {
            ddSpan.Finish(tracer.WithError(err))
        }
    }
}
```

## Operation names

| Repository method | Span Op |
|---|---|
| `repo.GetFirst` | `gerpo.GetFirst` |
| `repo.GetList`  | `gerpo.GetList`  |
| `repo.Count`    | `gerpo.Count`    |
| `repo.Insert`   | `gerpo.Insert`   |
| `repo.Update`   | `gerpo.Update`   |
| `repo.Delete`   | `gerpo.Delete`   |

`repo.Tx(tx)` does not open a span — it merely returns a tx-bound repository view; spans appear when the bound repository runs an actual operation.

## Disabled by default

If you don't pass `WithTracer`, the Repository short-circuits on a nil tracer — there is no allocation per call and no behavior change. Tracing is opt-in.

## Logs and metrics

Logging and per-op metrics are not exposed as dedicated hooks today. The recommended pattern is to wrap your `executor.Adapter` with a thin `tracingAdapter` (see [Adapters → Tracing wrapper](adapters.md#why-write-a-custom-adapter)) that captures duration and SQL text, then emits whatever your stack expects.

A natural extension would be `WithLogger` / `WithMetrics` callbacks similar to `WithTracer`. They are intentionally absent for now — open an issue if you have a concrete need.
