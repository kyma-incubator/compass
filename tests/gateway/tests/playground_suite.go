package tests

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/avast/retry-go/v4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type playgroundTestConfig struct {
	Gateway struct {
		Domain               string `envconfig:"DOMAIN"`
		JWTSubdomain         string
		OAuth20Subdomain     string `envconfig:"GATEWAY_OAUTH20_SUBDOMAIN"`
		ClientCertsSubdomain string
	}
	DirectorURLFormat          string `envconfig:"default=https://%s.%s/director"`
	DirectorGraphQLExamplePath string `envconfig:"default=examples/register-application/register-application.graphql"`
	DefaultTestTenant          string
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

type playgroundURLBuilder struct {
	format             string
	domain             string
	graphQLExamplePath string
}

func newPlaygroundURLBuilder(cfg *playgroundTestConfig) *playgroundURLBuilder {
	return &playgroundURLBuilder{format: cfg.DirectorURLFormat, domain: cfg.Gateway.Domain, graphQLExamplePath: cfg.DirectorGraphQLExamplePath}
}

func (b *playgroundURLBuilder) getRedirectionStartURL(subdomain string) string {
	redirectionURL := fmt.Sprintf(b.format, subdomain, b.domain)
	return redirectionURL
}

func (b *playgroundURLBuilder) getFinalURL(subdomain string) string {
	finalURL := fmt.Sprintf("%s/", b.getRedirectionStartURL(subdomain))
	return finalURL
}

func (b *playgroundURLBuilder) getGraphQLExampleURL(subdomain string) string {
	return fmt.Sprintf("%s%s", b.getFinalURL(subdomain), b.graphQLExamplePath)
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

func closeBody(t *testing.T, body io.ReadCloser) {
	err := body.Close()
	require.NoError(t, err)
}
