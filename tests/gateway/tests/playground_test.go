package tests

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

func TestDirectorPlaygroundAccess(t *testing.T) {
	cfg := &playgroundTestConfig{}
	err := envconfig.Init(&cfg)
	require.NoError(t, err)

	t.Run("Access playground via OAuth2 subdomain", func(t *testing.T) {
		subdomain := cfg.Gateway.OAuth20Subdomain
		testSuite := newPlaygroundTestSuite(t, cfg, subdomain)

		testSuite.checkDirectorPlaygroundWithRedirection()
		testSuite.checkDirectorGraphQLExample()
	})

	t.Run("Access playground via JWT subdomain", func(t *testing.T) {
		subdomain := cfg.Gateway.JWTSubdomain
		testSuite := newPlaygroundTestSuite(t, cfg, subdomain)

		testSuite.checkDirectorPlaygroundWithRedirection()
		testSuite.checkDirectorGraphQLExample()
	})

	t.Run("Access playground via client certificate subdomain", func(t *testing.T) {
		subdomain := cfg.Gateway.ClientCertsSubdomain
		tenant := cfg.DefaultTestTenant

		ctx := context.Background()
		appID := createApplicationForCertPlaygroundTest(t, ctx, tenant, dexGraphQLClient)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, appID)

		oneTimeToken := fixtures.GenerateOneTimeTokenForApplication(t, ctx, dexGraphQLClient, tenant, appID)

		certChain, clientKey := generateClientCertForApplication(t, oneTimeToken)
		client := getClientWithCert(certChain, clientKey)

		testSuite := newPlaygroundTestSuite(t, cfg, subdomain)
		testSuite.setHTTPClient(client)

		testSuite.checkDirectorPlaygroundWithRedirection()
		testSuite.checkDirectorGraphQLExample()
	})
}

func getClientWithCert(certificates []*x509.Certificate, key *rsa.PrivateKey) *http.Client {
	rawCerts := make([][]byte, len(certificates))
	for i, c := range certificates {
		rawCerts[i] = c.Raw
	}

	tlsCert := tls.Certificate{
		Certificate: rawCerts,
		PrivateKey:  key,
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{tlsCert},
			ClientAuth:         tls.RequireAndVerifyClientCert,
			InsecureSkipVerify: true,
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}
}

func createApplicationForCertPlaygroundTest(t *testing.T, ctx context.Context, tenant string, cli *gcli.Client) string {
	appInput := graphql.ApplicationRegisterInput{
		Name: "cert-playground-test",
	}
	app, err := fixtures.RegisterApplicationFromInput(t, ctx, cli, tenant, appInput)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	return app.ID
}

func generateClientCertForApplication(t *testing.T, oneTimeToken graphql.OneTimeTokenForApplicationExt) ([]*x509.Certificate, *rsa.PrivateKey) {
	connectorClient := clients.NewTokenSecuredClient(oneTimeToken.ConnectorURL)
	clientCertConfig, err := connectorClient.Configuration(oneTimeToken.Token)
	require.NoError(t, err)
	require.NotEmpty(t, clientCertConfig.Token)
	require.NotEmpty(t, clientCertConfig.CertificateSigningRequestInfo.Subject)

	clientCert, clientKey, err := connectorClient.GenerateAndSignCert(t, clientCertConfig)
	require.NoError(t, err)
	require.NotEmpty(t, clientCert.CertificateChain)

	certChain := certs.DecodeCertChain(t, clientCert.CertificateChain)
	require.NoError(t, err)

	return certChain, clientKey
}
