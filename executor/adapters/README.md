# Executor (db) Adapters
Executor adapters is advanced layer of abstraction for interaction with sql database drivers.
## Supported adapters
* [pgx v4](https://github.com/Insei/gerpo/tree/main/executor/adapters/pgx4)
* [pgx v5](https://github.com/Insei/gerpo/tree/main/executor/adapters/pgx5)
* [database/sql](https://github.com/Insei/gerpo/tree/main/executor/adapters/databasesql)

## How to add new Executor Adapter
Simply implement `executor/types/DBAdapter` interface and send pull request! Contributions are welcome!

See examples in already implemented adapters.