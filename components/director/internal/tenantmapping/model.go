package tenantmapping

import (
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
)

const (
	clientCredentialScopesPrefix = "clientCredentialsRegistrationScopes"
)

type TenantContext struct {
	ExternalTenantID string
	TenantID         string
}

func NewTenantContext(externalTenantID, tenantID string) TenantContext {
	return TenantContext{
		ExternalTenantID: externalTenantID,
		TenantID:         tenantID,
	}
}

type ObjectContext struct {
	TenantContext
	Scopes       string
	ConsumerID   string
	ConsumerType consumer.ConsumerType
}

func NewObjectContext(tenantCtx TenantContext, scopes, consumerID string, consumerType consumer.ConsumerType) ObjectContext {
	return ObjectContext{
		TenantContext: tenantCtx,
		Scopes:        scopes,
		ConsumerID:    consumerID,
		ConsumerType:  consumerType,
	}
}
