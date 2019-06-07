# Managed DB Comparison

To find out best storage solution for Compass, we defined following requirements: 
1. Fully managed solution - we don't want to spend time on managing DB
2. Extensible "schema" that allow searching by specifying JSON Path. 
We plan to allow clients storing  metadata for Runtime or Application as a JSON objects in labels or annotations. 
3. Easy local development
4. Have an alternative that can be installed in k8s cluster.
5. Store big documents ~10MB
6. No vendor lock-in.
At the moment we focus only on offerings provided by GCP, but need to be able to migrate
to other hyperscalers.
7. Cross-region replication
8. Cost-effective
9. Support rich queries

Below you can find list of the evaluated solutions.
> **NOTE**: Question mark next to the requirements mean that it given requirement was not evaluated, because
we find other blockers to use given solution.

## Cloud Spanner - GCP
Cloud Spanner has many blockers: no support for local development, vendor lock-in, no possibility to replace it with 
solution running inside a k8s cluster. In addition to that, it seems to be very expensive.

1. No operations - Yes
2. ? 
3. No
> We didn’t find a way to create a Cloud Spanner instance in a local environment. 
The closest we got was a docker image of CockroachDB, which is similar in principle, but very different in practice. 
For example, CockroachDB can use PostgreSQL JDBC. As it is imperative for a development environment to be as close a match as possible to production, Cloud Spanner is not ideal as one needs to rely on a full Spanner instance.
To save costs you can select a single region instance.

Source: https://www.lightspeedhq.com/blog/google-cloud-spanner-good-bad-ugly/ updated on: 2018-03-21

4. No
5. ?
6. No
7. ?
8. Expensive: ~6480$ per month for production environment
Regional: 0.9$ per node per hour
Multi-regional: 3$ per node per hour
> Minimum of 3 nodes recommended for production environments

Monthly regional cost: 0.9$ * 3 nodes * 720h = 1944$
Monthly multi-regional cost: 3$ * 3 nodes * 720h = 6480$

9. ?

## Cloud SQL (Postgres SQL) - GCP
Cloud SQL Postgres meets all our requirements.

1. No operations - YES
2. Postgres has JSON and JSONB column that is searchable. 
According to [this article](https://hackernoon.com/how-to-query-jsonb-beginner-sheet-cheat-4da3aa5082a3) querying on JSONB objects is almost as simple as classic SQL queries.
It seems to be possible to apply JSON schema validation as a PL SQL function: https://github.com/gavinwahl/postgres-json-schema

3. Postgres can be run locally.
4. Postgres chart can be used.
5.  According to the [this discussion](https://dba.stackexchange.com/questions/189876/size-limit-of-character-varying-postgresql)  text data type can store a string with 1GB size.
   In addition to that:
   > Different from other database systems, in PostgreSQL, there is no performance difference among three character types. In most situation, you should use text or varchar, and varchar(n) if you want PostgreSQL to check for the length limit.

6. Example Amazon RDS: https://aws.amazon.com/rds/postgresql/
7. Replication: Yes
> Cloud SQL provides the ability to replicate a master instance to one or more read replicas. A read replica is a copy of the master that reflects changes to the master instance in almost real time.

https://cloud.google.com/sql/docs/postgres/replication/

8. Yes

3 instances `db-pg-f1-micro` with 10GB Storage and 10GB backup = 70$ per month

9. SQL

## Cloud Bigtable - GCP
Bigtable is a petabyte-scale, fully managed NoSQL database service for large analytical and operational workloads.
>  is not a good solution for storing less than 1 TB of data.

https://cloud.google.com/bigtable/docs/overview#storage-model

It seems that it is designed for different use-case than our.


## Cloud Firestore - GCP
Firestore is GCP Specific Product - vendor lock-in. You can run it locally by using 
Firestore emulator but I do not see options to install it inside k8s cluster.

## Firebase Realtime Database - GCP
This database is best suited for real-time notifications, synchronization apps state but has
very limited query capabilities. 

Source: https://www.codementor.io/cultofmetatron/when-you-should-and-shouldn-t-use-firebase-f62bo3gxv

## Cloud Memorystore - GCP
> Fully-managed in-memory data store service for Redis

Redis does not suite our requirements, because it will be difficult to store our model in key-value DB 
and support rich queries. 
