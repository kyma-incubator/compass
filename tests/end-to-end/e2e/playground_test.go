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
	"os"
	"testing"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/connector"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func TestDirectorPlaygroundAccess(t *testing.T) {
	const (
		DirectorURLformat   = "https://%s.%s/director"
		OAuth2Subdomain     = "compass-gateway-auth-oauth"
		JWTSubdomain        = "compass-gateway"
		ClientCertSubdomain = "compass-gateway-mtls"
	)

	domain := os.Getenv("DOMAIN")
	require.NotEmpty(t, domain)

	t.Run("Access playground via OAuth2 subdomain", func(t *testing.T) {
		client := getClient()
		url := fmt.Sprintf(DirectorURLformat, OAuth2Subdomain, domain)
		resp, err := getPlaygroundWithRetries(client, url)
		require.NoError(t, err)

		defer closeBody(t, resp.Body)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Access playground via JWT subdomain", func(t *testing.T) {
		client := getClient()
		url := fmt.Sprintf(DirectorURLformat, JWTSubdomain, domain)
		resp, err := getPlaygroundWithRetries(client, url)
		require.NoError(t, err)

		defer closeBody(t, resp.Body)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Access playground via client certificate subdomain", func(t *testing.T) {
		ctx := context.Background()

		tenant := os.Getenv("DEFAULT_TENANT")
		require.NotEmpty(t, tenant)

		dexToken := getDexToken(t)
		dexGQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

		appID := createApplicationForCertPlaygroundTest(t, ctx, tenant, dexGQLClient)
		defer deleteApplication(t, ctx, dexGQLClient, tenant, appID)

		oneTimeToken := generateOneTimeTokenForApplication(t, ctx, dexGQLClient, tenant, appID)

		certChain, clientKey := generateClientCertForApplication(t, oneTimeToken)

		client := getClientWithCert(certChain, clientKey)
		url := fmt.Sprintf(DirectorURLformat, ClientCertSubdomain, domain)
		resp, err := getPlaygroundWithRetries(client, url)
		require.NoError(t, err)

		defer closeBody(t, resp.Body)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func getPlaygroundWithRetries(client *http.Client, url string) (*http.Response, error) {
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
