package main

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/sqlx/internal/domain/application"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/sqlx/internal/model"
	_ "github.com/lib/pq"
)

func main() {
	connStr := "user=postgres password=mysecretpassword dbname=compass sslmode=disable"

	db, err := sqlx.Connect("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	d := application.NewApplicationDao(db)
	app := model.Application{
		Name:        "my-app",
		Labels:      types.NullJSONText{Valid: true, JSONText: []byte("{\"group\":\"default\"}")},
		Description: "desc",
		Tenant:      "tenant",
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

	fmt.Println(brandNewApp)
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
