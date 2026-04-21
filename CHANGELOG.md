# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Commits follow [Conventional Commits](https://www.conventionalcommits.org/).

## [1.0.1] - 2026-04-21

### Documentation

- Add gerpolint + golangci-lint plugin page (38e6d99)

### Features

- **gerpolint:** Treat In(slice) same as In(slice...) (f653a82)
- **gerpolint:** Track append-chain accumulators in []any spread recovery (183d44c)
- **gerpolint:** Recover element types from []any spreads (5e0e256)
- Gerpolint golangci-lint v2 module plugin (5af05ff)
- Gerpolint static analyzer for type-safe WHERE filters (2535b9a)

### Misc

- **examples/todo-api:** Wire gerpolint as a golangci-lint plugin (6d872d6)

### Tests

- **bench:** Add isolated real-PG bench report (ac844de)
## [1.0.0] - 2026-04-20

### Bug Fixes

- **cache:** Wipe whole per-context storage on any repo write (eb49a83)

### Documentation

- Prepare changelog for v1.0.0 (db0efb4)
- Add examples/todo-api callouts across the site (d8c9ff7)
- Add production-ready setup page (08fd914)
- Document savepoints deferral (review 3.5) (f5d4e1d)
- State PostgreSQL-only scope prominently (fb360f2)
- Capture multi-database support backlog in TODO.md (53fd382)
- Drop LOC comparison, clarify SetValueFn probe scope (bc1da3a)

### Features

- Add InsertMany for batched multi-row inserts (dfbb249)
- RETURNING for Insert/Update with per-request override (b71ed1b)
- Hook callbacks return error, enable cascade-related-rows pattern (8ede5cd)
- **query:** Auto GROUP BY when an aggregate virtual column is in SELECT (c4e3700)
- **virtual:** New Compute/Aggregate/Filter API with FilterSpec sum-type (b849beb)

### Refactor

- Remove deprecated virtual-column API (ff28d90)
- Unify terminology — driver / adapter / backend (6433655)
- Replace repo.Tx(tx) with ctx-carried gerpo.WithTx / RunInTx (79a7f6a)
- Unify adapter/driver terminology (1ee19e1)
- **types:** Rename NEQ → NotEQ (and NEQFold → NotEQFold) (a3f3446)
- **types:** Rename IN/NIN, drop variadic ignoreCase, add *Fold ops (9499cf1)
- Rename gerpo.NewBuilder → gerpo.New (low-level New made internal) (27bf9d3)
- Remove deprecated callback JOIN API (LeftJoin/InnerJoin) (2cf7903)
- **cachectx:** Rename CtxCache→Cache, NewCtxCache→WrapContext (3fa7f8e)
- **column:** Rename *Protection to OmitOn* and add ReadOnly shortcut (9f25e4c)
- Drop AsColumn(), Field() returns a column directly (1fe1ab4)
- Move Tracer hook from executor to Repository with table info (6282ba0)
- **types:** Drop WhereOperation.OP escape hatch (8e96aab)
- **types:** Rename LIKE operators to readable names (b8d5406)

### Tests

- **integration:** Fix returning_test closure pointers (de5454f)
- Raise unit coverage from ~61% to ≥75% on every non-adapter package (e330ed1)

### Examples

- Replace kitchen-sink script with todo-api CRUD service (04f8af5)

### Examples/todo-api

- Remove stray .AsVirtual() from Title column (1a2a25e)
- Wrap plpgsql function in goose StatementBegin/End (5805e7c)
## [0.9.5] - 2026-04-19

### Bug Fixes

- Detect soft-delete value type mismatch at Build time (d95fd03)

### CI / Build

- Bump golangci-lint-action to v7 for golangci-lint v2 support (13e2430)
- Add golangci-lint v2 with baseline config and fix existing findings (762f315)
- Bump Go to 1.24 across go.mod, CI matrix and docs (02b5be1)
- Switch bench-diff and integration jobs to GitHub Actions (9fb2208)
- Post per-MR benchmark diff comment using benchstat (91ca0c5)
- **deps:** Bump golang.org/x/crypto from 0.31.0 to 0.45.0 (a4193e2)

### Documentation

