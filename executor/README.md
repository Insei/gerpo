# Executor
Executor is engaged in sending a request to the sql database obtained from sqlstmt. Passes this request to the db adapter and processes the execution result. Before sending, it contacts the caching engine and get the result from it.

## DB Adapters
Executor adapters is advanced layer of abstraction for interaction with sql database.

## Caching Engine
Allow store executions results to cache.

## SQL STMT
SQL query and arguments builder.

## Interactions scheme
### Read
```mermaid
sequenceDiagram
    Repository->>Executor: Read actions
    Executor->>SQL STMT: Get SQL query and arguments
    SQL STMT->>Executor: Return SQL query and arguments
    Executor->>Executor: Check result
    Executor-->>Repository: If error returns error
    Executor->>Cache Engine: Cached?
    Cache Engine->>Executor: Cached flag and value
    Executor->>Executor: Check cached flag
    Executor-->> Repository: Return cached value
    Executor->> DB Adapter: Execute SQL query with arguments
    DB Adapter->>Executor: Return result
    Executor->>Executor: Check result
    Executor-->>Repository: If error returns error
    Executor->>Executor: Map result to entities and values (count)
    Executor->>Cache Engine: Add (query, args) as key and entities and values to Cache engine
    Executor->>Repository: Return entities and values (count)
```
### Write
```mermaid
sequenceDiagram
    Repository->>Executor: Write actions
    Executor->>SQL STMT: Get SQL query and arguments
    SQL STMT->>Executor: Return SQL query and arguments
    Executor->>Executor: Check result
    Executor-->>Repository: If error returns error
    Executor->> DB Adapter: Execute SQL query with arguments
    DB Adapter->>Executor: Return result
    Executor->>Executor: Check result
    Executor-->>Repository: If error returns error
    Executor->>Executor: Map result to entities and values (count)
    Executor->>Cache Engine: Reset cache for entity type
    Executor->>Repository: Return entities and values (count)
```
