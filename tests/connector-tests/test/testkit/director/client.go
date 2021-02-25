package director

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	director "github.com/kyma-incubator/compass/tests/director/gateway-integration"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	gcli "github.com/machinebox/graphql"
)

type Client struct {
	ctx              context.Context
	tenant           string
	DexGraphqlClient *gcli.Client
}

func NewClient(ctx context.Context, url string, tenant string) (*Client, error) {
	token, err := idtokenprovider.GetDexToken()
	if err != nil {
		return nil, err
	}
	return &Client{
		ctx:              ctx,
		tenant:           tenant,
		DexGraphqlClient: gql.NewAuthorizedGraphQLClientWithCustomURL(token, url),
	}, nil
}

func (c *Client) GenerateApplicationToken(t *testing.T, appID string) (externalschema.Token, error) {
	tokenExt := director.GenerateOneTimeTokenForApplication(t, c.ctx, c.DexGraphqlClient, c.tenant, appID)
	return externalschema.Token{Token: tokenExt.Token}, nil
}

func (c *Client) GenerateRuntimeToken(t *testing.T, runtimeID string) (externalschema.Token, error) {
	tokenExt := director.GenerateOneTimeTokenForRuntime(t, c.ctx, c.DexGraphqlClient, c.tenant, runtimeID)
	return externalschema.Token{Token: tokenExt.Token}, nil
}
