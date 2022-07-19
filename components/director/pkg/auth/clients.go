package auth

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
)

func PrepareMTLSClient(timeout time.Duration, cache certloader.Cache) *http.Client {
	basicTransport := http.DefaultTransport.(*http.Transport).Clone()
	basicTransport.TLSClientConfig.GetClientCertificate = func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return cache.Get(), nil
	}
	httpTransport := httputil.NewCorrelationIDTransport(basicTransport)

	return &http.Client{
		Timeout:   timeout,
		Transport: httpTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func PrepareHTTPClient(timeout time.Duration) *http.Client {
	unsecuredClient := &http.Client{
		Timeout:   timeout,
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	basicProvider := NewBasicAuthorizationProvider()
	tokenProvider := NewTokenAuthorizationProvider(unsecuredClient)
	saTokenProvider := NewServiceAccountTokenAuthorizationProvider()

	securedTransport := httputil.NewSecuredTransport(httputil.NewCorrelationIDTransport(http.DefaultTransport), basicProvider, tokenProvider, saTokenProvider)
	securedClient := &http.Client{
		Timeout:   timeout,
		Transport: securedTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return securedClient
}
