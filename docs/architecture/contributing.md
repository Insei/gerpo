# Contributing

## Development environment

- Go 1.24+.
- Docker for integration tests.
- `mkdocs-material` if you want to preview the docs locally (`pip install mkdocs-material && mkdocs serve`).

## Check-in loop

```bash
# Unit tests + race detector
go test -race ./...

# Integration tests (PostgreSQL in Docker)
docker compose -f tests/integration/docker-compose.yml up -d
GERPO_INTEGRATION_DB_URL="postgres://gerpo:gerpo@localhost:5433/gerpo?sslmode=disable" \
    go test -tags=integration ./tests/integration/...

# Direct-vs-gerpo allocation benchmarks
go test -bench='^Benchmark(GetFirst|GetList|Count|Insert|Update|Delete)_(Direct|Gerpo)$' \
    -benchmem -run=^$ -count=5 ./tests/

# Formatted summary
GERPO_BENCH_REPORT=1 go test -run=TestCompareDirectVsGerpo -v ./tests/
```

## Code style

- Package names are lowercase and short.
- Interfaces end in `-er` / `-or` when they describe behaviour (`DBAdapter`, `WhereTarget`, `Operation`).
- Every public API ships with godoc — keep the tone concise.
- Generic parameter for the model is `[TModel any]`, consistently.

## Tests

- Unit tests live beside the code (`*_test.go`). Use `go-sqlmock` for `database/sql` paths.
- Integration tests go under `tests/integration/` with the `//go:build integration` tag. They target every adapter in a single run through `forEachAdapter`.
- Benchmarks live in `tests/` (no build tag — `go test -bench=`).

## Commit style

Follow the existing log: lowercase type prefix, imperative subject, optional body with bullet points:

```
perf: replace closures in query/linq builders with structured ops

Where/Order/Exclude/Group/Join no longer store per-condition closures …
```

Common types used in the repo: `feat:`, `fix:`, `perf:`, `test:`, `docs:`, `ci:`, `build:`, `refactor:`, `src:`.

## Opening a PR

A PR to `main` runs three jobs:

- `unit` — build, race detector, full `go test`.
- `integration` — `//go:build integration` against a PG service container.
- `bench-diff` — runs mock benchmarks on head and on base, posts a [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) summary as a PR comment.

`bench-diff` is `allow_failure: true` — a perf regression shows up in the comment but doesn't block merging on its own. Look at the comment before asking for review.

## Updating the docs

- English only.
- Prefer runnable snippets. If the snippet would need imports to compile, pick an example from `examples/` or from the integration tests and copy it verbatim.
- Don't duplicate godoc — link to `pkg.go.dev` instead.

## Releasing

Tag `vX.Y.Z`, push the tag. Prior to 1.0.0 the API is not guaranteed to be stable — call out anything breaking in the release notes.
