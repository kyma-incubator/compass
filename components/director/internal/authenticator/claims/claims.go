package claims

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/pkg/errors"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
)

// Claims missing godoc
type Claims struct {
	Tenant        map[string]string     `json:"tenant"`
	Scopes        string                `json:"scopes"`
	ConsumerID    string                `json:"consumerID"`
	ConsumerType  consumer.ConsumerType `json:"consumerType"`
	OnBehalfOf    string                `json:"onBehalfOf"`
	Region        string                `json:"region"`
	TokenClientID string                `json:"tokenClientID"`
	Flow          oathkeeper.AuthFlow   `json:"flow"`
	ZID           string                `json:"zid"`
	jwt.StandardClaims
}

// UnmarshalJSON implements Unmarshaler interface. The method unmarshal the data from b into Claims structure.
func (c *Claims) UnmarshalJSON(b []byte) error {
	tokenClaims := struct {
		TenantString  string                `json:"tenant"`
		Scopes        string                `json:"scopes"`
		ConsumerID    string                `json:"consumerID"`
		ConsumerType  consumer.ConsumerType `json:"consumerType"`
		OnBehalfOf    string                `json:"onBehalfOf"`
		Region        string                `json:"region"`
		TokenClientID string                `json:"tokenClientID"`
		Flow          oathkeeper.AuthFlow   `json:"flow"`
		ZID           string                `json:"zid"`
		jwt.StandardClaims
	}{}

	err := json.Unmarshal(b, &tokenClaims)
	if err != nil {
		return errors.Wrap(err, "while unmarshaling token claims:")
	}

	c.Scopes = tokenClaims.Scopes
	c.ConsumerID = tokenClaims.ConsumerID
	c.ConsumerType = tokenClaims.ConsumerType
	c.OnBehalfOf = tokenClaims.OnBehalfOf
	c.Region = tokenClaims.Region
	c.TokenClientID = tokenClaims.TokenClientID
	c.Flow = tokenClaims.Flow
	c.ZID = tokenClaims.ZID
	c.StandardClaims = tokenClaims.StandardClaims

	if err := json.Unmarshal([]byte(tokenClaims.TenantString), &c.Tenant); err != nil {
		log.D().Warnf("While unmarshaling tenants: %+v", err)
		c.Tenant = make(map[string]string)
	}

	return nil
}

// ContextWithClaims missing godoc
func (c *Claims) ContextWithClaims(ctx context.Context) context.Context {
	ctxWithTenants := tenant.SaveToContext(ctx, c.Tenant[tenantmapping.ConsumerTenantKey], c.Tenant[tenantmapping.ExternalTenantKey])
	scopesArray := strings.Split(c.Scopes, " ")
	ctxWithScopes := scope.SaveToContext(ctxWithTenants, scopesArray)
	apiConsumer := consumer.Consumer{ConsumerID: c.ConsumerID, ConsumerType: c.ConsumerType, Flow: c.Flow, OnBehalfOf: c.OnBehalfOf}
	ctxWithConsumerInfo := consumer.SaveToContext(ctxWithScopes, apiConsumer)
	return ctxWithConsumerInfo
}
