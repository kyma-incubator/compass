# SQLX Investigation

+ simple, very similar to standard library with some helpers
+ good [documentation](https://jmoiron.github.io/sqlx/) 
+ does not provide SQL Builder, but we can use [squirrel](github.com/Masterminds/squirrel) for that.
+ with small modification, implementing total count can be quite simple
- I have to enumarate all fields to be persisted:
```insert into applications(id,tenant,name,description,labels) values (:id, :tenant, :name, :description, :labels)```

# Beego
- I was not able to perform simple insert, it looks that Beego does not work correctly with Postgres: https://github.com/astaxie/beego/issues/3070
- awful filtering:
```
qs.Filter("name__icontains", "slene")
```
- bad support for transactions
```
err = d.ormer.Begin()
```
## Testing
https://medium.com/@romanyx90/testing-database-interactions-using-go-d9512b6bb449

testfixtures: 

- store data in fixtures filess that are used to populate db at the beginning of test

https://github.com/DATA-DOG/go-sqlmock

## Migration
https://flywaydb.org/download/
Community edition does not provide:
- dry run
- undo

https://stackoverflow.com/questions/33622214/what-package-to-use-for-database-migrations-in-go
FYI, goose is dead, but there is a maintained fork: https://github.com/pressly/goose (887 stars on github)


https://www.liquibase.org/quickstart.html <- seems to be ok, especially plain sql
https://stackoverflow.com/questionBs/21847482/does-liquibase-support-dry-run

> This is all depends on your DBMS. Not all DBMS support transactional DDL. In Oracle this would simple not be possible (because you cannot rollback a drop table, or alter table) . If your DBMS supports transactional DDL (e.g. Postgres), then everything will work without a special "dry run" mode because if an error occurs, Liquibase will rollback the unsuccessful changeset.

XXX
my approach:
start pods in one after another (so maxUnavailabe has to be adjusted) , every pod at the beginning execute migration code (so it has to be in Go).
First pod execute real migration, and next one will detect that there is nothing to do for them.

another ideas:

- init containers. Thanks to that it does not have to be Go code, we have separation of concerns etc, but idea about 
not repeating migration is the same as in previous approach

- use helm hooks: pre-upgrade, pre-install.



https://github.com/golang-migrate/migrate
- check if this store any info in db which migrations were performed, because others do so.

https://github.com/pressly/goose
sql + go
hybrid versioning which seems to be strange


----

even for hibarnete, they don't use ddl.auto: https://www.sitepoint.com/schema-migration-hibernate-flywaydb/ and write migration scrips 


