package clientset

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type CertificateSecuredClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func newCertificateSecuredConnectorClient(endpoint string, tlsCert tls.Certificate) *CertificateSecuredClient {
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: true,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &CertificateSecuredClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c CertificateSecuredClient) Configuration(headers ...http.Header) (externalschema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := newRequest(query, headers...)

	var response ConfigurationResponse
	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return externalschema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c CertificateSecuredClient) SignCSR(csr string, headers ...http.Header) (externalschema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)
	req := newRequest(query, headers...)

	var response CertificationResponse
	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return externalschema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

func (c CertificateSecuredClient) RevokeCertificate(headers ...http.Header) (bool, error) {
	query := c.queryProvider.revokeCert()
	req := newRequest(query, headers...)

	var response RevokeResult
	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return false, errors.Wrap(err, "Failed to revoke certificate")
	}
	return response.Result, nil
}
