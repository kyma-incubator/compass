package clientset

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	TokenHeader = "Connector-Token"
)

type TokenSecuredClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func newTokenSecuredClient(endpoint string, opts *clientsetOptions) *TokenSecuredClient {
	tr := httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.skipTLSVerify,
		},
	}))

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   opts.timeout,
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &TokenSecuredClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c *TokenSecuredClient) Configuration(ctx context.Context, token string, headers ...http.Header) (externalschema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := newRequest(query, headers...)
	req.Header.Add(TokenHeader, token)

	var response ConfigurationResponse

	err := c.graphQlClient.Run(ctx, req, &response)
	if err != nil {
		return externalschema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c *TokenSecuredClient) SignCSR(ctx context.Context, csr string, token string, headers ...http.Header) (externalschema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)
	req := newRequest(query, headers...)
	req.Header.Add(TokenHeader, token)

	var response CertificationResponse

	err := c.graphQlClient.Run(ctx, req, &response)
	if err != nil {
		return externalschema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

func newRequest(query string, headers ...http.Header) *gcli.Request {
	req := gcli.NewRequest(query)

	header := mergeHeaders(headers)
	req.Header = header

	return req
}

func mergeHeaders(headers []http.Header) http.Header {
	mergedHeaders := http.Header{}

	for _, header := range headers {
		for h, v := range header {
			mergedHeaders[h] = append(mergedHeaders[h], v...)
		}
	}

	return mergedHeaders
}

type ConfigurationResponse struct {
	Result externalschema.Configuration `json:"result"`
}

type CertificationResponse struct {
	Result externalschema.CertificationResult `json:"result"`
}

type RevokeResult struct {
	Result bool `json:"result"`
}
