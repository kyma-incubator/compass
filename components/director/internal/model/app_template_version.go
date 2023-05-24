package model

import (
	"time"
)

// ApplicationTemplateVersion missing godoc
type ApplicationTemplateVersion struct {
	ID                    string
	Version               string
	Title                 *string
	ReleaseDate           *time.Time
	CorrelationIDs        []string
	CreatedAt             time.Time
	ApplicationTemplateID string
}

type ApplicationTemplateVersionInput struct {
	Version        string
	Title          *string
	ReleaseDate    *time.Time
	CorrelationIDs []string
}

// ToApplicationTemplateVersion missing godoc
func (a *ApplicationTemplateVersionInput) ToApplicationTemplateVersion(id, appTemplateID string) ApplicationTemplateVersion {
	if a == nil {
		return ApplicationTemplateVersion{}
	}

	return ApplicationTemplateVersion{
		ID:                    id,
		Version:               a.Version,
		Title:                 a.Title,
		ReleaseDate:           a.ReleaseDate,
		CorrelationIDs:        a.CorrelationIDs,
		ApplicationTemplateID: appTemplateID,
	}
}
