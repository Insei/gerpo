# Contributing

## Development environment

- Go 1.24+.
- Docker for integration tests.
- `mkdocs-material` if you want to preview the docs locally (`pip install mkdocs-material && mkdocs serve`).

## Check-in loop

Common tasks live in a `Makefile` — run `make help` for the catalog.

```bash
make lint               # golangci-lint v2
make test               # go test -race ./...
make integration-full   # docker up → integration tests → docker down
make bench              # Direct vs Gerpo mock benchmarks (5 runs)
make bench-report       # formatted summary table (~20s)
```

If you want finer control over the integration suite:

```bash
make integration-up     # start Postgres once
make integration        # run /tests/integration/ against the running PG
make integration-down   # stop Postgres
```

Override the DSN if your local PG differs:

```bash
make integration INTEGRATION_DSN="postgres://..."
```

To preview the MkDocs site:

```bash
make docs-serve   # http://127.0.0.1:8000
make docs-build   # build with --strict
```

You can of course still call the underlying `go test` / `docker compose` /
`golangci-lint` commands directly; the Makefile is a convenience layer, not
a requirement.

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
make release TAG=v0.2.0   # wraps scripts/release.sh
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
