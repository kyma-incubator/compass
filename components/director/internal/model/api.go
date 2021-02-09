package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	BundleID    string
	Tenant      string
	Name        string
	Description *string
	TargetURL   string
	//  group allows you to find the same API but in different version
	Group   *string
	Version *Version
	*BaseEntity
}

func (_ *APIDefinition) GetType() string {
	return resource.API.ToLower()
}

type Timestamp time.Time

type APIDefinitionInput struct {
	Name        string
	Description *string
	TargetURL   string
	Group       *string
	Version     *VersionInput
}

type APIDefinitionPage struct {
	Data       []*APIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (APIDefinitionPage) IsPageable() {}

func (a *APIDefinitionInput) ToAPIDefinitionWithinBundle(id string, bundleID string, tenant string) *APIDefinition {
	if a == nil {
		return nil
	}

	return &APIDefinition{
		BundleID:    bundleID,
		Tenant:      tenant,
		Name:        a.Name,
		Description: a.Description,
		TargetURL:   a.TargetURL,
		Group:       a.Group,
		Version:     a.Version.ToVersion(),
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}
