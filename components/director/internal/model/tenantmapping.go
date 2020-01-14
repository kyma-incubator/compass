package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type TenantStatus string

const (
	Active   TenantStatus = "Active"
	Inactive TenantStatus = "Inactive"
)

type TenantMapping struct {
	ID             string
	Name           string
	ExternalTenant string
	InternalTenant string
	Provider       string
	Status         TenantStatus
}

type TenantMappingPage struct {
	Data       []*TenantMapping
	PageInfo   *pagination.Page
	TotalCount int
}

type TenantMappingInput struct {
	Name           string
	ExternalTenant string
	Provider       string
}
