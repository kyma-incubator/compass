package model

import "github.com/kyma-incubator/compass/components/director/pkg/tenant"

type TenantModel struct {
	ID             string
	TenantId       string
	CustomerId     string
	Subdomain      string
	TenantProvider string
	Status         tenant.TenantStatus
}
