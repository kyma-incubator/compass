package broker

import (
	"context"
	"crypto/rsa"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type SystemBrokerTestContext struct {
	Tenant  string
	Context context.Context

	SystemBrokerURL string
	ORDServiceURL   string

	ClientKey *rsa.PrivateKey

	ConnectorTokenSecuredClient *clients.TokenSecuredClient
	CertSecuredGraphQLClient    *gcli.Client
}

func NewSystemBrokerTestContext(cfg config.SystemBrokerTestConfig) (*SystemBrokerTestContext, error) {
	clientKey, err := certs.GenerateKey()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate private key")
	}

	ctx := context.Background()
	cc, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while starting cert cache")
	}

	if err := util.WaitForCache(cc); err != nil {
		return nil, err
	}

	return &SystemBrokerTestContext{
		Tenant:                      cfg.Tenant,
		Context:                     context.Background(),
		SystemBrokerURL:             cfg.SystemBrokerURL,
		ORDServiceURL:               cfg.ORDServiceURL,
		ClientKey:                   clientKey,
		ConnectorTokenSecuredClient: clients.NewTokenSecuredClient(cfg.ConnectorURL),
		CertSecuredGraphQLClient:    gql.NewCertAuthorizedGraphQLClientWithCustomURL(cfg.DirectorExternalCertSecuredURL, cc.Get()[cfg.ExternalClientCertSecretName].PrivateKey, cc.Get()[cfg.ExternalClientCertSecretName].Certificate, cfg.SkipSSLValidation),
	}, nil
}
