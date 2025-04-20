# PGX v5 Pool Executor db Adapter
Executor db adapter implementation for [pgx/v5](https://github.com/jackc/pgx) pkg.

Works with `*github.com/jackc/pgx/v5/pgxpool.Pool`.

## Example
Postgres SQL
```go
package main

import (
    "database/sql"
    "github.com/insei/gerpo"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/insei/gerpo/executor/adapters/pgx5"
)

func main() {
	// already initialized pgxv5 pool
    var poolv5 *pgxpool.Pool
    dbWrap := pgx4.NewPoolAdapter(poolv5)
    
    repo, err := gerpo.NewBuilder[ModelType]().DB(dbWrap)
    // ... Configuring repository
}
```