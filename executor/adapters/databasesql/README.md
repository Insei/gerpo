# Database SQL Executor db Adapter
Executor db adapter implementation for default database/sql golang pkg.

Supported any database that works with `*database/sql.DB`.
## Options
We support different arguments placeholders for different databases:
* Dollar (`$1, $2`)
* Question (`?, ?`)
* Colon (`:1, :2`)
* AtP (`@p1, @p2`)

## Restrictions
Database should support `CONCAT` function.

## Example
Postgres SQL
```go
package main

import (
  "database/sql"
  "github.com/insei/gerpo"
  "github.com/insei/gerpo/executor/adapters/databasesql"
  "github.com/insei/gerpo/executor/adapters/placeholder"
)

func main() {
  // for database/sql postgres variant
  var db *sql.DB
  // for postgres change placeholder to dollar, by default placeholder is Question
  phOption := databasesql.WithPlaceholder(placeholder.Dollar)
  dbWrap := databasesql.NewAdapter(db, phOption)

  repo, err := gerpo.NewBuilder[ModelType]().DB(dbWrap)
  // ... Configuring repository
}
```