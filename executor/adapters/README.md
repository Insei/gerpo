# Executor (db) Adapters
Executor adapters is advanced layer of abstraction for interaction with sql database.
## Supported adapters
* pgx v4 (pkg pgx4)
* pgx v5 (pkg pgx5)
* database/sql (pkg databasesql)

## How to add new Executor Adapter
Simply implement `executor/types/DBAdapter` interface and send pull request! Contributions are welcome!

See examples in already implemented adapters.