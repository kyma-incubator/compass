package clients

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	gcli "github.com/machinebox/graphql"
)

type StaticUserClient struct {
	ctx              context.Context
	tenant           string
	DexGraphqlClient *gcli.Client
}

func NewStaticUserClient(ctx context.Context, url string, tenant string) (*StaticUserClient, error) {
	token, err := idtokenprovider.GetDexToken()
	if err != nil {
		return nil, err
	}
	return &StaticUserClient{
		ctx:              ctx,
		tenant:           tenant,
		DexGraphqlClient: gql.NewAuthorizedGraphQLClientWithCustomURL(token, url),
	}, nil
}

func (c *StaticUserClient) GenerateApplicationToken(t *testing.T, appID string) (externalschema.Token, error) {
	tokenExt := fixtures.GenerateOneTimeTokenForApplication(t, c.ctx, c.DexGraphqlClient, c.tenant, appID)
	return externalschema.Token{Token: tokenExt.Token}, nil
}

func (c *StaticUserClient) GenerateRuntimeToken(t *testing.T, runtimeID string) (externalschema.Token, error) {
	tokenExt := fixtures.RequestOneTimeTokenForRuntime(t, c.ctx, c.DexGraphqlClient, c.tenant, runtimeID)
	return externalschema.Token{Token: tokenExt.Token}, nil
}
