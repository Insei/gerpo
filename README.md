# GERPO

[![codecov](https://codecov.io/gh/Insei/gerpo/graph/badge.svg?token=LGY9O9OJF5)](https://codecov.io/gh/Insei/gerpo)
[![build](https://github.com/Insei/gerpo/actions/workflows/go.yml/badge.svg)](https://github.com/Insei/gerpo/actions/workflows/go.yml)
[![Goreport](https://goreportcard.com/badge/github.com/insei/gerpo)](https://goreportcard.com/report/github.com/insei/gerpo)
[![GoDoc](https://godoc.org/github.com/insei/gerpo?status.svg)](https://godoc.org/github.com/insei/gerpo)

Welcome to the **GERPO** repository! This document provides a brief overview of the project, build and run instructions, and other helpful information.

## About GERPO
**GERPO** (Golang + Repository) is a generic repository implementation with advanced configuration capabilities and easy to use builders.

This project under active development.

Release 1.0.0 Road Map:
 * Repository builder changes:
   ~~* Add Caching engine configuration in repository builder.~~
   * Column builder changes:
     * New API for configuring virtual (calculated fields). Current Virtual fields configuration API marked as deprecated.
 * **All other API is stable and not planned to change in 1.0.0 release.**

Release 1.1.0 Road Map:
 * Add support to retrieve inserted ID, Time/Timestamps and some other returning values from db.
 * Add multiple insert support.

### Why GERPO?
1. Support any database drivers via simple [db adapters wrappers](https://github.com/Insei/gerpo/tree/main/executor/adapters#executor-db-adapters).
2. Fast repository level implementation, oriented on microservices.
3. Easily handle CRUD operations (Create, Read, Update, Delete) with powerful filtering and sorting.
4. Already implemented pagination for list queries.
5. Straightforward configuration user-friendly builders and there needed you can use SQL commands.
6. All SQL code in one place — inside the configuration.
7. Virtual (calculated, joined) columns with mapping to struct fields.
8. No dependent to other libraries. Only [fmap](https://github.com/Insei/fmap) was used for working with fields pointers.
9. [Caching support](https://github.com/Insei/gerpo/tree/main/executor/cache) (currently only context-oriented cache is supported, but it’s easy to implement other caching mechanisms).

### Ideology

The GERPO ideology consists of several rules:
1) If SQL code is used, then only in the repository configuration.
2) All columns are attached to entity fields via pointers to them in the repository settings.
3) There are no references in the entities that they are stored in the database (i.e. there are no tags)
4) We do not implement entity relationships.
5) We do not do migrations or other actions on database structure.

## Features
Essentially, **GERPO** is generic repository pattern implementation,
yes GERPO looks like ORM in some cases, but it's not an ORM.

- **Database adapters support**:
  - Any database can be used.
  - You can use tracing wrappers and any other wrappers.
- **Caching Engine**
  - Cache results in context for do not thinking about duplicated queries to database.
  - Cache in external store, like redis.
- **Repository configuration**
  - Map struct fields to SQL columns via pointers. 
    - Easy rename and refactor them.
    - Easy delete fields. Errors can be found at build time.
    - You always know where you can found columns mapping settings.
    - Protect fields to insert/update in database.
    - Define virtual(calculated)/joined fields.
  - Add callbacks and hooks.
    - Before insert.
    - Before update.
    - After select.
  - Define persistent filters, groupings, and joins.
  - Configure soft deletion.
    - Use special builder that replace delete function with update with needed fields update.
    - Configure Persistent filters for excluding soft deleted entities.
  - Configure Errors transformer to transform GERPO errors to you business Errors.

- **Per-query configuration**:
  - Query builder allows you to set up some query rules:
    - Execution fields selector allows you manage fields at update/insert operations:
        - Exclude certain fields by fields pointers.
        - Select only needed columns by fields pointers.
    - Where builder allows you:
      - Configure grouped filters.
      - Support OR/AND cases.
      - All of this via fields pointers.
    - Order builder:
      - Configure order via fields pointers.
    - Work with transactions.
    - Use already implemented pagination in your List queries.

## Performance
GERPO uses a minimal amount of reflection and is designed to have minimal allocations when used,
most allocations are initialized during configuration.
We work with unsafe pointers and offsets to determine the necessary fields in
the repository configuration and when querying the database.

I made 2 tests with absolutely identical conditions. Pure PGX V4 Pool vs GERPO over PGX V4 Pool.
Yes, we do 2x more allocations in a heap.
But I think our functionality is worth it.
In terms of time per operation, we are behind pure PGX v4 Pool by 8%.

Well, I didn't know what the results would be at the beginning of the design. But I have ideas for optimizations.
#### Pure PGX v4 pool:
```
BenchmarkGetOneFromDb-32           18049             65033 ns/op            1554 B/op         21 allocs/op
BenchmarkGetOneFromDb-32           18288             66617 ns/op            1555 B/op         21 allocs/op
BenchmarkGetOneFromDb-32           17860             66640 ns/op            1555 B/op         21 allocs/op
BenchmarkGetOneFromDb-32           18193             64665 ns/op            1555 B/op         21 allocs/op
BenchmarkGetOneFromDb-32           18198             65171 ns/op            1559 B/op         21 allocs/op
```
#### GERPO:
```
BenchmarkGetFirst/GetFirst-32              16808             69738 ns/op            2961 B/op         51 allocs/op
BenchmarkGetFirst/GetFirst-32              16614             70462 ns/op            2961 B/op         51 allocs/op
BenchmarkGetFirst/GetFirst-32              16905             70796 ns/op            2961 B/op         51 allocs/op
BenchmarkGetFirst/GetFirst-32              17059             70737 ns/op            2961 B/op         51 allocs/op
BenchmarkGetFirst/GetFirst-32              17184             69587 ns/op            2961 B/op         51 allocs/op
```

```
+------------------------------------------------------------+
|                  |              Get One/First              |
+------------------+---------------------+-------------------+
|                  |  PURE GPX Pool v4   | GERPO PGX Pool v4 |
|                  |                     | v4 Adapter        |
+------------------+---------------------+-------------------+
| B/op             | 0%                  | +100%             |
| allocs/op        | 0%                  | +120%             |
| ns/op            | 0%                  | +7.75%            |
+------------------+---------------------+-------------------+
```

## Installation
Go minimal version is `1.21`.
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
## Documentation:
* repository (main repository code - uses sqlstmt, query and executor for work, execute hooks, callbacks and error transformer)
* query (User API Query interface - works with sqlstmt via linq API)
  * linq (Internal Query API interface - works with sqlstmt)
* sqlstmt (Internal SQL queries generator interface - generates SQL queries and stores arguments)
* [executor](https://github.com/Insei/gerpo/tree/main/executor) (Internal SQL Queries executor API - execute SQL queries and map values to entities)