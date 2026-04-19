# Tracing

`executor.WithTracer(fn)` installs a hook that the executor opens around every public operation (`GetOne`, `GetMultiple`, `Count`, `InsertOne`, `Update`, `Delete`). gerpo deliberately does not pull a tracing library into its dependencies — adapt the hook to whatever your stack uses.

## The hook

```go
type SpanEnd func(err error)

type Tracer func(ctx context.Context, op string) (context.Context, SpanEnd)
```

- `op` is the operation name, prefixed with `gerpo.` (`gerpo.GetOne`, `gerpo.Update`, …).
- The returned `context.Context` is propagated downstream — child operations (driver calls, cache reads) land inside the same span.
- `SpanEnd` is called once on the operation's terminal error (or `nil` on success).

## Wiring with OpenTelemetry

```go
import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"

    "github.com/insei/gerpo"
    "github.com/insei/gerpo/executor"
    "github.com/insei/gerpo/executor/adapters/pgx5"
)

func otelTracer() executor.Tracer {
    tr := otel.Tracer("gerpo")
    return func(ctx context.Context, op string) (context.Context, executor.SpanEnd) {
        ctx, span := tr.Start(ctx, op,
            trace.WithAttributes(attribute.String("db.system", "postgresql")),
        )
        return ctx, func(err error) {
            if err != nil {
                span.RecordError(err)
                span.SetStatus(codes.Error, err.Error())
            }
            span.End()
        }
    }
}

repo, _ := gerpo.NewBuilder[User]().
    DB(pgx5.NewPoolAdapter(pool), executor.WithTracer(otelTracer())).
    Table("users").
    Columns(...).
    Build()
```

## Wiring with Datadog (`dd-trace-go`)

```go
import (
    "context"

    "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

    "github.com/insei/gerpo/executor"
)

func datadogTracer() executor.Tracer {
    return func(ctx context.Context, op string) (context.Context, executor.SpanEnd) {
        span, ctx := tracer.StartSpanFromContext(ctx, op,
            tracer.ResourceName(op),
            tracer.SpanType("sql"),
        )
        return ctx, func(err error) {
            span.Finish(tracer.WithError(err))
        }
    }
}
```

## Operation names

| Op | Emitted span |
|---|---|
| `repo.GetFirst` | `gerpo.GetOne` |
| `repo.GetList` | `gerpo.GetMultiple` |
| `repo.Count` | `gerpo.Count` |
| `repo.Insert` | `gerpo.InsertOne` |
| `repo.Update` | `gerpo.Update` |
| `repo.Delete` | `gerpo.Delete` |

## Disabled by default

If you don't pass `WithTracer`, the executor short-circuits on a nil tracer — there is no allocation per call and no behavior change. Tracing is opt-in.

## Logs and metrics

Logging and per-op metrics are not exposed as dedicated hooks today. The recommended pattern is to wrap your `executor.DBAdapter` with a thin `tracingAdapter` (see [Adapters → Tracing wrapper](adapters.md#why-write-a-custom-adapter)) that captures duration and SQL text, then emits whatever your stack expects.

A natural extension would be `WithLogger` / `WithMetrics` callbacks similar to `WithTracer`. They are intentionally absent for now — open an issue if you have a concrete need.
