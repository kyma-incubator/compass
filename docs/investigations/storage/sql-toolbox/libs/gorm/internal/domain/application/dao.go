package application

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/gormtoolbox/libs/gorm/db"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/gormtoolbox/libs/gorm/dto"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/gormtoolbox/libs/gorm/model"
)

func NewApplicationDao(db *gorm.DB) *Dao {
	return &Dao{
		db: db,
	}
}

type Dao struct {
	db *gorm.DB
}

func (d *Dao) GetApplications(ctx context.Context, p model.PageRequest, sel model.Filer) (*model.ApplicationPage, error) {
	// TODO NOTE When query with struct, GORM will only query with those fields has non-zero value, that means if your field’s value is 0, '', false or other zero values,
	//  it won’t be used to build query conditions, for example:
	//db.Where(&User{Name: "jinzhu", Age: 0}).Find(&users)
	////// SELECT * FROM users WHERE name = "jinzhu";

	// TODO cool Count: !!! db.Where("name = ?", "jinzhu").Or("name = ?", "jinzhu 2").Find(&users).Count(&count)

	var apps []dto.ApplicationDTO
	// TODO error handling

	if err := d.db.Limit(p.PageSize).Order("id").Find(&apps).Error; err != nil {
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

	tx := d.db.Begin()
	if tx.Error != nil {
		return nil, err
	}

	defer func() {
		tx.RollbackUnlessCommitted() //TODO cool
	}()

	if err := tx.Create(&app).Error; err != nil {
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

			if err := tx.Create(&apiDTO).Error; err != nil {
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

			if err := tx.Create(&dDTO).Error; err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &app, nil
}

func (d *Dao) DeleteApplication(ctx context.Context, id string) (bool, error) {
	// TODO WARNING When deleting a record, you need to ensure its primary field has value, and GORM will use the primary key to delete the record, if the primary key field is blank, GORM will delete all records for the model

	// TODO we cannot use db.First because primary key is not an integer

	// TODO placeholder differs between DBs, so it is better to use named params, but I cannot find it in gorm

	app := dto.ApplicationDTO{ID: id}

	res := d.db.Delete(app)
	rows := res.RowsAffected
	err := res.Error
	if err != nil {
		return false, err
	}

	switch rows {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("incorrect number of removed applications: %d", rows)
	}

}
