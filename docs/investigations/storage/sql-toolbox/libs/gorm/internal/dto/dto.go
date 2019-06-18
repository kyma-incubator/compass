package dto

import (
	"github.com/jmoiron/sqlx/types" //TODO
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/gorm/internal/model"
)

type ApplicationDTO struct {
	ID          string
	Tenant      string
	Name        string
	Description string
	Labels      types.NullJSONText // JSON
}

func (app *ApplicationDTO) TableName() string {
	return "applications"
}



type APIDTO struct {
	ID        string
	AppID     string
	TargetURL string	`gorm:"column:target_url"`
}

func (a *APIDTO) TableName() string {
	return "apis"
}

type DocumentDTO struct {
	ID     string
	AppID  string `json:"app_id"`
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

func  DocumentFromModel(appID string, in model.Document) *DocumentDTO {
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

func  APIFromModel(appID string, in model.API) *APIDTO {
	return &APIDTO{
		ID:        in.ID,
		TargetURL: in.TargetURL,
		AppID:     appID,
	}
}

func (app *ApplicationDTO) ToModel() model.Application {
	return model.Application{
		ID:          app.ID,
		Tenant:      app.Tenant,
		Name:        app.Name,
		Labels:      app.Labels,
		Description: app.Description,
	}
}
