package dto

import (
	"github.com/jmoiron/sqlx/types"
	"github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/libs/sqlx/internal/model"
)

type ApplicationDTO struct {
	ID          string
	Tenant      string
	Name        string
	Description string
	Labels      types.NullJSONText // JSON
}

type APIDTO struct {
	ID        string
	AppID     string
	Targeturl string
}

type DocumentDTO struct {
	ID     string
	Appid  string `json:"app_id"`
	Title  string
	Format string
	Data   string
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
		Appid:  appID,
		Data:   in.Data,
		Title:  in.Title,
		ID:     in.ID,
		Format: in.Format,
	}
}

func (a *APIDTO) ToModel() model.API {
	return model.API{
		ID:        a.ID,
		TargetURL: a.Targeturl,
	}
}

func APIFromModel(appID string, in model.API) *APIDTO {
	return &APIDTO{
		ID:        in.ID,
		Targeturl: in.TargetURL,
		AppID:     appID,
	}
}

func (app *ApplicationDTO) ToModel() model.Application {
	return model.Application{
		ID:          app.ID,
		Tenant:      app.Tenant,
		Name:        app.Name,
		Labels:      app.Labels.String(),
		Description: app.Description,
	}
}

func ApplicationFromModel(in model.Application) ApplicationDTO {
	return ApplicationDTO{
		Tenant:      in.Tenant,
		Description: in.Description,
		ID:          in.ID,
		Name:        in.Name,
		Labels:      types.NullJSONText{Valid: true, JSONText: ([]byte)(in.Labels)},
	}
}
