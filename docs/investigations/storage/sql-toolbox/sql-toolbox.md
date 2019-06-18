# SQL Toolbox

In this document, you can find consideratons about which Golang library to use for DAO layer and 
how to handle testing and database migration. 

## Go Libraries Comparison

Three the most popular libraries were evaluated:
- sqlx + Squirrel
- beego
- gorm

### Beego
1. Beego does not work correctly with Postgres, [issue](https://github.com/astaxie/beego/issues/3070) 
2. Odd filtering support:
```
qs.Filter("name__icontains", "slene")
```
3. Very limited support for transactions:
```
err = d.ormer.Begin()
```
In contrast to other libraries, starting transaction does not create explicit Transaction object. If a developer forget to commit or rollback transaction, it can interfere with another transaction. 

4. According to their [documentation](https://beego.me/docs/mvc/model/overview.md):

> This framework is still under development so compatibility is not guaranteed.

### GORM
GORM has a lot of helper functions, such as:
- total count query
```
db.Where("name = ?", "jinzhu").Or("name = ?", "jinzhu 2").Find(&users).Count(&count)
```
- rollback transaction if needed:
```
tx.RollbackUnlessCommitted()
```
On the other hand, I found also some drawbacks:

1. GORM seems to be very error-prone, see those 2 quotes from their documentation:

> NOTE When query with struct, GORM will only query with those fields has non-zero value, that means if your field’s value is 0, '', false or other zero values:
```
	//  it won’t be used to build query conditions, for example:
	//db.Where(&User{Name: "jinzhu", Age: 0}).Find(&users)
	////// SELECT * FROM users WHERE name = "jinzhu";
```


>  WARNING When deleting a record, you need to ensure its primary field has value, and GORM will use the primary key to delete the record, if the primary key field is blank, GORM will delete all records for the model

2. GORM uses unusual error handling, IMO overlooking errors can be more frequent with that approach:

```
	if err := d.db.Limit(p.PageSize).Order("id").Find(&apps).Error; err != nil {
		return nil, err
	}
```

3. Some people [claim](https://www.reddit.com/r/golang/comments/8j3219/anyone_using_gorm_in_production_is_it_slow/) that GORM is not performant and too complex.  

### Sqlx + Squirrel
1. Helper functions very similar to standard library
2. Good [documentation](https://jmoiron.github.io/sqlx/)
3. For building SQL Queries,  [squirrel](github.com/Masterminds/squirrel) can be used.
```
	selBuilder := sq.Select("*").From("applications").OrderBy("id").Limit(uint64(p.PageSize))
	str, args, err := selBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	err = d.db.Select(&apps, str, args...)
```


4. Sqlx or Squirrel don't use Go tags, so insert commands have to be created manually:
```
	_, err = txx.NamedExecContext(ctx, "insert into applications(id,tenant,name,description,labels) values (:id, :tenant, :name, :description, :labels)", app)
```

No one library provides support for JSON queries. 

Sqlx and Squirrel is our first-choice library because of it's simplicity and explicity. 

## Testing
For mocking interactions with DB, we can use [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)
```
// a successful case
func TestShouldUpdateStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE products").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO product_viewers").WithArgs(2, 3).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	if err = recordStats(db, 2, 3); err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

```

For pre-populating DB with test data, we can use [testfixtures](https://github.com/go-testfixtures/testfixtures) or [polluter](https://github.com/romanyx/polluter). For code samples, see [this article](https://medium.com/@romanyx90/testing-database-interactions-using-go-d9512b6bb449)

## Schema updates
Schema updated can be performed using helm hooks `pre-upgrade` and `pre-install`. 
For performing migration, there are 2 interesting projects written in Go:
1. [Golang-migrate](https://github.com/golang-migrate/migrate) with 2301 stars on Github
- for every migration, 2 files are created: `up` and `down`
- supports only SQL 
- creates additional table: `schema_migrations`

2. [Goose](https://github.com/pressly/goose) with 887 stars on Github
- creates only one file per migration, to distinguish `up` and `down` SQL statements, comments are used
- supports SQL and Go binaries
- creates additional table: `goose_db_version`



Go-migrate seems to be more popular, is easier (use file names, instead of custom comments). It does not support Go binaries but at the moment I don't see advantages of them over plain SQL, so I suggest to use `golang-migrate`. 