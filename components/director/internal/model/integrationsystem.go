package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type IntegrationSystem struct {
	ID          string
	Name        string
	Description *string
}

type IntegrationSystemPage struct {
	Data       []*IntegrationSystem
	PageInfo   *pagination.Page
	TotalCount int
}

type IntegrationSystemInput struct {
	Name        string
	Description *string
}

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
