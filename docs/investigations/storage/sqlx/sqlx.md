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
