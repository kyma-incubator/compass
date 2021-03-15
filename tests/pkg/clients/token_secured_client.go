package clients

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"testing"

	"net/http"

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

func NewTokenSecuredClient(endpoint string) *TokenSecuredClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &TokenSecuredClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c *TokenSecuredClient) Configuration(token string, headers ...http.Header) (externalschema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var response certs.ConfigurationResponse

	if err := c.graphQlClient.Run(context.Background(), req, &response); err != nil {
		return externalschema.Configuration{}, errors.Wrap(err, "failed to get configuration")
	}
	return response.Result, nil
}

func (c *TokenSecuredClient) SignCSR(csr string, token string, headers ...http.Header) (externalschema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var response certs.CertificationResponse

	if err := c.graphQlClient.Run(context.Background(), req, &response); err != nil {
		return externalschema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

func (c *TokenSecuredClient) GenerateAndSignCert(t *testing.T, certConfig externalschema.Configuration) (*externalschema.CertificationResult, *rsa.PrivateKey, error) {
	clientKey, err := certs.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	csr := certs.CreateCsr(t, certConfig.CertificateSigningRequestInfo.Subject, clientKey)
	if err != nil {
		return nil, nil, err
	}

	certResult, err := c.SignCSR(certs.EncodeBase64(csr), certConfig.Token.Token)
	if err != nil {
		return nil, nil, err
	}

	return &certResult, clientKey, nil
}
