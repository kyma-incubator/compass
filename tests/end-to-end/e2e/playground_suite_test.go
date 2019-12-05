package e2e

import (
	"fmt"
	"net/http"
	"testing"

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