- Prepare changelog for v0.9.5 (12dd77b)
- Add "Why gerpo?" comparison page (ced6c5f)
- Bootstrap CHANGELOG with git-cliff and add release tooling (2e62be1)
- Add runnable examples for godoc / pkg.go.dev (518f74f)
- Use simple/go icon instead of fontawesome/brands/go (2e1fcbd)
- Pin MkDocs deps in docs/requirements.txt (6b93b6e)
- Bootstrap MkDocs Material site with Features and Architecture (c280449)

### Features

- Add tracer hook for executor operations (c6879e4)
- Add LeftJoinOn/InnerJoinOn with bound parameters (8203804)

### Misc

- Add Makefile with common project commands (7eab67e)
- Fix "commited" typo in tx wrappers, add unit tests (c6d60d7)
- Drop unused private sql() helpers in sqlstmt (3b308a4)

### Performance

- Replace closures in query/linq builders with structured ops (9b6bf80)
- Reduce heap allocations on read hot path (5421d22)

### Refactor

- Factor query helpers around small composable interfaces (ec5f9f2)
- Extract shared placeholder-rewriting adapter base (0f36e36)

### Tests

- Pin JOIN/WHERE/Count argument ordering after LeftJoinOn (370d715)
- Add TestCompareDirectVsGerpo that prints a mock-bench summary table (6d1383c)
- Cover hooks, soft delete, virtual columns, transactions, cache, error transformer; fix pgx tx state bug (069c92a)
- Add query-layer integration tests and fix 3 bugs they uncovered (d98a89d)
- Add CRUD integration tests covering GetFirst/GetList/Count/Insert/Update/Delete (3fd13ee)
- Add integration test harness with docker-compose and per-adapter matrix (528c4be)
## [0.9.1] - 2025-12-06

### Misc

- Repository && executor: tx creates without return error (7ff9446)
## [0.9.0] - 2025-07-06

### Misc

- Builder && executor: add ability to set cache storage engine (60d2d98)
## [0.8.9] - 2025-05-23

### Bug Fixes

- Src: executor: adapters: pgx5: fix package name (542a10f)

### CI / Build

- Exclude adapters from code coverage tests result (202fdd8)
- **deps:** Bump golang.org/x/crypto from 0.22.0 to 0.35.0 (f860213)

### Documentation

- README.md: add performance metrics gerpo vs pure pgx v4 pool (57e768c)
- Update README.md, add ideology, add release road map, add documentation title, restructure features (1968852)
- Add executor adapters readme and executor readme with sequence scheme (d624c52)
- Add go 1.21 minimal version to readme (11eb428)
- Update README.md (833d33d)

### Misc

- Sqlstmt: sqlpart: where: add gte and lte support for time type (39fdfa8)
- Executor: adapters: add pgx v5 adapter (a38b8ba)
- Sqlstmt: use string builder for sql queries generation (e6de7df)
- Query: CT, NCT, BW, NBW. EW, NEW now case sensitive by default, insensitive option was added to methods (6ba85db)
- Sqlstmt: sqlpart: where: use concat instead `||` in sql filters (f85b191)
- Reorganize types package (bd27881)
- Exclude panics, return errors instead (0447f7b)
- Executor: add ability to set placeceholder for databasesql adapter (0c2f47c)
- Query: add Only method for columns select (2fa173a)
## [0.8.4] - 2025-03-25

### Misc

- Sqlstmt: sqlpart: where: NIN and IN fix with nil or empty slices (4d6828b)
## [0.8.3] - 2025-03-20

### Misc

- Downgrade crypto to support go 1.18 (4425ddf)
## [0.8.2] - 2025-03-20

### Misc

- Repostory: use executor.ErrNoRows as gerpo.ErrNotFound (d07e6cd)
- Copy slices go package to local repo for compatible with go 1.18 (8d4a750)
## [0.8.1] - 2025-03-20

### Misc

- Types: column: don't use slice package (1866ac4)
## [0.8.0] - 2025-03-20

### Bug Fixes

- Executor: add rows close on count (8d10655)
- Columns storage in query and sqlstmt packages (1e7099d)

### CI / Build

- **deps:** Bump golang.org/x/crypto from 0.20.0 to 0.31.0 (9c76a8f)

### Features

- Add INNER JOIN support (b3396d1)
- Add db adapters abstraction level for support pgxv4 and any other sql drivers/libs (6e04c45)

