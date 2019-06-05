# Storage Summary

Requirements: 
1. No operations - Fully managed
2. Search by JSON Path  
3. Support local Development
4. Possible hosting in the k8s cluster
5. Store big documents 10GB
6. No vendor lock-in
7. Cross-region replication
8. Cost-effective
9. Rich queries

## Cloud Spanner - GCP
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

## Cloud SQL - GCP

## Cloud Bigtable - GCP

## Cloud Firestore - GCP

## Firebase Realtime Database - GCP
This database is best suited for real-time notifications, synchronization application state but has
very limited query capabilities. 

Source: https://www.codementor.io/cultofmetatron/when-you-should-and-shouldn-t-use-firebase-f62bo3gxv

## Cloud Memorystore - GCP
> Fully-managed in-memory data store service for Redis

Redis does not suite our requirements, because it will be difficult to store our model in key-value DB.
