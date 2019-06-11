package application

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sqlx/db"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sqlx/dto"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sqlx/model"
)

func NewApplicationDao(db *sqlx.DB) *Dao {
	return &Dao{
		db: db,
	}
}

type Dao struct {
	db *sqlx.DB
}

func (d *Dao) GetApplications(ctx context.Context, p model.PageRequest, sel model.Filer) (*model.ApplicationPage, error) {
	var apps []dto.ApplicationDTO
	err := d.db.SelectContext(ctx, &apps, "select * from applications")
	if err != nil {
		return nil, err
	}
	var out []model.Application
	for _, app := range apps {
		out = append(out, app.ToModel())
	}
	return &model.ApplicationPage{Data: out}, nil
}

func (d *Dao) CreateApplication(ctx context.Context, app model.Application) (*model.Application, error) {
	idp := db.IDProvider{}
	id, err := idp.GenID()
	if err != nil {
		return nil, err
	}
	app.ID = id

	//TODO handle storing dependant objects

	txx, err := d.db.Beginx()
	if err != nil {
		return nil, err
	}

	_, err = txx.NamedExecContext(ctx, "insert into applications(id,tenant,name,description,labels) values (:id, :tenant, :name, :description, :labels)", app)
	if err != nil {
		txx.Rollback()
		return nil, err
	}

	if app.Apis.Data != nil {
		for _, a := range app.Apis.Data {
			apiDTO := dto.APIFromModel(app.ID, a)
			id, err := idp.GenID()
			if err != nil {
				txx.Rollback()
				return nil, err
			}
			fmt.Println("TUTUAJ",id)
			apiDTO.ID = id
			apiDTO.Appid = app.ID
			_, err = txx.NamedExecContext(ctx, "insert into apis(id, app_id, target_url) values(:id,:appid, :targeturl)", apiDTO)//TODO
			if err != nil {
				txx.Rollback()
				return nil, err
			}
		}
	}

	if app.Documents.Data != nil {
		for _, d := range app.Documents.Data {
			dDTO := dto.DocumentFromModel(app.ID, d)
			id, err := idp.GenID()
			if err != nil {
				txx.Rollback()
				return nil, err
			}
			dDTO.ID = id
			_, err = txx.NamedExecContext(ctx, "insert into documents(id,appID, title, format,data) values(:id, :appid, :title, :format, :data)", dDTO)
			if err != nil {
				txx.Rollback()
				return nil, err
			}
		}
	}

	if err := txx.Commit(); err != nil {
		return nil, err
	}
	return nil, nil
}



func (d *Dao) DeleteApplication(ctx context.Context, id string) (bool, error) {
	res, err := d.db.NamedExecContext(ctx, "delete from applications where id=:id", map[string]string{"id": id})
	if err != nil {
		panic(err)
	}
	aff, err := res.RowsAffected()
	fmt.Println("Affected", aff)
	switch aff {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("incorrect number of removed applications: %d", aff)
	}

}
