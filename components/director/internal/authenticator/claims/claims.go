package claims

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
)

// Claims missing godoc
type Claims struct {
	Tenant       map[string]string     `json:"tenant"`
	Scopes       string                `json:"scopes"`
	ConsumerID   string                `json:"consumerID"`
	ConsumerType consumer.ConsumerType `json:"consumerType"`
	OnBehalfOf   string                `json:"onBehalfOf"`
	Flow         oathkeeper.AuthFlow   `json:"flow"`
	ZID          string                `json:"zid"`
	jwt.StandardClaims
}

// UnmarshalJSONClaims parses the bearerToken using keyFunc. After the token's claims are extracted the "tenant" claim is unmarshaled into Claim's map "Tenant"
func (c *Claims) UnmarshalJSONClaims(ctx context.Context, bearerToken string, keyFunc func(token *jwt.Token) (interface{}, error)) error {
	tokenClaims := struct {
		TenantString string                `json:"tenant"`
		Scopes       string                `json:"scopes"`
		ConsumerID   string                `json:"consumerID"`
		ConsumerType consumer.ConsumerType `json:"consumerType"`
		OnBehalfOf   string                `json:"onBehalfOf"`
		Flow         oathkeeper.AuthFlow   `json:"flow"`
		ZID          string                `json:"zid"`
		jwt.StandardClaims
	}{}

	token, err := jwt.Parse(bearerToken, keyFunc)
	if err != nil {
		return err
	}

	marshaled, err := json.Marshal(token.Claims)
	if err != nil {
		return errors.Wrap(err, "while marshaling token claims:")
	}

	if err := json.Unmarshal(marshaled, &tokenClaims); err != nil {
		return errors.Wrap(err, "while unmarshaling token claims:")
	}

	c.Scopes = tokenClaims.Scopes
	c.ConsumerID = tokenClaims.ConsumerID
	c.ConsumerType = tokenClaims.ConsumerType
	c.OnBehalfOf = tokenClaims.OnBehalfOf
	c.Flow = tokenClaims.Flow
	c.ZID = tokenClaims.ZID
	c.StandardClaims = tokenClaims.StandardClaims

	if err := json.Unmarshal([]byte(tokenClaims.TenantString), &c.Tenant); err != nil {
		log.C(ctx).Warnf("While unmarshaling tenants: %v", err)
	}

	return nil
}

// ContextWithClaims missing godoc
func (c *Claims) ContextWithClaims(ctx context.Context) context.Context {
	ctxWithTenants := tenant.SaveToContext(ctx, c.Tenant[tenantmapping.ConsumerTenantKey], c.Tenant[tenantmapping.ExternalTenantKey])
	scopesArray := strings.Split(c.Scopes, " ")
	ctxWithScopes := scope.SaveToContext(ctxWithTenants, scopesArray)
	apiConsumer := consumer.Consumer{}
	apiConsumer = consumer.Consumer{ConsumerID: c.ConsumerID, ConsumerType: c.ConsumerType, Flow: c.Flow, OnBehalfOf: c.OnBehalfOf}
	ctxWithConsumerInfo := consumer.SaveToContext(ctxWithScopes, apiConsumer)
	return ctxWithConsumerInfo
}
