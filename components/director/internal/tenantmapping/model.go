package tenantmapping

import (
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
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
	KeysExtra
	Scopes       string
	ConsumerID   string
	AuthFlow     oathkeeper.AuthFlow
	ConsumerType consumer.ConsumerType
}

// KeysExtra contains the keys that should be used for Tenant and ExternalTenant in the IDToken claims
type KeysExtra struct {
	TenantKey         string
	ExternalTenantKey string
}

// NewObjectContext missing godoc
func NewObjectContext(tenantCtx TenantContext, keysExtra KeysExtra, scopes string, consumerID string, authFlow oathkeeper.AuthFlow, consumerType consumer.ConsumerType) ObjectContext {
	return ObjectContext{
		TenantContext: tenantCtx,
		KeysExtra:     keysExtra,
		Scopes:        scopes,
		ConsumerID:    consumerID,
		AuthFlow:      authFlow,
		ConsumerType:  consumerType,
	}
}
