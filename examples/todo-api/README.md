# todo-api — gerpo example

Minimal but production-shaped example: PostgreSQL, goose migrations, a CRUD JSON
API for a `tasks` table, request-scope cache wired as HTTP middleware, domain
error mapping, graceful shutdown. Total: ~350 lines of Go.

## Run with Docker Compose

Brings up Postgres + the API:

```bash
cd examples/todo-api
docker compose up --build
```

The API listens on `:8080`. Postgres is exposed on `:5432` with credentials
`todo` / `todo`.

Hit it:

```bash
# create
curl -s -X POST localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"write the docs","description":"section 5.3"}' | jq

# list (pagination + filter)
curl -s 'localhost:8080/tasks?page=1&size=10&done=false' | jq

# fetch one
curl -s localhost:8080/tasks/<id> | jq

# patch
curl -s -X PATCH localhost:8080/tasks/<id> \
  -H 'Content-Type: application/json' \
  -d '{"done":true}' | jq

# delete
curl -i -X DELETE localhost:8080/tasks/<id>
```

Tear down (keeps data volume):

```bash
docker compose down
```

Wipe the database too:

```bash
docker compose down -v
```

## Run locally (no Docker for the API)

Just Postgres from compose, the API from your shell:

```bash
cd examples/todo-api
docker compose up -d postgres
DATABASE_URL='postgres://todo:todo@localhost:5432/todo?sslmode=disable' \
    go run ./cmd/server
```

## Structure

```
cmd/server/
    main.go                  wiring: pool → migrations → repo → http
    migrations/
        0001_init_tasks.sql  goose single-file migration (up/down)
internal/task/
    model.go                 Task struct — no tags, no db markers
    repo.go                  gerpo.Repository[Task] + error transformer
    service.go               domain logic; Update goes through RunInTx
    http.go                  REST handlers on net/http (Go 1.22+ routing)
```

## What this example demonstrates

- **Column bindings by pointer** — see `internal/task/repo.go`. `ReadOnly()`,
  `ReturnedOnInsert()`, `ReturnedOnUpdate()` for the server-generated id,
  created_at and the trigger-maintained updated_at.
- **RETURNING out of the box** — `Insert` fills the zero-valued `ID` / `CreatedAt`
  from the DB's `DEFAULT gen_random_uuid()` / `DEFAULT NOW()` back onto the
  caller's struct.
- **Request-scope cache** — `cacheMiddleware` in `main.go` wraps every incoming
  request's ctx; two back-to-back `GetFirst` calls in the PATCH handler hit the
  cache, not the DB.
- **Transactions via context** — `service.Update` calls `gerpo.RunInTx`; every
  repository operation inside the closure picks up the same tx automatically
  without any extra parameter.
- **Domain error mapping** — `gerpo.ErrNotFound` is rewritten to `task.ErrNotFound`
  in the repository's `WithErrorTransformer`; HTTP handlers only know about the
  domain shape and reply with `404`.

## Schema

One table, one trigger:

```sql
CREATE TABLE tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    done        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);
```

`updated_at` is maintained by a `BEFORE UPDATE` trigger so the client never
writes to it — `OmitOnInsert().ReturnedOnUpdate()` in the column definition
keeps the write path clean and still reads the post-trigger value back.

## What this example deliberately leaves out

- Authentication, rate limiting, structured logs, TLS — add to taste.
- OpenTelemetry spans — see [`docs/production-setup.md`](../../docs/production-setup.md)
  for the `WithTracer` wiring; the example keeps the dependency tree small.
- Config via Viper/Koanf — two env variables are enough here.
