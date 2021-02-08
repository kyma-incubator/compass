package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	ID          string
	BundleID    string
	Tenant      string
	Name        string
	Description *string
	TargetURL   string
	//  group allows you to find the same API but in different version
	Group     *string
	Version   *Version
	Ready     bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Error     *string
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
		ID:          id,
		BundleID:    bundleID,
		Tenant:      tenant,
		Name:        a.Name,
		Description: a.Description,
		TargetURL:   a.TargetURL,
		Group:       a.Group,
		Version:     a.Version.ToVersion(),
		Ready:       true,
	}
}
