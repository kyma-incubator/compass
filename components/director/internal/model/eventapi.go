package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type EventDefinition struct {
	ID          string
	Tenant      string
	BundleID    string
	Name        string
	Description *string
	Group       *string
	Version     *Version
}

type EventDefinitionPage struct {
	Data       []*EventDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (EventDefinitionPage) IsPageable() {}

type EventDefinitionInput struct {
	Name        string
	Description *string
	Group       *string
	Version     *VersionInput
}

func (e *EventDefinitionInput) ToEventDefinitionWithinBundle(id string, bndlID string, tenant string) *EventDefinition {
	if e == nil {
		return nil
	}

	return &EventDefinition{
		ID:          id,
		BundleID:    bndlID,
		Tenant:      tenant,
		Name:        e.Name,
		Description: e.Description,
		Group:       e.Group,
		Version:     e.Version.ToVersion(),
	}
}