### Misc

- Add tests package with basic usage use cases of repository with mocksql driver, for functional regress testing (341d9fc)
- Builder: add With prefix for all builder options (a2c815e)
- Now update return count of updated rows (68082d9)
- Replace cache package inside executor package (c708520)
- Executor: tx: adds rollback unless commit function for use with defer (2927499)
- Add comments for a lot of public methods and functions (666a5e9)
- Executor: rename cache source option (a721963)
- Columns builder: change field column type choose (virtual/column) (218fe4b)
- Cache: remove cache bundle interface, use basic Source interface instead (efd1f79)
- Add a lot of tests (424c06b)
- Refactoring: a lot of refactoring sql and query packages, remove soft deletion (ed196b5)

### Go

- Downgrade go to 1.18 (40c5e61)

### Readme

- Fix AsColumn method name in configuration examples (5d06586)
## [0.1.9] - 2024-11-27

### Core

- Fix join position in select sql builder (2459c66)
## [0.1.8] - 2024-11-08

### CI / Build

- Add build and tests with coverate on PR (cc912a6)

### Misc

- Repository: add missed error wrapping when delete method calls with zero deleted elements (8595c51)
- Add transactions support (6e067ba)
- Exclude repository test in auto tests run (fb0da7e)
- Refactoring cache and sql packages, add new executor and logger pkgs (2e246b6)

### Readme

- Fix typos in badges (121e359)
## [0.1.7] - 2024-11-01

### Sql

- Allow to use slices in where IN and NIN filters (a2b1cf0)
## [0.1.6] - 2024-10-29

### Misc

- Cache: ctx: disable and enable cache key func renaming (de1fe3f)
- Api: removed: this sample example is not needed in public, i don't wont support this all time (1dc8407)

### Sql

- Added test (0cf48e4)
## [0.1.4] - 2024-10-23

### Misc

- Repo: add error transformer func for wrapping errors to needed bussines type (be1fdab)
- Api: remove empty filters file (8082448)
## [0.1.3] - 2024-10-23

### Misc

- Api: sorts: init available sorts at column link to dto (ca25362)
- Repo: add new after insert and after update hooks (7ab058f)
## [0.1.2] - 2024-10-22

### Api

- Noop: removed (35ae6cb)
## [0.1.1] - 2024-10-21

### Api

- Join core and applier to core and add real example usage (31ad55d)
## [0.1.0] - 2024-10-21

### Misc

- Remove tmp file (a83a6db)
- Api: add example query integration with filters and sorts (759d058)
- Cleanup options and builder (remove unused and not implemented) (7786197)
- Query: compact query helpers to bundle encapsulate query calls (a8fdd90)
- Column: allow group sql action by default (2ed14a3)
- Query: make get first helper interface without depends to count helper interface (c8ae97e)
- Add persistent query to repository configuration (13afbac)

### Cache

- Ctx: rename ctx key and add Ctx prefix in function names (7f6bf5b)

### Query

- Linq: join: fix joins with empty string join (8702c8b)
- Linq: extends api to using with external tools (ed58827)
- Linq: exclude remove old not used code (b6bc62c)

### Sql

- Select: fix columns exclude (deleteFunc) (c44e17a)
- Query: fix in and nin operators (7c8e490)
- Executor: determine placeholder inside ctor (411a4ae)

### Types

- Columns: use add without fmap.Field (a5dc4f6)
## [0.0.9] - 2024-10-15

### Query

- Use update sql builder when we update entity (f360d32)
## [0.0.6] - 2024-10-15

### Hack

- Src: sql: always use postgres placeholder (0a6d2d9)
## [0.0.5] - 2024-10-15

### Misc

- Sql: placeholder: workaround otelsql connector (4d0c0d8)
## [0.0.4] - 2024-10-15

### Sql

- Placeholder: determine place holder with otelsql wrapper (3339252)
## [0.0.3] - 2024-10-15

### Misc

- Sql: fix placeholder determination for lib/pq (d52714e)
## [0.0.2] - 2024-10-11

### Misc

- Add Repository interface (d1e52af)
- Add specific user helper for each repository method (a4ab5d4)
## [0.0.1] - 2024-10-10

### Misc

- Make repository global (72da700)

### Query

- Add uuid support (08fbe60)

