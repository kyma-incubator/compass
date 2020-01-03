package auth_related_tests

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director-tests/pkg/connector"
	"github.com/kyma-incubator/compass/tests/director-tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director-tests/pkg/idtokenprovider"
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
		tenant := cfg.DefaultTenant

		dexToken, err := idtokenprovider.GetDexToken()
		require.NoError(t, err)
		dexGQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

		ctx := context.Background()
		appID := createApplicationForCertPlaygroundTest(t, ctx, tenant, dexGQLClient)
		defer unregisterApplication(t, ctx, dexGQLClient, tenant, appID)

		oneTimeToken := generateOneTimeTokenForApplication(t, ctx, dexGQLClient, tenant, appID)
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
	app := registerApplicationFromInputWithinTenant(t, ctx, cli, tenant, appInput)
	require.NotEmpty(t, app.ID)

	return app.ID
}

func generateClientCertForApplication(t *testing.T, oneTimeToken graphql.OneTimeTokenExt) ([]*x509.Certificate, *rsa.PrivateKey) {
	connectorClient := connector.NewClient(oneTimeToken.ConnectorURL)
	clientCertConfig, err := connectorClient.GetConfiguration(oneTimeToken.Token)
	require.NoError(t, err)
	require.NotEmpty(t, clientCertConfig.Token)
	require.NotEmpty(t, clientCertConfig.CertificateSigningRequestInfo.Subject)

	clientCert, clientKey, err := connectorClient.GenerateAndSignCert(clientCertConfig)
	require.NoError(t, err)
	require.NotEmpty(t, clientCert.CertificateChain)

	certChain, err := connector.DecodeCertChain(clientCert.CertificateChain)
	require.NoError(t, err)

	return certChain, clientKey
}
