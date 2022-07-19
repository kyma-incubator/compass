package utils

import (
	"crypto/tls"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	httpbroker "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
)

//CertificateCache is an interface which provides a certificate which is
//dynamically reloaded when its Get method is called
//go:generate mockery --name=CertificateCache --output=automock --outpkg=automock --case=underscore --disable-version-string
type CertificateCache interface {
	Get() *tls.Certificate
}

func PrepareMTLSClient(cfg *httpbroker.Config, cache CertificateCache) *http.Client {
	basicTransport := httpbroker.NewHTTPTransport(cfg)
	basicTransport.TLSClientConfig.GetClientCertificate = func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return cache.Get(), nil
	}
	httpTransport := httputil.NewCorrelationIDTransport(basicTransport)

	return &http.Client{
		Transport: httpTransport,
		Timeout:   cfg.Timeout,
	}
}

func PrepareHttpClient(cfg *httpbroker.Config) (*http.Client, error) {
	httpTransport := httputil.NewCorrelationIDTransport(httpbroker.NewHTTPTransport(cfg))

	unsecuredClient := &http.Client{
		Transport: httpTransport,
		Timeout:   cfg.Timeout,
	}

	basicProvider := auth.NewBasicAuthorizationProvider()
	tokenProvider := auth.NewTokenAuthorizationProvider(unsecuredClient)
	saTokenProvider := auth.NewServiceAccountTokenAuthorizationProvider()

	securedTransport := httputil.NewSecuredTransport(httpTransport, basicProvider, tokenProvider, saTokenProvider)
	securedClient := &http.Client{
		Transport: securedTransport,
		Timeout:   cfg.Timeout,
	}

	return securedClient, nil
}
