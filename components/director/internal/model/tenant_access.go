package model

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// TenantAccess represents the tenant's access level to the resource
type TenantAccess struct {
	ExternalTenantID string
	InternalTenantID string
	ResourceType     resource.Type
	ResourceID       string
	Owner            bool
}
