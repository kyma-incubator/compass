# Managed DBs Evaluation

Following document describes the evaluation of managed databases that can be used as a
persistence layer for Compass. 

## Requirements
To find out the best storage solution for Compass, we defined the following requirements: 
1. Fully managed solution - we don't want to spend time on managing DB. Priority: Super-high.
2. Extensible "schema" that allow searching by specifying JSON Path. Priority: Super-high.
We plan to allow clients storing metadata for Runtime or Application as a JSON object in labels or annotations. 
3. Easy local development. Priority: High.
4. Have an alternative that can be installed in k8s cluster. Priority: High.
5. Store big documents ~10MB. Priority: Super-high.
6. No vendor lock-in. At the moment we focus only on offerings provided by GCP, but the migration to other hyperscalers has to be easy. Priority: High.
7. Cross-region replication. Priority: Medium.
8. Cost-effective. Priority: Medium.
9. Support rich queries. Priority: Super high.

Beware that strong consistency is not on the requirements list, because eventual consistency is acceptable.

Below you can find a list of the evaluated solutions.
> **NOTE**: Question mark next to the requirements mean that it given requirement was not evaluated, because
we find other blockers to use given solution.

## Evaluated Solutions

### Cloud Spanner - GCP
Cloud Spanner has many blockers: no support for local development, vendor lock-in, no possibility to replace it with solution running inside a k8s cluster. In addition to that, it seems to be very expensive.

1. Yes
2. ? 
3. No
> We didn’t find a way to create a Cloud Spanner instance in a local environment. 
The closest we got was a docker image of CockroachDB, which is similar in principle, but very different in practice. 
For example, CockroachDB can use PostgreSQL JDBC. As it is imperative for a development environment to be as close a match as possible to production, Cloud Spanner is not ideal as one needs to rely on a full Spanner instance.
To save costs you can select a single region instance.

Source: https://www.lightspeedhq.com/blog/google-cloud-spanner-good-bad-ugly/ updated on 2018-03-21

4. No
5. ?
6. No
7. ?
8. Expensive: ~6480$ per month for a production environment
Regional: 0.9$ per node per hour
Multi-regional: 3$ per node per hour
> Minimum of 3 nodes recommended for production environments

Monthly regional cost: 0.9$ * 3 nodes * 720h = 1944$
Monthly multi-regional cost: 3$ * 3 nodes * 720h = 6480$

9. ?

### Cloud SQL (PostgreSQL) - GCP
Cloud SQL Postgres meets all our requirements apart from cross-region replication.

1. YES
2. Postgres has a JSON and JSONB column that is searchable. 
According to [this article](https://hackernoon.com/how-to-query-jsonb-beginner-sheet-cheat-4da3aa5082a3) querying on JSONB objects is almost as simple as classic SQL queries.
It seems to be possible to apply JSON schema validation as a PL SQL function: https://github.com/gavinwahl/postgres-json-schema

3. Postgres can be run locally.
4. Postgres chart can be used.
5.  According to the [this discussion](https://dba.stackexchange.com/questions/189876/size-limit-of-character-varying-postgresql)  text data type can store a string with 1GB size.
   In addition to that:
   > Different from other database systems, in PostgreSQL, there is no performance difference among three character types. In most situation, you should use text or varchar, and varchar(n) if you want PostgreSQL to check for the length limit.

6. [Amazon RDS](https://aws.amazon.com/rds/postgresql/) or [Azure DB for PostgreSQL](https://azure.microsoft.com/en-in/services/postgresql/).
7. For PostgreSQL, read replicas must be in the same region as the master instance, so cross-region replication is not fulfilled. 

8. Yes

3 instances `db-pg-f1-micro` with 10GB storage and 10GB backup = 30.5$ per month.

9. SQL

### Cloud SQL (MySQL) - GCP
Cloud SQL MyQL meets all our requirements.

1. YES
2. MySQL support JSON: https://dev.mysql.com/doc/refman/8.0/en/json.html. It looks that MySQL and Postgres don't have compatible API 
for querying JSON. For more details compare https://www.postgresql.org/docs/9.5/functions-json.html and https://dev.mysql.com/doc/refman/8.0/en/json-search-functions.html.
3. YES
4. YES
5. According to [this page](http://www.herongyang.com/JDBC/MySQL-CLOB-Columns-CREATE-TABLE.html), `LONGTEXT` can store 4GB of data.
6. [Amazon RDS](https://aws.amazon.com/rds/mysql/) or [Azure DB for MySQL](https://azure.microsoft.com/pl-pl/services/mysql/).
7. For MySQL, the same as for Postgres, read replicas must be in the same region as the master instance. In addition to that,
there is an option to manually [configure external replicas](https://cloud.google.com/sql/docs/mysql/replication/configure-external-replica).
Detailed instruction can be found [here](https://medium.com/searce/how-to-configure-mysql-replication-between-cloudsql-to-cloudsql-82362ce730f7). 
8. YES

3 instances `db-f1-micro` with 10GB storage nad 10 GB backup = 30.5$ per month.

9. YES

### Cloud Bigtable - GCP
Bigtable is a petabyte-scale, fully managed NoSQL database service for large analytical and operational workloads.
>  is not a good solution for storing less than 1 TB of data.

https://cloud.google.com/bigtable/docs/overview#storage-model

It seems that it is designed for different use-case than ours.


### Cloud Firestore - GCP
Firestore is GCP Specific Product - vendor lock-in. It seems that local development is also very, limited, for example
according to [this discussion](https://stackoverflow.com/questions/46563885/running-firestore-local-e-g-for-testing) you can run 
Firestore emulator to test security rules. 

### Firebase Realtime Database - GCP
This database is best suited for real-time notifications, synchronization apps state but has
very limited query capabilities. 

Source: https://www.codementor.io/cultofmetatron/when-you-should-and-shouldn-t-use-firebase-f62bo3gxv

### Cloud Memorystore - GCP
> Fully-managed in-memory data store service for Redis

Redis does not suite our requirements, because it will be difficult to store our model in key-value DB 
and support rich queries. 

## Summary
CloudSQL Postgres or MySQL suits the best our requirements about persisting data for Compass.
MySQL can be configured manually to replicate data across regions, but 
according to [PostgreSQL vs MySQL article](https://www.2ndquadrant.com/en/postgresql/postgresql-vs-mysql/), PostgreSQL is 
more SQL compliant, provides a better performance, implements more NoSQL features and because of that PostgreSQL 
is our first-choice DB.
