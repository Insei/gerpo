# PGX v4 Pool Executor db Adapter
Executor db adapter implementation for [pgx/v4](https://github.com/jackc/pgx/tree/v4) pkg.

Works with `*github.com/jackc/pgx/v4/pgxpool.Pool`.

## Example
Postgres SQL
```go
package main

import (
    "database/sql"
    "github.com/insei/gerpo"
    "github.com/jackc/pgx/v4/pgxpool"
    "github.com/insei/gerpo/executor/adapters/pgx4"
)

func main() {
	// already initialized pgxv4 pool
    var poolv4 *pgxpool.Pool
    dbWrap := pgx4.NewPoolAdapter(poolv4)
    
    repo, err := gerpo.NewBuilder[ModelType]().DB(dbWrap)
    // ... Configuring repository
}
```