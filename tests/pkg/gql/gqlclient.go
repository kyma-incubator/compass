package gql

import (
	"crypto"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	gcli "github.com/machinebox/graphql"
)

func NewAuthorizedGraphQLClientWithCustomURL(bearerToken, url string) *gcli.Client {
	authorizedClient := NewAuthorizedHTTPClient(bearerToken)
	return gcli.NewClient(url, gcli.WithHTTPClient(authorizedClient))
}

func NewCertAuthorizedGraphQLClientWithCustomURL(url string, key crypto.PrivateKey, rawCertChain [][]byte, skipSSLValidation bool) *gcli.Client {
	certAuthorizedClient := NewCertAuthorizedHTTPClient(key, rawCertChain, skipSSLValidation)
	return gcli.NewClient(url, gcli.WithHTTPClient(certAuthorizedClient))
}

type authenticatedTransport struct {
	http.Transport
	token string
}

func NewAuthorizedHTTPClient(bearerToken string) *http.Client {
	transport := &authenticatedTransport{
		Transport: http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		token: bearerToken,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	return t.Transport.RoundTrip(req)
}

func NewCertAuthorizedHTTPClient(key crypto.PrivateKey, rawCertChain [][]byte, skipSSLValidation bool) *http.Client {
	tlsCert := tls.Certificate{
		Certificate: rawCertChain,
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: skipSSLValidation,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Second * 120,
	}

	return httpClient
}
