package tenantmapping

import (
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
)

const (
	scopesPerConsumerTypePrefix = "scopesPerConsumerType"
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
	Scopes          string
	Region          string
	OauthClientID   string
	ConsumerID      string
	AuthFlow        oathkeeper.AuthFlow
	ConsumerType    consumer.ConsumerType
	ContextProvider string
}

// KeysExtra contains the keys that should be used for Tenant and ExternalTenant in the IDToken claims
type KeysExtra struct {
	TenantKey         string
	ExternalTenantKey string
}

// NewObjectContext missing godoc
func NewObjectContext(tenantCtx TenantContext, keysExtra KeysExtra, scopes string, region string, clientID string, consumerID string, authFlow oathkeeper.AuthFlow, consumerType consumer.ConsumerType, contextProvider string) ObjectContext {
	return ObjectContext{
		TenantContext:   tenantCtx,
		KeysExtra:       keysExtra,
		Scopes:          scopes,
		Region:          region,
		OauthClientID:   clientID,
		ConsumerID:      consumerID,
		AuthFlow:        authFlow,
		ConsumerType:    consumerType,
		ContextProvider: contextProvider,
	}
}
