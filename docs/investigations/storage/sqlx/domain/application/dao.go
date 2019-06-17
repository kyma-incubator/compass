package application

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sqlx/db"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sqlx/dto"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sqlx/model"
	"github.com/lann/builder"
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

	// TODO squirrel usage example
	selBuilder := sq.Select("*").From("applications").OrderBy("id").Limit(uint64(p.PageSize))
	//totalCnt := d.getTotalCountQuery(selBuilder)

	str, args, err := selBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	err = d.db.Select(&apps, str, args...)
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

	txx, err := d.db.Beginx()
	if err != nil {
		return nil, err
	}

	txStatus := &txStatus{}
	defer d.rollbackIfNeeded(txStatus, txx) // TODO I have to implement it on my own, to avoid rollbacking on every error

	// TODO: I have to specify all fields explicitly here to be persisted
	_, err = txx.NamedExecContext(ctx, "insert into applications(id,tenant,name,description,labels) values (:id, :tenant, :name, :description, :labels)", app)
	if err != nil {
		return nil, err
	}

	if app.Apis.Data != nil {
		for _, a := range app.Apis.Data {
			apiDTO := dto.APIFromModel(app.ID, a)
			id, err := idp.GenID()
			if err != nil {
				return nil, err
			}
			apiDTO.ID = id
			a.ID = id
			apiDTO.AppID = app.ID
			_, err = txx.NamedExecContext(ctx, "insert into apis(id, app_id, target_url) values(:id,:appid, :targeturl)", apiDTO) //TODO
			if err != nil {
				return nil, err
			}
		}
	}

	if app.Documents.Data != nil {
		for _, d := range app.Documents.Data {
			dDTO := dto.DocumentFromModel(app.ID, d)
			id, err := idp.GenID()
			if err != nil {
				return nil, err
			}
			dDTO.ID = id
			d.ID = id
			_, err = txx.NamedExecContext(ctx, "insert into documents(id,app_id, title, format,data) values(:id, :appid, :title, :format, :data)", dDTO)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := txx.Commit(); err != nil {
		return nil, err
	}

	txStatus.Committed = true

	return &app, nil
}

func (d *Dao) DeleteApplication(ctx context.Context, id string) (bool, error) {
	res, err := d.db.NamedExecContext(ctx, "delete from applications where id= :id", map[string]interface{}{"id": id})
	if err != nil {
		panic(err)
	}
	aff, err := res.RowsAffected()
	switch aff {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("incorrect number of removed applications: %d", aff)
	}

}

// TODO workaround how to generate query that returns total count
func (d *Dao) getTotalCountQuery(originalQuery sq.SelectBuilder) sq.SelectBuilder {
	countQuery := originalQuery.RemoveOffset().RemoveLimit()
	return builder.Set(countQuery, "Columns", []sq.Sqlizer{stringSqlizer("count(*)")}).(sq.SelectBuilder)
}

type stringSqlizer string

func (s stringSqlizer) ToSql() (string, []interface{}, error) {
	return string(s), nil, nil
}

func (d *Dao) rollbackIfNeeded(status *txStatus, tx *sqlx.Tx) {
	if !status.Committed {
		tx.Rollback()
	}
}

type txStatus struct {
	Committed bool
}
