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
	r, err := db.Query("SELECT id, f1->'age', f1->'name', f2 from custom where f1 ->> 'name' = 'Adam' ")
	if err != nil {
		panic(err)
	}

	bef := time.Now()
	i:=0
	for i = 0; r.Next();i++ {
		var id int
		var age string
		var name string
		var f2 string
		if err := r.Scan(&id, &age, &name, &f2); err != nil {
			panic(err)
		}
		fmt.Println(id,age,name,f2)

	}
	fmt.Println(time.Since(bef))
	fmt.Println(i)


	db.Close()
}

