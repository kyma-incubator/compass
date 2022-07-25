package auth

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
)

func PrepareMTLSClient(timeout time.Duration, cache certloader.Cache) *http.Client {
	return PrepareMTLSClientWithSSLValidation(timeout, cache, false)
}

func PrepareMTLSClientWithSSLValidation(timeout time.Duration, cache certloader.Cache, skipSSLValidation bool) *http.Client {
	basicTransport := http.DefaultTransport.(*http.Transport).Clone()
	basicTransport.TLSClientConfig.InsecureSkipVerify = skipSSLValidation
	basicTransport.TLSClientConfig.GetClientCertificate = func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return cache.Get(), nil
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

func PrepareHTTPClient(timeout time.Duration) *http.Client {
	return PrepareHTTPClientWithSSLValidation(timeout, false)
}

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
	saTokenProvider := NewServiceAccountTokenAuthorizationProvider()

	securedTransport := httputil.NewSecuredTransport(httputil.NewCorrelationIDTransport(roundTripper), basicProvider, tokenProvider, saTokenProvider)
	securedClient := &http.Client{
		Timeout:   timeout,
		Transport: securedTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return securedClient
}
