# Contributing

## Development environment

- Go 1.24+.
- Docker for integration tests.
- `mkdocs-material` if you want to preview the docs locally (`pip install mkdocs-material && mkdocs serve`).

## Check-in loop

```bash
# Lint (golangci-lint v2)
golangci-lint run ./...

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

## Commit style — Conventional Commits

The repo uses [Conventional Commits](https://www.conventionalcommits.org/).
Format:

```
<type>(<optional-scope>)?(!)?: <subject under ~70 chars>

<optional body explaining the why, in full sentences>
<empty line>
<optional BREAKING CHANGE: explanation> | <optional Closes #123>
```

The `commit-lint` workflow validates every commit on a PR — anything that
doesn't start with one of the allowed types is rejected.

Allowed types and what they mean:

| Type        | Use for                                              | CHANGELOG section |
|-------------|------------------------------------------------------|-------------------|
| `feat:`     | new public API or capability                         | Features          |
| `fix:`      | bug fix                                              | Bug Fixes         |
| `perf:`     | performance improvement without behavior change      | Performance       |
| `refactor:` | code change with no behavior change                  | Refactor          |
| `docs:`     | documentation only                                   | Documentation     |
| `test:`     | test-only change                                     | Tests             |
| `ci:`       | CI / pipelines                                       | CI / Build        |
| `build:`    | dependencies, build files                            | CI / Build        |
| `chore:`    | tooling, repo housekeeping, formatting               | Misc              |
| `src:`      | low-level repo-internal change without other prefix  | Misc              |
| `revert:`   | git revert                                           | Reverts           |
| `style:`    | whitespace / formatting only (skipped in CHANGELOG)  | —                 |

Add `!` after the type or include `BREAKING CHANGE:` in the body to mark a
breaking change — it'll bubble up under "BREAKING CHANGES" in the CHANGELOG.

Examples:

```
feat: add LeftJoinOn helper for parameter-bound JOINs
fix(executor): skip nil tracer
refactor!: rename Repository.Tx to WithTx
```

## Opening a PR

A PR to `main` runs five jobs:

- `lint` — `golangci-lint run ./...` with the config in `.golangci.yml`.
- `unit` — build, race detector, full `go test`.
- `integration` — `//go:build integration` against a PG service container.
- `bench-diff` — runs mock benchmarks on head and on base, posts a [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) summary as a PR comment.
- `commit-lint` — validates every commit message against Conventional Commits.

`bench-diff` is `allow_failure: true` — a perf regression shows up in the comment but doesn't block merging on its own.

## Updating the docs

- English only.
- Prefer runnable snippets. If the snippet would need imports to compile, pick an example from `examples/` or from the integration tests and copy it verbatim.
- Don't duplicate godoc — link to `pkg.go.dev` instead.

## Releasing

The release flow is semi-automated: you regenerate `CHANGELOG.md` and tag
locally, the `release` workflow builds the GitHub Release notes from the
same `cliff.toml` config so both stay in sync.

Prerequisites (one time):

```bash
# git-cliff is the markdown generator. Pick one:
brew install git-cliff
# or:
cargo install git-cliff
# or download a binary: https://github.com/orhun/git-cliff/releases
```

Per release:

```bash
./scripts/release.sh v0.2.0
```

The script

1. refuses to run unless you're on a clean `main` and the tag does not exist;
2. fast-forwards `main` from `origin`;
3. regenerates `CHANGELOG.md` with the new tag at the top;
4. shows the diff so you can eyeball or edit the file;
5. on confirmation, commits the file and creates the annotated tag.

It deliberately does **not** push. Review with `git show v0.2.0`, then:

```bash
git push --follow-tags origin main
```

That push triggers `.github/workflows/release.yml`, which runs
`git cliff --latest --strip header` against the same config and creates a
GitHub Release with that excerpt as the body. Tags carrying a suffix
(`v1.0.0-rc1`) are marked as pre-releases automatically.

Pre-1.0.0 the API is not guaranteed to be stable — call breaking changes out
explicitly with `!` in the type or `BREAKING CHANGE:` in the body so they
appear in their own section of the CHANGELOG.
