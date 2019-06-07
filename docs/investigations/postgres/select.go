package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

func main() {
	connStr := "user=postgres password=mysecretpassword dbname=compass sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	r, err := db.Query("SELECT app.id, app.tenant, app.name from applications app join apis api on app.id=api.app_id join events ev on app.id = ev.app_id join documents d on app.id=d.app_id LIMIT 100;")
	if err != nil {
		panic(err)
	}

	bef := time.Now()
	i:=0
	for i = 0; r.Next();i++ {
		var id int
		var tenant string
		var name string
		if err := r.Scan(&id, &tenant, &name); err != nil {
			panic(err)
		}

	}
	fmt.Println(time.Since(bef))
fmt.Println(i)


	db.Close()
}

type Application struct {
	ID     int
	Tenant string
	Name   string
}

type Api struct {
	ID    int
	Name  string
	AppID int
}

type Event struct {
	ID    int
	Name  string
	AppID int
}

type Document struct {
	ID    int
	Name  string
	AppID int
}
