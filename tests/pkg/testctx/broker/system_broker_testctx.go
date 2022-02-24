package broker

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type SystemBrokerTestContext struct {
	Tenant  string
	Context context.Context

	SystemBrokerURL                string
	DirectorURL                    string
	DirectorExternalCertSecuredURL string
	ORDServiceURL                  string

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

	for cc.Get() == nil {
		log.D().Info("Waiting for certificate cache to load, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}

	return &SystemBrokerTestContext{
		Tenant:                         cfg.Tenant,
		Context:                        context.Background(),
		SystemBrokerURL:                cfg.SystemBrokerURL,
		DirectorURL:                    cfg.DirectorURL,
		DirectorExternalCertSecuredURL: cfg.DirectorExternalCertSecuredURL,
		ORDServiceURL:                  cfg.ORDServiceURL,
		ClientKey:                      clientKey,
		ConnectorTokenSecuredClient:    clients.NewTokenSecuredClient(cfg.ConnectorURL),
		CertSecuredGraphQLClient:       gql.NewCertAuthorizedGraphQLClientWithCustomURL(cfg.DirectorExternalCertSecuredURL, cc.Get().PrivateKey, cc.Get().Certificate, cfg.SkipSSLValidation),
	}, nil
}
