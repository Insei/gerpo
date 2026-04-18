# Architecture

This section is for people who want to know gerpo from the inside — contributors, reviewers, and anyone building custom adapters or features on top of it.

| Page | Topic |
|---|---|
| [Ideology](ideology.md) | The five rules that shape the library |
| [Layers](layers.md) | How a request travels from `Repository` down to the driver |
| [SQL generation](sql-generation.md) | `sqlstmt` and `sqlpart` — assembling the SQL text |
| [Field mapping](field-mapping.md) | How gerpo sees struct fields through pointers |
| [Caching internals](caching.md) | Inside `CtxCache` |
| [Adapter internals](adapters-internals.md) | How an adapter is written — placeholder rewrite, `Rows`, transactions |
| [Contributing](contributing.md) | Building, testing, and shipping a change |

## TL;DR

```
Repository[T]  ─►  query/*Helper[T]  ─►  query/linq (internal builders)
                                              │
                                              ▼
                                         sqlstmt (SQL codegen) ──► sqlstmt/sqlpart
                                              │
                                              ▼
                                         executor (run + cache + map) ──► executor/cache/*
                                              │
                                              ▼
                                         executor/adapters/{pgx5,pgx4,databasesql}
```

Each arrow is a request object moving down a layer. `Repository` doesn't know about the driver; the driver doesn't know about the model. In between there's pure SQL generation and value mapping.
