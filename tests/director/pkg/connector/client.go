package connector

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pkg/errors"

	gcli "github.com/machinebox/graphql"
)

const TokenHeader = "Connector-Token"

type client struct {
	gqlClient *gcli.Client
}

func NewClient(url string) *client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}

	gqlClient := gcli.NewClient(url, gcli.WithHTTPClient(httpClient))

	return &client{
		gqlClient: gqlClient,
	}
}

func (c *client) GetConfiguration(token string) (Configuration, error) {
	query := configurationQuery()
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var resp ConfigurationResponse

	err := c.gqlClient.Run(context.Background(), req, &resp)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "failed to get the configuration from the connector")
	}
	return resp.Result, nil
}

func (c *client) GenerateAndSignCert(certConfig Configuration) (CertificationResult, *rsa.PrivateKey, error) {
	clientKey, err := GenerateKey()
	if err != nil {
		return CertificationResult{}, nil, err
	}

	csr, err := CreateCsr(certConfig.CertificateSigningRequestInfo.Subject, clientKey)
	if err != nil {
		return CertificationResult{}, nil, err
	}

	certResult, err := c.SignCert(csr, certConfig.Token.Token)
	if err != nil {
		return CertificationResult{}, nil, err
	}

	return certResult, clientKey, nil
}

func (c *client) SignCert(csr, token string) (CertificationResult, error) {
	query := signCSRMutation(csr)
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var resp CertificationResponse

	err := c.gqlClient.Run(context.Background(), req, &resp)
	if err != nil {
		return CertificationResult{}, errors.Wrap(err, "failed to sign the certificate by the connector")
	}
	return resp.Result, nil
}

type ConfigurationResponse struct {
	Result Configuration `json:"result"`
}

type CertificationResponse struct {
	Result CertificationResult `json:"result"`
}

// copy of connector types, to get rid of dependency to connector
type CertificateSigningRequestInfo struct {
	Subject      string `json:"subject"`
	KeyAlgorithm string `json:"keyAlgorithm"`
}

type CertificationResult struct {
	CertificateChain  string `json:"certificateChain"`
	CaCertificate     string `json:"caCertificate"`
	ClientCertificate string `json:"clientCertificate"`
}

type Configuration struct {
	Token                         *Token                         `json:"token"`
	CertificateSigningRequestInfo *CertificateSigningRequestInfo `json:"certificateSigningRequestInfo"`
	ManagementPlaneInfo           *ManagementPlaneInfo           `json:"managementPlaneInfo"`
}

type ManagementPlaneInfo struct {
	DirectorURL                    *string `json:"directorURL"`
	CertificateSecuredConnectorURL *string `json:"certificateSecuredConnectorURL"`
}

type Token struct {
	Token string `json:"token"`
}
