package claims

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"strings"

	"github.com/pkg/errors"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
)

// Claims missing godoc
type Claims struct {
	Tenant          map[string]string   `json:"tenant"`
	Scopes          string              `json:"scopes"`
	Consumers       []consumer.Consumer `json:"consumers"`
	ZID             string              `json:"zid"`
	jwt.StandardClaims
}

// UnmarshalJSON parses the bearerToken using keyFunc. After the token's claims are extracted the "consumers" claim is unmarshaled into the Claim's slice of
// "Consumers" and the "tenant" claim is unmarshaled into Claim's map "Tenant"
func (c *Claims) UnmarshalJSON(ctx context.Context, bearerToken string, keyFunc func(token *jwt.Token) (interface{}, error)) (err error) {
	temp := struct {
		TenantString    string `json:"tenant"`
		Scopes          string `json:"scopes"`
		ConsumersString string `json:"consumers"`
		ZID             string `json:"zid"`
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

	if err := json.Unmarshal(marshaled, &temp); err != nil {
		return errors.Wrap(err, "while unmarshaling token claims:")
	}

	c.Scopes = temp.Scopes
	c.ZID = temp.ZID
	c.StandardClaims = temp.StandardClaims

	if err := json.Unmarshal([]byte(temp.ConsumersString), &c.Consumers); err != nil {
		log.C(ctx).Warnf("While unmarshaling consumers: %v", err)
	}

	if err := json.Unmarshal([]byte(temp.TenantString), &c.Tenant); err != nil {
		log.C(ctx).Warnf("While unmarshaling tenants: %v", err)
	}

	return nil
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
