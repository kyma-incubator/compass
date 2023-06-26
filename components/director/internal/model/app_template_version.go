package model

import (
	"encoding/json"
	"time"
)

// ApplicationTemplateVersion represents a struct for Application Template Version data
type ApplicationTemplateVersion struct {
	ID                    string
	Version               string
	Title                 *string
	ReleaseDate           *string
	CorrelationIDs        json.RawMessage
	CreatedAt             time.Time
	ApplicationTemplateID string
}

// ApplicationTemplateVersionInput represents a struct with the updatable fields
type ApplicationTemplateVersionInput struct {
	Version        string
	Title          *string
	ReleaseDate    *string
	CorrelationIDs json.RawMessage
}

// ToApplicationTemplateVersion converts ApplicationTemplateVersionInput into ApplicationTemplateVersion
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
