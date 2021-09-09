package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// IntegrationSystem missing godoc
type IntegrationSystem struct {
	ID          string
	Name        string
	Description *string
}

// IntegrationSystemPage missing godoc
type IntegrationSystemPage struct {
	Data       []*IntegrationSystem
	PageInfo   *pagination.Page
	TotalCount int
}

// IntegrationSystemInput missing godoc
type IntegrationSystemInput struct {
	Name        string
	Description *string
}

// ToIntegrationSystem missing godoc
func (i *IntegrationSystemInput) ToIntegrationSystem(id string) IntegrationSystem {
	if i == nil {
		return IntegrationSystem{}
	}

	return IntegrationSystem{
		ID:          id,
		Name:        i.Name,
		Description: i.Description,
	}
}
