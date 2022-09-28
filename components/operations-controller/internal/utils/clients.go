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
	Get() map[string]*tls.Certificate
}

// PrepareMtlsClient creates a secured MTLS http client with given certificate cache
func PrepareMtlsClient(cfg *httpbroker.Config, cache CertificateCache, secretName string) *http.Client {
	basicTransport := httpbroker.NewHTTPTransport(cfg)
	basicTransport.TLSClientConfig.GetClientCertificate = func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return cache.Get()[secretName], nil
	}
	httpTransport := httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(basicTransport))

	return &http.Client{
		Transport: httpTransport,
		Timeout:   cfg.Timeout,
	}
}

// PrepareHttpClient creates a http client with given http config
func PrepareHttpClient(cfg *httpbroker.Config) *http.Client {
	httpTransport := httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(httpbroker.NewHTTPTransport(cfg)))

	unsecuredClient := &http.Client{
		Transport: httpTransport,
		Timeout:   cfg.Timeout,
	}

	basicProvider := auth.NewBasicAuthorizationProvider()
	tokenProvider := auth.NewTokenAuthorizationProvider(unsecuredClient)

	securedTransport := httputil.NewSecuredTransport(httpTransport, basicProvider, tokenProvider)
	securedClient := &http.Client{
		Transport: securedTransport,
		Timeout:   cfg.Timeout,
	}

	return securedClient
}
