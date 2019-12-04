package e2e

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/connector"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

type playgroundTestConfig struct {
	Gateway struct {
		Domain               string `envconfig:"DOMAIN"`
		JWTSubdomain         string
		OAuth20Subdomain     string `envconfig:"GATEWAY_OAUTH2_SUBDOMAIN"`
		ClientCertsSubdomain string
	}
	DirectorURLFormat          string `envconfig:"default=https://%s.%s/director"`
	DirectorGraphQLExamplePath string `envconfig:"default=examples/create-application/create-application.graphql"`
	DefaultTenant              string
}

type playgroundTestSuite struct {
	t          *testing.T
	client     *http.Client
	urlBuilder *playgroundURLBuilder
	subdomain  string
}

func newPlaygroundTestSuite(t *testing.T, cfg *playgroundTestConfig, subdomain string) *playgroundTestSuite {
	urlBuilder := newPlaygroundURLBuilder(cfg)
	return &playgroundTestSuite{t: t, urlBuilder: urlBuilder, subdomain: subdomain, client: getClient()}
}

func (ts *playgroundTestSuite) setHTTPClient(client *http.Client) {
	ts.client = client
}

func (ts *playgroundTestSuite) checkDirectorPlaygroundWithRedirection() {
	resp, err := getURLWithRetries(ts.client, ts.urlBuilder.getRedirectionStartURL(ts.subdomain))
	require.NoError(ts.t, err)
	defer closeBody(ts.t, resp.Body)

	assert.Equal(ts.t, ts.urlBuilder.getFinalURL(ts.subdomain), resp.Request.URL.String()) // test redirection to URL with trailing slash
	assert.Equal(ts.t, http.StatusOK, resp.StatusCode)
}

func (ts *playgroundTestSuite) checkDirectorGraphQLExample() {
	resp, err := getURLWithRetries(ts.client, ts.urlBuilder.getGraphQLExampleURL(ts.subdomain))
	require.NoError(ts.t, err)
	defer closeBody(ts.t, resp.Body)

	assert.Equal(ts.t, http.StatusOK, resp.StatusCode)
}

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

		dexToken := getDexToken(t)
		dexGQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

		ctx := context.Background()
		appID := createApplicationForCertPlaygroundTest(t, ctx, tenant, dexGQLClient)
		defer deleteApplication(t, ctx, dexGQLClient, tenant, appID)

		oneTimeToken := generateOneTimeTokenForApplication(t, ctx, dexGQLClient, tenant, appID)
		certChain, clientKey := generateClientCertForApplication(t, oneTimeToken)
		client := getClientWithCert(certChain, clientKey)

		testSuite := newPlaygroundTestSuite(t, cfg, subdomain)
		testSuite.setHTTPClient(client)

		testSuite.checkDirectorPlaygroundWithRedirection()
		testSuite.checkDirectorGraphQLExample()
	})
}

func getURLWithRetries(client *http.Client, url string) (*http.Response, error) {
	const (
		maxAttempts = 10
		delay       = 10
	)
	var resp *http.Response

	happyRun := true
	err := retry.Do(
		func() error {
			_resp, err := client.Get(url)
			if err != nil {
				return err
			}

			if _resp.StatusCode >= 400 {
				return fmt.Errorf("got status code %d when accessing %s", _resp.StatusCode, url)
			}

			resp = _resp

			return nil
		},
		retry.Attempts(maxAttempts),
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(retryNo uint, err error) {
			happyRun = false
			log.Printf("Retry: [%d / %d], error: %s", retryNo, maxAttempts, err)
		}),
	)

	if err != nil {
		return nil, err
	}

	if happyRun {
		log.Printf("Address %s reached successfully", url)
	}

	return resp, nil
}

func getClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}
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

func closeBody(t *testing.T, body io.ReadCloser) {
	err := body.Close()
	require.NoError(t, err)
}

func getDexToken(t *testing.T) string {
	config, err := idtokenprovider.LoadConfig()
	require.NoError(t, err)

	dexToken, err := idtokenprovider.Authenticate(config.IdProviderConfig)
	require.NoError(t, err)

	return dexToken
}

func createApplicationForCertPlaygroundTest(t *testing.T, ctx context.Context, tenant string, cli *gcli.Client) string {
	appInput := graphql.ApplicationCreateInput{
		Name: "cert-playground-test",
	}
	app := createApplicationFromInputWithinTenant(t, ctx, cli, tenant, appInput)
	require.NotEmpty(t, app.ID)

	return app.ID
}

func generateClientCertForApplication(t *testing.T, oneTimeToken graphql.OneTimeToken) ([]*x509.Certificate, *rsa.PrivateKey) {
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
