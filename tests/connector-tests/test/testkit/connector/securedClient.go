package connector

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"github.com/sirupsen/logrus"

	schema "github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type SecuredClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func NewSecuredConnectorClient(endpoint string, key *rsa.PrivateKey, certificates ...*x509.Certificate) *SecuredClient {
	rawCerts := make([][]byte, len(certificates))
	for i, c := range certificates {
		rawCerts[i] = c.Raw
	}

	tlsCert := tls.Certificate{
		Certificate: rawCerts,
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

	graphQlClient.Log = func(s string) {
		logrus.Info(s)
	}

	return &SecuredClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c SecuredClient) Configuration(headers ...http.Header) (schema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := gcli.NewRequest(query)

	var response ConfigurationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c SecuredClient) RenewCert(csr string, headers ...http.Header) (schema.CertificationResult, error) {
	query := c.queryProvider.generateCert(csr)
	req := gcli.NewRequest(query)

	var response CertificationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

func (c SecuredClient) RevokeCertificate() (bool, error) {
	query := c.queryProvider.revokeCert()
	req := gcli.NewRequest(query)

	var response RevokeResult

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return false, errors.Wrap(err, "Failed to revoke certificate")
	}
	return response.Result, nil
}
