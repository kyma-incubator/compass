package accessstrategy

import (
	"crypto/tls"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
)

// HTTPRoundTripper missing godoc
type HTTPRoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
	Clone() HTTPRoundTripper
	GetTransport() *http.Transport
}

type cmpMTLSAccessStrategyExecutor struct {
	certCache certloader.Cache
}

// NewCMPmTLSAccessStrategyExecutor creates a new Executor for the CMP mTLS Access Strategy
func NewCMPmTLSAccessStrategyExecutor(certCache certloader.Cache) *cmpMTLSAccessStrategyExecutor {
	return &cmpMTLSAccessStrategyExecutor{
		certCache: certCache,
	}
}

// Execute performs the access strategy's specific execution logic
func (as *cmpMTLSAccessStrategyExecutor) Execute(baseClient *http.Client, documentURL string) (*http.Response, error) {
	clientCert := as.certCache.Get()
	if clientCert == nil {
		return nil, errors.New("did not find client certificate in the cache")
	}

	tr := &http.Transport{}
	if baseClient.Transport != nil {
		switch v := baseClient.Transport.(type) {
		case *http.Transport:
			tr = v.Clone()
		case HTTPRoundTripper:
			tr = v.GetTransport()
		default:
			return nil, errors.New("unsupported transport type")
		}
	}

	tr.TLSClientConfig.Certificates = []tls.Certificate{*clientCert}

	client := &http.Client{
		Timeout:   baseClient.Timeout,
		Transport: tr,
	}

	return client.Get(documentURL)
}
