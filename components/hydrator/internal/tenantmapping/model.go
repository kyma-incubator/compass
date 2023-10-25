package tenantmapping

import (
	"crypto/sha256"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
	//TenantContext
	KeysExtra
	Tenant              *graphql.Tenant
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
func NewObjectContext(tenant *graphql.Tenant, keysExtra KeysExtra, scopes string, scopesMergeStrategy scopesMergeStrategy, region, clientID, consumerID string, authFlow oathkeeper.AuthFlow, consumerType consumer.ConsumerType, contextProvider string) ObjectContext {
	return ObjectContext{
		Tenant:              tenant,
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

func RedactConsumerIDForLogging(original ObjectContext) ObjectContext {
	if original.ConsumerType == consumer.User {
		return ObjectContext{
			Tenant:              original.Tenant,
			KeysExtra:           original.KeysExtra,
			Scopes:              original.Scopes,
			ScopesMergeStrategy: original.ScopesMergeStrategy,
			Region:              original.Region,
			OauthClientID:       original.OauthClientID,
			ConsumerID:          fmt.Sprintf("REDACTED_%x", sha256.Sum256([]byte(original.ConsumerID))),
			AuthFlow:            original.AuthFlow,
			ConsumerType:        original.ConsumerType,
			ContextProvider:     original.ContextProvider,
		}
	}
	return original
}
