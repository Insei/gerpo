# Ideology

The five rules gerpo commits to (from the README):

1. **If SQL exists in your project, it lives only in the repository configuration.**
2. **Every column is bound to a struct field through a pointer.** Not by name, not by tag.
3. **Entities carry no database markers.** No tags, no special interfaces.
4. **We do not implement relations between entities.** No `hasMany`, no `belongsTo` — that's the business layer's job.
5. **We do not modify the database schema.** No migrations, no `CREATE`/`ALTER` — gerpo only reads and writes data.

These rules aren't stylistic preferences — they are constraints that define the shape of the whole library. Any proposal that breaks one of them is almost certainly heading toward an ORM, and gerpo is deliberately not an ORM.

## Why "not an ORM"?

- **ORM models** try to hide SQL and table structure behind an object model. But SQL leaks: through N+1, through implicit migrations, through queries you can't express. Sooner or later, the team has to understand both layers anyway.
- **gerpo's repository model** is honest: yes, this is a SQL database, yes, there are tables. You work with them directly — just with the convenience of a type-safe configuration.

## The cost and the payoff of each rule

### 1. SQL only in the config

**Cost:** you have to describe columns and persistent conditions upfront; you cannot slip a JOIN into a handler "just for this call".

**Payoff:** exactly one place where the schema is read from. SQL changes are always visible in the PR.

### 2. Bindings via pointers

**Cost:** one struct field = one pointer in the config. No `c.Column("name")`.

**Payoff:** renames are a plain refactor, typos are caught by the compiler. You can't misspell a column name in a string.

### 3. No database markers on entities

**Cost:** a plain Go struct with no schema information — you keep it in a separate config.

**Payoff:** a domain entity stays a domain entity. It doesn't drag `json:"foo" db:"bar" validate:"…"` along with it — those concerns live where they should.

### 4. No relations

**Cost:** if there's a 1-N relationship between `User` and `Post`, you write the matching methods yourself (`FindPostsByUser`) — gerpo offers no magic navigation.

**Payoff:** no lazy-load tornadoes, a predictable number of queries.

### 5. No migrations

**Cost:** the database schema is managed by a separate tool (`golang-migrate`, `goose`, `atlas`, …).

**Payoff:** one layer, one responsibility. gerpo happily survives any schema-versioning scheme.

## Side-by-side summary

|  | GORM / ent | gerpo |
|---|---|---|
| SQL hidden | ✔ | ✘ (visible in the config) |
| Migrations | ✔ | ✘ |
| Relations | ✔ | ✘ |
| Struct tags | ✔ | ✘ |
| Field pointers | ✘ | ✔ |
| CRUD + WHERE DSL | ✔ | ✔ |
| Multiple drivers | ✔ | ✔ |
| Context-aware cache | partial | ✔ |

gerpo is intentionally **smaller**, and that is its value proposition.
