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

