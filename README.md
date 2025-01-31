# GERPO

[![codecov](https://codecov.io/gh/Insei/gerpo/graph/badge.svg?token=LGY9O9OJF5)](https://codecov.io/gh/Insei/gerpo)
[![build](https://github.com/Insei/gerpo/actions/workflows/go.yml/badge.svg)](https://github.com/Insei/gerpo/actions/workflows/go.yml)
[![Goreport](https://goreportcard.com/badge/github.com/insei/gerpo)](https://goreportcard.com/report/github.com/insei/gerpo)
[![GoDoc](https://godoc.org/github.com/insei/gerpo?status.svg)](https://godoc.org/github.com/insei/gerpo)

Welcome to the **GERPO** repository! This document provides a brief overview of the project, build and run instructions, and other helpful information.

## About GERPO
**GERPO** (Golang + Repository) is a generic repository implementation with advanced configuration capabilities and LINQ-like (Language Integrated Query) support.

### Why GERPO?
1. Easily handle CRUD operations (Create, Read, Update, Delete) with powerful filtering and sorting.
2. LINQ-like capabilities (while not exactly LINQ, it’s conceptually close).
3. Straightforward configuration using SQL commands and user-friendly builders.
4. All SQL code in one place — inside the configuration.
5. Virtual (calculated, joined) columns with mapping to struct fields.
6. [Caching support](https://github.com/insei/gerpo/cache/README.md) (currently only context-oriented cache is supported, but it’s easy to implement other caching mechanisms).

## Features
Essentially, **GERPO** is a helper for building SQL queries and mapping results to Go structs.

- **Repository configuration**:
    - Map struct fields to SQL columns via a LINQ-like builder.
    - Define virtual (calculated) fields (currently supports only bool; contributions welcome).
    - Protect certain fields from being updated or inserted.
    - Add callbacks and hooks.
    - Define persistent filters, groupings, and joins.

- **Per-request configuration**:
    - Exclude certain columns (SELECT/INSERT/UPDATE) using a builder.
    - Work with transactions.
    - Configure filtering and sorting via a LINQ-like builder.
    - Implement pagination in your GetList requests.

## Installation

```bash
go get github.com/insei/gerpo@latest
```

## Examples
Below you’ll find various configurations and usage examples.

### Repository Configuration

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
        DB(db).
        Table("tests").
        Columns(func(m *test, columns *gerpo.ColumnBuilder[test]) {
            columns.Column(&m.ID).WithUpdateProtection()
            columns.Column(&m.CreatedAt).WithUpdateProtection()
            columns.Column(&m.UpdatedAt).WithInsertProtection()
            columns.Column(&m.Name)
            columns.Column(&m.Age)
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
        DB(db).
        Table("tests").
        Columns(func(m *test, columns *gerpo.ColumnBuilder[test]) {
            columns.Column(&m.ID).WithUpdateProtection()
            columns.Column(&m.CreatedAt).WithUpdateProtection()
            columns.Column(&m.UpdatedAt).WithInsertProtection()
            columns.Column(&m.Name)
            columns.Column(&m.Age)
            columns.Column(&m.Joined).WithTable("joined_table")
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
    list, err := repo.GetList(ctxCache, func(m *test, h query.GetListHelper[test]) {
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
    list, err := repo.GetList(ctxCache, func(m *test, h query.GetListHelper[test]) {
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
    item, err := repo.GetFirst(ctxCache, func(m *test, h query.GetFirstHelper[test]) {
        h.OrderBy().Field(&m.CreatedAt).DESC()
    })
    // Handle err and use 'item'
}
```

---

We hope this information helps you quickly get started with **GERPO** and integrate it into your own projects. If you have any questions or suggestions, feel free to open an issue or contribute to the repository.