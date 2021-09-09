package tenantmapping

import (
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
)

const (
	clientCredentialScopesPrefix = "clientCredentialsRegistrationScopes"
)

// TenantContext missing godoc
type TenantContext struct {
	ExternalTenantID string
	TenantID         string
}

// NewTenantContext missing godoc
func NewTenantContext(externalTenantID, tenantID string) TenantContext {
	return TenantContext{
		ExternalTenantID: externalTenantID,
		TenantID:         tenantID,
	}
}

// ObjectContext missing godoc
type ObjectContext struct {
	TenantContext
	Scopes       string
	ConsumerID   string
	ConsumerType consumer.ConsumerType
}

// NewObjectContext missing godoc
func NewObjectContext(tenantCtx TenantContext, scopes, consumerID string, consumerType consumer.ConsumerType) ObjectContext {
	return ObjectContext{
		TenantContext: tenantCtx,
		Scopes:        scopes,
		ConsumerID:    consumerID,
		ConsumerType:  consumerType,
	}
}
