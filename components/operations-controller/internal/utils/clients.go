package utils

import (
	"crypto/tls"
	http2 "net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	http3 "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/http"
)

//CertificateCache is an interface which provides a certificate which is
//dynamically reloaded when its Get method is called
//go:generate mockery --name=CertificateCache --output=automock --outpkg=automock --case=underscore --disable-version-string
type CertificateCache interface {
	Get() *tls.Certificate
}

func PrepareMTLSClient(cfg *http.Config, cache CertificateCache) *http2.Client {
	basicTransport := http.NewHTTPTransport(cfg)
	basicTransport.TLSClientConfig.GetClientCertificate = func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return cache.Get(), nil
	}
	httpTransport := http3.NewCorrelationIDTransport(basicTransport)

	return &http2.Client{
		Transport: httpTransport,
		Timeout:   cfg.Timeout,
	}
}

func PrepareHttpClient(cfg *http.Config) (*http2.Client, error) {
	httpTransport := http3.NewCorrelationIDTransport(http.NewHTTPTransport(cfg))

	unsecuredClient := &http2.Client{
		Transport: httpTransport,
		Timeout:   cfg.Timeout,
	}

	basicProvider := auth.NewBasicAuthorizationProvider()
	tokenProvider := auth.NewTokenAuthorizationProvider(unsecuredClient)
	saTokenProvider := auth.NewServiceAccountTokenAuthorizationProvider()

	securedTransport := http3.NewSecuredTransport(httpTransport, basicProvider, tokenProvider, saTokenProvider)
	securedClient := &http2.Client{
		Transport: securedTransport,
		Timeout:   cfg.Timeout,
	}

	return securedClient, nil
}
