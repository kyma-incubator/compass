package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

type EventDefinition struct {
	Tenant      string
	BundleID    string
	Name        string
	Description *string
	Group       *string
	Version     *Version
	*BaseEntity
}

func (_ *EventDefinition) GetType() string {
	return resource.EventDefinition.ToLower()
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
		BundleID:    bndlID,
		Tenant:      tenant,
		Name:        e.Name,
		Description: e.Description,
		Group:       e.Group,
		Version:     e.Version.ToVersion(),
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}
