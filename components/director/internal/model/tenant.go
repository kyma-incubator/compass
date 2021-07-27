package model

import "github.com/kyma-incubator/compass/components/director/pkg/tenant"

type TenantModel struct {
	ID               string
	Name             string
	TenantId         string
	ParentInternalId string
	ParentExternalId string
	Subdomain        string
	Type             tenant.Type
	Provider         string
	Status           tenant.Status
}
