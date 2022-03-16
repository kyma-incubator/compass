package clients

import (
	"context"
	"crypto"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	gcli "github.com/machinebox/graphql"
)

type CertSecuredGraphQLClient struct {
	ctx                      context.Context
	tenant                   string
	CertSecuredGraphqlClient *gcli.Client
}

func NewCertSecuredGraphQLClient(ctx context.Context, url, tenant string, key crypto.PrivateKey, certChain [][]byte, skipSSLValidation bool) (*CertSecuredGraphQLClient, error) {
	return &CertSecuredGraphQLClient{
		ctx:                      ctx,
		tenant:                   tenant,
		CertSecuredGraphqlClient: gql.NewCertAuthorizedGraphQLClientWithCustomURL(url, key, certChain, skipSSLValidation),
	}, nil
}

func (c *CertSecuredGraphQLClient) GenerateApplicationToken(t *testing.T, appID string) (externalschema.Token, error) {
	tokenExt := fixtures.GenerateOneTimeTokenForApplication(t, c.ctx, c.CertSecuredGraphqlClient, c.tenant, appID)
	return externalschema.Token{Token: tokenExt.Token}, nil
}

func (c *CertSecuredGraphQLClient) GenerateRuntimeToken(t *testing.T, runtimeID string) (externalschema.Token, error) {
	tokenExt := fixtures.RequestOneTimeTokenForRuntime(t, c.ctx, c.CertSecuredGraphqlClient, c.tenant, runtimeID)
	return externalschema.Token{Token: tokenExt.Token}, nil
}
