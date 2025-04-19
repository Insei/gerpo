# GERPO

[![codecov](https://codecov.io/gh/Insei/gerpo/graph/badge.svg?token=LGY9O9OJF5)](https://codecov.io/gh/Insei/gerpo)
[![build](https://github.com/Insei/gerpo/actions/workflows/go.yml/badge.svg)](https://github.com/Insei/gerpo/actions/workflows/go.yml)
[![Goreport](https://goreportcard.com/badge/github.com/insei/gerpo)](https://goreportcard.com/report/github.com/insei/gerpo)
[![GoDoc](https://godoc.org/github.com/insei/gerpo?status.svg)](https://godoc.org/github.com/insei/gerpo)

Welcome to the **GERPO** repository! This document provides a brief overview of the project, build and run instructions, and other helpful information.

## About GERPO
**GERPO** (Golang + Repository) is a generic repository implementation with advanced configuration capabilities and easy to use query builder.
This project under active development.

### Why GERPO?
1. Easily handle CRUD operations (Create, Read, Update, Delete) with powerful filtering and sorting.
2. Query builder (while not exactly LINQ, it’s conceptually close).
3. Straightforward configuration using SQL commands and user-friendly builders.
4. All SQL code in one place — inside the configuration.
5. Virtual (calculated, joined) columns with mapping to struct fields.
6. [Caching support](https://github.com/Insei/gerpo/tree/main/executor/cache) (currently only context-oriented cache is supported, but it’s easy to implement other caching mechanisms).

## Features
Essentially, **GERPO** is a helper for building SQL queries and mapping results to Go structs.
- **Data Sources (executor adapters)**:
    - pgx pool v4/v5
    - any database/sql driver
    - any other: you can add dbWrapper for any other database library, by implementing simple wrapper - executor.DBWrapper.
- **Repository configuration**:
    - Map struct fields to SQL columns via a LINQ-like builder.
    - Define virtual (calculated) fields (currently supports only bool; contributions welcome).
    - Protect certain fields from being updated or inserted.
    - Add callbacks and hooks.
    - Define persistent filters, groupings, and joins.
    - Configure soft deletion

- **Per-request configuration**:
    - Exclude certain fields by fields pointers.
    - Select only needed columns by fields pointers.
    - Work with transactions.
    - Configure filtering and sorting via fields pointers.
    - Use pagination in your GetList requests.

## Installation
Go minimal version is `1.21`.
```bash
go get github.com/insei/gerpo@latest
```

## Examples
Below you’ll find various configurations and usage examples.

### Repository Configuration

#### DB Adapter
Choose Database adapter that you need:

```go
package main

import (
  "database/sql"
  "time"
  "github.com/insei/gerpo"
  "github.com/insei/gerpo/executor/adapters/databasesql"
  "github.com/insei/gerpo/executor/adapters/pgx4"
  "github.com/insei/gerpo/executor/adapters/pgx5"
  "github.com/insei/gerpo/executor/adapters/placeholder"
  "github.com/jackc/pgx/v4/pgxpool"
)

func main() {
  // for database/sql postgres variant
  // "github.com/insei/gerpo/executor/adapters/databasesql"
  var db *sql.DB
  // for postgres change placeholder to dollar, by default placeholder is Question
  phOption := databasesql.WithPlaceholder(placeholder.Dollar)
  dbWrap := databasesql.NewAdapter(db, phOption)

  // for pgx4 pool
  // "github.com/insei/gerpo/executor/adapters/pgx4"
  var poolv4 *pgxpool.Pool
  dbWrap = pgx4.NewPoolAdapter(poolv4)

  // for pgx5 pool
  // "github.com/insei/gerpo/executor/adapters/pgx5"
  var poolv5 *pgxpool.Pool
  dbWrap = pgx5.NewPoolAdapter(poolv5)

  repo, err := gerpo.NewBuilder[ModelType]().DB(dbWrap)
  // ...
}
```

#### Columns
```go
package main

import (
    "time"
    "github.com/insei/gerpo"
)

type test struct {
    ID        int
    CreatedAt time.Time
    UpdatedAt *time.Time
    Name      string
    Age       int
}

func main() {
    repo, err := gerpo.NewBuilder[test]().
        DB(dbWrap).
        Table("tests").
        Columns(func(m *test, columns *gerpo.ColumnBuilder[test]) {
            columns.Field(&m.ID).AsColumn().WithUpdateProtection()
            columns.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
            columns.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
            columns.Field(&m.Name).AsColumn()
            columns.Field(&m.Age).AsColumn()
        }).
        Build()

    // Handle err and proceed with repo usage
}
```

#### Joins
```go
package main

import (
    "context"
    "time"
    "github.com/insei/gerpo"
    "github.com/insei/gerpo/query"
)

type test struct {
    ID        int
    CreatedAt time.Time
    UpdatedAt *time.Time
    Name      string
    Age       int
    Joined    string
}

func main() {
    repo, err := gerpo.NewBuilder[test]().
        DB(dbWrap).
        Table("tests").
        Columns(func(m *test, columns *gerpo.ColumnBuilder[test]) {
            columns.Field(&m.ID).AsColumn().WithUpdateProtection()
            columns.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
            columns.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
            columns.Field(&m.Name).AsColumn()
            columns.Field(&m.Age).AsColumn()
            columns.Field(&m.Joined).AsColumn().WithTable("joined_table")
        }).
        WithQuery(func(m *test, h query.PersistentHelper[test]) {
            h.LeftJoin(func(ctx context.Context) string {
                return "<SQL JOIN COMMAND>"
            })
        }).
        Build()

    // Handle err and proceed with repo usage
}
```

#### Soft Deletion
```go
package main

import (
    "context"
    "time"
    "github.com/insei/gerpo"
    "github.com/insei/gerpo/query"
)

type test struct {
    ID        int
    CreatedAt time.Time
    UpdatedAt *time.Time
    Name      string
    Age       int
	DeletedAt *time.Time
}

func main() {
    repo, err := gerpo.NewBuilder[test]().
        DB(dbWrap).
        Table("tests").
        Columns(func(m *test, columns *gerpo.ColumnBuilder[test]) {
            columns.Field(&m.ID).AsColumn().WithUpdateProtection()
            columns.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
            columns.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
            columns.Field(&m.Name).AsColumn()
            columns.Field(&m.Age).AsColumn()
            columns.Field(&m.DeletedAt).AsColumn().WithInsertProtection() // configure soft deletion field/column
        }).
        WithSoftDeletion(func(m *User, softDeletion *gerpo.SoftDeletionBuilder[User]) {
            //Configure set value for soft deletion fields/columns
            softDeletion.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
                deletedAt := time.Now().UTC()
                return &deletedAt
            })
        }).
        WithQuery(func(m *test, h query.PersistentHelper[test]) {
            // Permanently exclude deleted elements from all queries
            h.Where().Field(&m.DeletedAt).EQ(nil)
        }).
        Build()

    // Handle err and proceed with repo usage
}
```

### Per-request Configuration

#### Exclude
Exclude certain fields from commands like SELECT/UPDATE/INSERT (Update/GetFirst/Insert/GetList):
```go
package main

import (
    "context"
    "github.com/insei/gerpo"
    "github.com/insei/gerpo/query"
)

type test struct {
  ID        int
  CreatedAt time.Time
  UpdatedAt *time.Time
  Name      string
  Age       int
  Joined    string
}

func main() {
  var repo gerpo.Repository[test] // Already initialized
    list, err := repo.GetList(ctx, func(m *test, h query.GetListHelper[test]) {
        h.Page(1).Size(2) // Pagination
        h.Exclude(&m.UpdatedAt, &m.ID)
    })
    // Handle err and work with the list
}
```

#### Where
Available for Count/GetFirst/GetList/Delete/Update, supporting where grouping (AND/OR):
```go
package main

import (
    "context"
    "github.com/insei/gerpo"
    "github.com/insei/gerpo/query"
)

type test struct {
  ID        int
  CreatedAt time.Time
  UpdatedAt *time.Time
  Name      string
  Age       int
  Joined    string
}

func main() {
    var repo gerpo.Repository[test] // Already initialized
    list, err := repo.GetList(ctx, func(m *test, h query.GetListHelper[test]) {
        h.Where().Field(&m.ID).LT(7) // Items with ID < 7
    })
    // Handle err and use the list
}
```

#### Order
Available for GetFirst/GetList:
```go
package main

import (
    "context"
    "github.com/insei/gerpo"
    "github.com/insei/gerpo/query"
)

type test struct {
  ID        int
  CreatedAt time.Time
  UpdatedAt *time.Time
  Name      string
  Age       int
  Joined    string
}

func main() {
  var repo gerpo.Repository[test] // Already initialized
    item, err := repo.GetFirst(ctx, func(m *test, h query.GetFirstHelper[test]) {
        h.OrderBy().Field(&m.CreatedAt).DESC()
    })
    // Handle err and use 'item'
}
```

---

We hope this information helps you quickly get started with **GERPO** and integrate it into your own projects. If you have any questions or suggestions, feel free to open an issue or contribute to the repository.
