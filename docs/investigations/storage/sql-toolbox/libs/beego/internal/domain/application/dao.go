package application

import (
	"context"
	"github.com/astaxie/beego/orm"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/beego/internal/db"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/beego/internal/dto"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/beego/internal/model"
	"github.com/pkg/errors"
)

func NewApplicationDao(ormer orm.Ormer) *Dao {
	return &Dao{
		ormer: ormer,
	}
}

type Dao struct {
	ormer orm.Ormer
}

func (d *Dao) GetApplications(ctx context.Context, p model.PageRequest, sel model.Filer) (*model.ApplicationPage, error) {
	var apps []dto.ApplicationDTO

	qs := d.ormer.QueryTable(apps)
	_, err := qs.Limit(p.PageSize).OrderBy("id").All(&apps)
	if err != nil {
		return nil, err
	}

	// TODO not nice at all: qs.Filter("name__contains", "slene")

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

	appDTO := dto.AppFromModel(app)

	// TODO this is soo bad, because if I forget to commit transaction it can be extended to other calls IMO
	// in other approaches, there is a separate object for transactions!!!
	err = d.ormer.Begin()
	if err != nil {
		return nil, err
	}
	if _, err := d.ormer.Insert(appDTO); err != nil {
		d.ormer.Rollback()
		return nil, err
	}
	//
	//if app.Apis.Data != nil {
	//	for _, a := range app.Apis.Data {
	//		apiDTO := dto.APIFromModel(app.ID, a)
	//		id, err := idp.GenID()
	//		if err != nil {
	//			d.ormer.Rollback()
	//			return nil, err
	//		}
	//		apiDTO.ID = id
	//		a.ID = id
	//		apiDTO.AppID = app.ID
	//
	//		if _, err := d.ormer.Insert(apiDTO); err != nil {
	//			d.ormer.Rollback()
	//			return nil, err
	//		}
	//	}
	//}
	//
	//if app.Documents.Data != nil {
	//	for _, doc := range app.Documents.Data {
	//		dDTO := dto.DocumentFromModel(app.ID, doc)
	//		id, err := idp.GenID()
	//		if err != nil {
	//			d.ormer.Rollback()
	//			return nil, err
	//		}
	//		dDTO.ID = id
	//		doc.ID = id
	//
	//		if _, err := d.ormer.Insert(dDTO); err != nil {
	//			d.ormer.Rollback()
	//			return nil, err
	//		}
	//	}
	//}

	if err := d.ormer.Commit(); err != nil {
		return nil, err
	}

	return &app, nil
}

func (d *Dao) DeleteApplication(ctx context.Context, id string) (bool, error) {
	rows, err := d.ormer.Delete(dto.ApplicationDTO{ID: id})
	if err != nil {
		return false, err
	}
	switch rows {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, errors.New("wrong number of applications removed")

	}
}
