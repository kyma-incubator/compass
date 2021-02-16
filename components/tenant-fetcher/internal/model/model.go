package model

import "github.com/kyma-incubator/compass/components/director/pkg/tenant"

type TenantModel struct {
	ID             string
	AccountId      string
	TenantProvider string
	Status         tenant.TenantStatus
}
