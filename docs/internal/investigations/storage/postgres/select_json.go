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
	r, err := db.Query("SELECT id, data->'age', data->'name' from custom where data ->> 'name' = 'John' ")
	if err != nil {
		panic(err)
	}

	bef := time.Now()
	i := 0
	for i = 0; r.Next(); i++ {
		var id int
		var age string
		var name string
		if err := r.Scan(&id, &age, &name); err != nil {
			panic(err)
		}
		fmt.Println(id, age, name)

	}
	fmt.Println(time.Since(bef))
	fmt.Println(i)

	db.Close()
}
