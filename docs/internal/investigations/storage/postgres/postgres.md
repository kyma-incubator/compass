# PostgreSQL Evaluation


Our GraphQL API allows clients to perform very sophisticated queries, like get all applications with API, event definitions and all documentation.
In addition to that, we plan to extend our schema with schemaless data provided by clients. 
In this document, we evaluate if Postgres meets our expectations. 

## Usage
To run Postgres with populated data, use a `run_postgres.sh` script. This script performs the following tasks:

- generate data by using `gen.go` as CSV files and stores them in `sql` directory. In `gen.go` you can decide how
many entities to create. 
- run Postgres as a Docker image with mounted `sql` directory. This directory is mounted in `docker-entrypoint-initdb.d`
 and thanks to that all `SQL` files will be executed automatically on the DB startup. 
 The directory contains data in CSV files, schema definition and commands
to populate DB.

## Queries Performance
1. The query for all details of the given application.
To run a query, `select.go` was used.
Every application has 10 APIs, 10 events definition and 10 documents.

| Apps No.  | With indexes | Without indexes  |
|---------- |--------------|------------------|
| 1000      | 1-2 ms       | 2.5ms            |
| 10 000    | 1.4 ms       | 7.5 ms           |
| 100 000   | 1.4 ms       | 8-10 ms          |

2. The query for all details of all applications.
To run a query, use `select.go` and remove from query `where` clause.

| Apps No.  | With indexes | Without indexes  |
|---------- |--------------|------------------|
| 1000      | 1.5 s        | 1.5 s            |


As you can see, even for a small number of applications, this query is extremely slow.
Below you can find, that when querying for all applications, indexes are not used at all.
According to [a_horse_with_no_name](https://stackoverflow.com/users/330315/a-horse-with-no-name) in [this discussion](https://stackoverflow.com/questions/5203755/why-does-postgresql-perform-sequential-scan-on-indexed-column):

> If the SELECT returns more than approximately 5-10% of all rows in the table, a sequential scan is much faster than an index scan.

```
explain SELECT app.id, app.tenant, app.name from applications app join apis api on app.id=api.app_id join events ev on app.id = ev.app_id join documents d on app.id=d.app_id where  app.id=1;
                                                 QUERY PLAN
-------------------------------------------------------------------------------------------------------------
 Nested Loop  (cost=1.13..47.57 rows=1000 width=21)
   ->  Nested Loop  (cost=0.85..26.59 rows=100 width=21)
         ->  Nested Loop  (cost=0.56..16.85 rows=10 width=21)
               ->  Index Scan using applications_pkey on applications app  (cost=0.28..8.29 rows=1 width=21)
                     Index Cond: (id = 1)
               ->  Index Only Scan using apps_apis on apis api  (cost=0.29..8.46 rows=10 width=4)
                     Index Cond: (app_id = 1)
         ->  Materialize  (cost=0.29..8.51 rows=10 width=4)
               ->  Index Only Scan using apps_events on events ev  (cost=0.29..8.46 rows=10 width=4)
                     Index Cond: (app_id = 1)
   ->  Materialize  (cost=0.29..8.51 rows=10 width=4)
         ->  Index Only Scan using apps_documents on documents d  (cost=0.29..8.46 rows=10 width=4)
               Index Cond: (app_id = 1)
```


```
compass=# explain SELECT app.id, app.tenant, app.name from applications app join apis api on app.id=api.app_id join events ev on app.id = ev.app_id join documents d on app.id=d.app_id;
                                         QUERY PLAN
---------------------------------------------------------------------------------------------
 Hash Join  (cost=597.50..13436.86 rows=1000000 width=21)
   Hash Cond: (app.id = d.app_id)
   ->  Hash Join  (cost=309.50..1648.86 rows=100000 width=29)
         Hash Cond: (app.id = ev.app_id)
         ->  Hash Join  (cost=29.50..218.86 rows=10000 width=25)
               Hash Cond: (api.app_id = app.id)
               ->  Seq Scan on apis api  (cost=0.00..163.00 rows=10000 width=4)
               ->  Hash  (cost=17.00..17.00 rows=1000 width=21)
                     ->  Seq Scan on applications app  (cost=0.00..17.00 rows=1000 width=21)
         ->  Hash  (cost=155.00..155.00 rows=10000 width=4)
               ->  Seq Scan on events ev  (cost=0.00..155.00 rows=10000 width=4)
   ->  Hash  (cost=163.00..163.00 rows=10000 width=4)
         ->  Seq Scan on documents d  (cost=0.00..163.00 rows=10000 width=4)
```

3. The query for all details of all applications with page size = 100.
To run a query, use `select.go` and remove from query `where` clause and add `LIMIT 100`.

| Apps No.  | With indexes |
|---------- |--------------|
| 1000      | 200 us       |
| 10 000    | 200 us       |                   
| 100 000   | 200 us       |                         

As you can see, in this case, `limit` protects us from long-running queries. 

## PostgreSQL JSON
Postgres support querying on JSON fields. Run `select_json.go` to see it in action.
A script `run_postgres.sh` defined a table `custom` with 2 JSON columns and populate it with data.
Thanks to that capability, we can let clients store their data in our DB as JSON fields and later 
enable searching on those objects.