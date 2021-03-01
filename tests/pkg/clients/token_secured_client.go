package clients

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	var response ConfigurationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return externalschema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c *TokenSecuredClient) SignCSR(csr string, token string, headers ...http.Header) (externalschema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var response CertificationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return externalschema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

func (c *TokenSecuredClient) GenerateAndSignCert(t *testing.T, certConfig externalschema.Configuration) (externalschema.CertificationResult, *rsa.PrivateKey, error) {
	clientKey, err := certs.GenerateKey()
	if err != nil {
		return externalschema.CertificationResult{}, nil, err
	}

	csr := certs.CreateCsr(t, certConfig.CertificateSigningRequestInfo.Subject, clientKey)
	if err != nil {
		return externalschema.CertificationResult{}, nil, err
	}

	certResult, err := c.SignCSR(certs.EncodeBase64(csr), certConfig.Token.Token)
	if err != nil {
		return externalschema.CertificationResult{}, nil, err
	}

	return certResult, clientKey, nil
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


func GenerateRuntimeCertificate(t *testing.T, token *externalschema.Token, connectorClient *TokenSecuredClient, clientKey *rsa.PrivateKey) (externalschema.CertificationResult, externalschema.Configuration) {
	return generateCertificateForToken(t, connectorClient, token.Token, clientKey)
}

func GetConfiguration(t *testing.T, internalClient *InternalClient, connectorClient *TokenSecuredClient, appID string) externalschema.Configuration {
	token, err := internalClient.GenerateApplicationToken(appID)
	require.NoError(t, err)

	configuration, err := connectorClient.Configuration(token.Token)
	require.NoError(t, err)
	certs.AssertConfiguration(t, configuration)

	return configuration
}

func GenerateApplicationCertificate(t *testing.T, internalClient *InternalClient, connectorClient *TokenSecuredClient, appID string, clientKey *rsa.PrivateKey) (externalschema.CertificationResult, externalschema.Configuration) {
	token, err := internalClient.GenerateApplicationToken(appID)
	require.NoError(t, err)

	return generateCertificateForToken(t, connectorClient, token.Token, clientKey)
}

func generateCertificateForToken(t *testing.T, connectorClient *TokenSecuredClient, token string, clientKey *rsa.PrivateKey) (externalschema.CertificationResult, externalschema.Configuration) {
	configuration, err := connectorClient.Configuration(token)
	require.NoError(t, err)
	certs.AssertConfiguration(t, configuration)

	certToken := configuration.Token.Token
	subject := configuration.CertificateSigningRequestInfo.Subject

	csr:= certs.CreateCsr(t,subject, clientKey)
	require.NoError(t, err)

	result, err := connectorClient.SignCSR(certs.EncodeBase64(csr), certToken)
	require.NoError(t, err)

	return result, configuration
}

func ChangeCommonName(subject, commonName string) string {
	splitSubject := certs.ParseSubject(subject)

	splitSubject.CommonName = commonName

	return splitSubject.String()
}

func CreateCertDataHeader(subject, hash string) string {
	return fmt.Sprintf(`By=spiffe://cluster.local/ns/kyma-system/sa/default;Hash=%s;Subject="%s";URI=`, hash, subject)
}

func Cleanup(t *testing.T, configmapCleaner *pkg.ConfigmapCleaner, certificationResult externalschema.CertificationResult) {
	hash := certs.GetCertificateHash(t, certificationResult.ClientCertificate)
	err := configmapCleaner.CleanRevocationList(hash)
	assert.NoError(t, err)
}
