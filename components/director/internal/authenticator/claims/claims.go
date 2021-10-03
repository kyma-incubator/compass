package claims

import (
	"context"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
)

// Claims missing godoc
type Claims struct {
	TenantString    string              `json:"tenant"`
	Tenant          map[string]string   `json:"-"`
	Scopes          string              `json:"scopes"`
	ConsumersString string              `json:"consumers"`
	Consumers       []consumer.Consumer `json:"-"`
	ZID             string              `json:"zid"`
	jwt.StandardClaims
}

// ContextWithClaims missing godoc
func (c *Claims) ContextWithClaims(ctx context.Context) context.Context {
	ctxWithTenants := tenant.SaveToContext(ctx, c.Tenant["consumerTenant"], c.Tenant["externalTenant"])
	scopesArray := strings.Split(c.Scopes, " ")
	ctxWithScopes := scope.SaveToContext(ctxWithTenants, scopesArray)
	apiConsumer := consumer.Consumer{}
	if len(c.Consumers) > 0 {
		apiConsumer = consumer.Consumer{ConsumerID: c.Consumers[0].ConsumerID, ConsumerType: c.Consumers[0].ConsumerType, Flow: c.Consumers[0].Flow}
	}
	ctxWithConsumerInfo := consumer.SaveToContext(ctxWithScopes, apiConsumer)
	return ctxWithConsumerInfo
}
