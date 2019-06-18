package dto

import (
	"database/sql"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/beegooolbox/libs/beego/model"
)

type ApplicationDTO struct {
	ID          string `orm:"pk;column(id)"` // TODO really???
	Tenant      string
	Name        string
	Description string
	Labels      sql.NullString // JSON
}

func (app *ApplicationDTO) TableName() string {
	return "applications"
}

type APIDTO struct {
	ID        string `orm:"pk"`
	AppID     string
	TargetURL string
}

func (a *APIDTO) TableName() string {
	return "apis"
}

type DocumentDTO struct {
	ID     string `orm:"pk"`
	AppID  string
	Title  string
	Format string
	Data   string
}

func (d *DocumentDTO) TableName() string {
	return "documents"
}

func (d *DocumentDTO) ToModel() model.Document {
	return model.Document{
		ID:     d.ID,
		Data:   d.Data,
		Format: d.Format,
		Title:  d.Title,
	}
}

func DocumentFromModel(appID string, in model.Document) *DocumentDTO {
	return &DocumentDTO{
		AppID:  appID,
		Data:   in.Data,
		Title:  in.Title,
		ID:     in.ID,
		Format: in.Format,
	}
}

func (a *APIDTO) ToModel() model.API {
	return model.API{
		ID:        a.ID,
		TargetURL: a.TargetURL,
	}
}

func APIFromModel(appID string, in model.API) *APIDTO {
	return &APIDTO{
		ID:        in.ID,
		TargetURL: in.TargetURL,
		AppID:     appID,
	}
}

func AppFromModel(in model.Application) *ApplicationDTO {
	var l sql.NullString
	if in.Labels == "" {
		l.Valid = false
	} else {
		l.Valid = true
		l.String = in.Labels
	}

	return &ApplicationDTO{
		ID:          in.ID,
		Labels:      l,
		Name:        in.Name,
		Description: in.Description,
		Tenant:      in.Tenant,
	}
}

func (app *ApplicationDTO) ToModel() model.Application {
	return model.Application{
		ID:          app.ID,
		Tenant:      app.Tenant,
		Name:        app.Name,
		Labels:      app.Labels.String,
		Description: app.Description,
	}
}
