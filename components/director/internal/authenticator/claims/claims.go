package claims

import (
	"context"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
)

type Claims struct {
	Tenant         string                `json:"tenant"`
	ExternalTenant string                `json:"externalTenant"`
	Scopes         string                `json:"scopes"`
	ConsumerID     string                `json:"consumerID"`
	ConsumerType   consumer.ConsumerType `json:"consumerType"`
	Flow           oathkeeper.AuthFlow   `json:"flow"`
	ZID            string                `json:"zid"`
	jwt.StandardClaims
}

func (c *Claims) ContextWithClaims(ctx context.Context) context.Context {
	ctxWithTenants := tenant.SaveToContext(ctx, c.Tenant, c.ExternalTenant)
	scopesArray := strings.Split(c.Scopes, " ")
	ctxWithScopes := scope.SaveToContext(ctxWithTenants, scopesArray)
	apiConsumer := consumer.Consumer{ConsumerID: c.ConsumerID, ConsumerType: c.ConsumerType, Flow: c.Flow}
	ctxWithConsumerInfo := consumer.SaveToContext(ctxWithScopes, apiConsumer)
	return ctxWithConsumerInfo
}
