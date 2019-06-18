package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/beegooolbox/libs/beego/domain/application"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/beegooolbox/libs/beego/dto"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/beegooolbox/libs/beego/model"
	_ "github.com/lib/pq"
)

func main() {
	connStr := "user=postgres password=mysecretpassword dbname=compass sslmode=disable"

	orm.RegisterModel(new(dto.ApplicationDTO), new(dto.APIDTO), new(dto.DocumentDTO))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	if db.Ping() != nil {
		panic("ping failed")
	}
	o, err := orm.NewOrmWithDB("postgres", "default", db)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	d := application.NewApplicationDao(o)
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
	// TODO It does not work: https://github.com/astaxie/beego/issues/3070
	// TODO panic: no LastInsertId available
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
