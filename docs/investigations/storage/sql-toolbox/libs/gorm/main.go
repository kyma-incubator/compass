package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/gorm/internal/domain/application"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/gorm/internal/model"
	_ "github.com/lib/pq"
)

func main() {
	connStr := "user=postgres password=mysecretpassword dbname=compass sslmode=disable"


	db, err := gorm.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()
	// migrate the schema:  run auto migration for given models, will only add missing fields, won't delete/change current data
	db.AutoMigrate()

	d := application.NewApplicationDao(db)
	app := model.Application{
		Name:   "my-app",
		Labels: "{}",
	}
	app.Documents = model.DocumentPage{
		Data: []model.Document{
			{
				Title: "abcd",
			},
			{Title: "xyz"},
		},
	}
	app.Apis = model.APIPage{
		Data: []model.API{
			{
				TargetURL: "googgle.com",
			},
			{
				TargetURL: "cncf.io",
			},
		},
	}
	ctx := context.TODO()
	brandNewApp, err := d.CreateApplication(ctx, app)
	if err != nil {
		panic(err)
	}

	appPage, err := d.GetApplications(ctx, model.PageRequest{
		PageSize: 50,
	}, model.Filer{})

	if err != nil {
		panic(err)
	}

	fmt.Println(len(appPage.Data))
	nextPage, err := d.GetApplications(ctx, model.PageRequest{PageSize: 50, AfterCursor: appPage.PageInfo.EndCursor}, model.Filer{})
	if err != nil {
		panic(err)
	}
	fmt.Println(len(nextPage.Data))

	ex, err := d.DeleteApplication(ctx, brandNewApp.ID)
	if err != nil {
		panic(err)

	}
	fmt.Println("Exist", ex)
}
