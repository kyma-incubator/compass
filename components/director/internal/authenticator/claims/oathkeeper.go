package claims

import (
	"context"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
)

type oathkeeperClaims struct {
	Tenant         string                `json:"tenant"`
	ExternalTenant string                `json:"externalTenant"`
	Scopes         string                `json:"scopes"`
	ConsumerID     string                `json:"consumerID"`
	ConsumerType   consumer.ConsumerType `json:"consumerType"`
	Flow           oathkeeper.AuthFlow   `json:"flow"`
	ZID            string                `json:"zid"`
	jwt.StandardClaims
}

func (c *oathkeeperClaims) ContextWithClaims(ctx context.Context) context.Context {
	ctxWithTenants := tenant.SaveToContext(ctx, c.Tenant, c.ExternalTenant)
	scopesArray := strings.Split(c.Scopes, " ")
	ctxWithScopes := scope.SaveToContext(ctxWithTenants, scopesArray)
	apiConsumer := consumer.Consumer{ConsumerID: c.ConsumerID, ConsumerType: c.ConsumerType, Flow: c.Flow}
	ctxWithConsumerInfo := consumer.SaveToContext(ctxWithScopes, apiConsumer)
	return ctxWithConsumerInfo
}

type oathkeeperClaimsParser struct {
}

func NewOathkeeperClaimsParser() *oathkeeperClaimsParser {
	return &oathkeeperClaimsParser{}
}

func (p *oathkeeperClaimsParser) ParseClaims(ctx context.Context, bearerToken string, keyfunc jwt.Keyfunc) (Claims, error) {
	parsed := oathkeeperClaims{}
	_, err := jwt.ParseWithClaims(bearerToken, &parsed, keyfunc)
	if err != nil {
		return nil, err
	}

	if err := p.validateClaims(ctx, parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}

func (*oathkeeperClaimsParser) validateClaims(ctx context.Context, claims oathkeeperClaims) error {
	if err := claims.Valid(); err != nil {
		return err
	}

	if claims.Tenant == "" && claims.ExternalTenant != "" {
		log.C(ctx).Errorf("Tenant not found in auth claims")
		return apperrors.NewTenantNotFoundError(claims.ExternalTenant)
	}

	return nil
}
