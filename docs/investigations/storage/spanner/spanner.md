# GCP Spanner

## Costs
Regional: 0.9$ per node per hour
Multi-regional: 3$ per node per hour

https://cloud.google.com/spanner/docs/quickstart-console

> Minimum of 3 nodes recommended for production environments

Monthly regional cost: 0.9$ * 3 nodes * 720h = 1944$
Monthly multi-regional cost: 3$ * 3 nodes * 720h = 6480$

## Local development

https://www.lightspeedhq.com/blog/google-cloud-spanner-good-bad-ugly/

> “At present, the drivers do not support DML or DDL statements.”
  — Spanner documentation
  
> We didn’t find a way to create a Cloud Spanner instance in a local environment. 
The closest we got was a docker image of CockroachDB, which is similar in principle, but very different in practice. 
For example, CockroachDB can use PostgreSQL JDBC. As it is imperative for a development environment to be as close a match as possible to production, Cloud Spanner is not ideal as one needs to rely on a full Spanner instance.
To save costs you can select a single region instance.

> Build and deploy for the cloud faster because Cloud SQL offers standard MySQL, PostgreSQL, and SQL Server databases. Use standard connection drivers and built-in migration tools to get started quickly.

## Cost

from 600 MB to 416 GB RAM

$0.0150–$8.0480 per hour

0.0150$ * 720h = 10,8 $


## Problems

Preview: Cloud SQL for SQL Server (alpha). Request access.

:-1: