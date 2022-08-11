package tenantmapping

import (
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
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
	Scopes              string
	ScopesMergeStrategy scopesMergeStrategy
	Region              string
	OauthClientID       string
	ConsumerID          string
	AuthFlow            oathkeeper.AuthFlow
	ConsumerType        consumer.ConsumerType
	ContextProvider     string
}

type scopesMergeStrategy string

const (
	overrideAllScopes        scopesMergeStrategy = "overrideAllScopes"
	mergeWithOtherScopes     scopesMergeStrategy = "mergeWithOtherScopes"
	intersectWithOtherScopes scopesMergeStrategy = "intersectWithOtherScopes"
)

// KeysExtra contains the keys that should be used for Tenant and ExternalTenant in the IDToken claims
type KeysExtra struct {
	TenantKey         string
	ExternalTenantKey string
}

// NewObjectContext missing godoc
func NewObjectContext(tenantCtx TenantContext, keysExtra KeysExtra, scopes string, scopesMergeStrategy scopesMergeStrategy, region, clientID, consumerID string, authFlow oathkeeper.AuthFlow, consumerType consumer.ConsumerType, contextProvider string) ObjectContext {
	return ObjectContext{
		TenantContext:       tenantCtx,
		KeysExtra:           keysExtra,
		Scopes:              scopes,
		ScopesMergeStrategy: scopesMergeStrategy,
		Region:              region,
		OauthClientID:       clientID,
		ConsumerID:          consumerID,
		AuthFlow:            authFlow,
		ConsumerType:        consumerType,
		ContextProvider:     contextProvider,
	}
}
