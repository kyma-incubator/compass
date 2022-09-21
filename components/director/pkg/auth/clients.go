package auth

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
)

// PrepareExtSvcMTLSClient creates an ext svc mtls secured http client with given timeout and cert cache
func PrepareExtSvcMTLSClient(timeout time.Duration, cache certloader.Cache) *http.Client {
	return PrepareMTLSClientWithSSLValidation(timeout, cache, false, 1)
}

// PrepareMTLSClient creates a mtls secured http client with given timeout and cert cache
func PrepareMTLSClient(timeout time.Duration, cache certloader.Cache) *http.Client {
	return PrepareMTLSClientWithSSLValidation(timeout, cache, false, 0)
}

// PrepareMTLSClientWithSSLValidation creates a mtls secured http client with given timeout, SSL validation and cert cache
func PrepareMTLSClientWithSSLValidation(timeout time.Duration, cache certloader.Cache, skipSSLValidation bool, idx int) *http.Client {
	basicTransport := http.DefaultTransport.(*http.Transport).Clone()
	basicTransport.TLSClientConfig.InsecureSkipVerify = skipSSLValidation
	basicTransport.TLSClientConfig.GetClientCertificate = func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return cache.Get()[idx], nil
	}
	roundTripper := httputil.NewHTTPTransportWrapper(basicTransport)
	httpTransport := httputil.NewCorrelationIDTransport(roundTripper)

	return &http.Client{
		Timeout:   timeout,
		Transport: httpTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// PrepareHTTPClient creates a http client with given timeout
func PrepareHTTPClient(timeout time.Duration) *http.Client {
	return PrepareHTTPClientWithSSLValidation(timeout, false)
}

// PrepareHTTPClientWithSSLValidation creates a secured http client with given timeout and SSL validation
func PrepareHTTPClientWithSSLValidation(timeout time.Duration, skipSSLValidation bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	roundTripper := httputil.NewHTTPTransportWrapper(transport)

	unsecuredClient := &http.Client{
		Timeout:   timeout,
		Transport: httputil.NewCorrelationIDTransport(roundTripper),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	basicProvider := NewBasicAuthorizationProvider()
	tokenProvider := NewTokenAuthorizationProvider(unsecuredClient)

	securedTransport := httputil.NewSecuredTransport(httputil.NewCorrelationIDTransport(roundTripper), basicProvider, tokenProvider)
	securedClient := &http.Client{
		Timeout:   timeout,
		Transport: securedTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return securedClient
}
