package testkit

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"net/http"
)

//Currently unused. SecuredConnectorClient will be used in future test cases
type SecuredConnectorClient interface {
	Configuration() (schema.Configuration, error)
	RenewCertificate(csr string) (schema.CertificationResult, error)
	RevokeCertificate() (bool, error)
}

type securedClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func NewSecuredConnectorClient(endpoint string, key *rsa.PrivateKey, certificate ...[]byte) ConnectorClient {
	tlsCert := tls.Certificate{
		Certificate: certificate,
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: true,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &client{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c securedClient) Configuration() (schema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := gcli.NewRequest(query)

	var response ConfigurationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c securedClient) RenewCert(csr string, token string) (schema.CertificationResult, error) {
	query := c.queryProvider.generateCert(csr)
	req := gcli.NewRequest(query)

	var response CertificationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

func (c securedClient) RevokeCertificate() (bool, error) {
	query := c.queryProvider.revokeCert()
	req := gcli.NewRequest(query)

	var response RevokeResult

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return false, errors.Wrap(err, "Failed to revoke certificate")
	}
	return response.Result, nil
}
