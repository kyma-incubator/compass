package accessstrategy

import (
	"context"
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

const tenantHeader = "tenant"

type cmpMTLSAccessStrategyExecutor struct {
	certCache          certloader.Cache
	tenantProviderFunc func(ctx context.Context) (string, error)
}

// NewCMPmTLSAccessStrategyExecutor creates a new Executor for the CMP mTLS Access Strategy
func NewCMPmTLSAccessStrategyExecutor(certCache certloader.Cache, tenantProviderFunc func(ctx context.Context) (string, error)) *cmpMTLSAccessStrategyExecutor {
	return &cmpMTLSAccessStrategyExecutor{
		certCache:          certCache,
		tenantProviderFunc: tenantProviderFunc,
	}
}

// Execute performs the access strategy's specific execution logic
func (as *cmpMTLSAccessStrategyExecutor) Execute(ctx context.Context, baseClient *http.Client, documentURL, tnt string) (*http.Response, error) {
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
			tr = v.GetTransport().Clone()
		default:
			return nil, errors.New("unsupported transport type")
		}
	}

	tr.TLSClientConfig.Certificates = []tls.Certificate{*clientCert}

	client := &http.Client{
		Timeout:   baseClient.Timeout,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", documentURL, nil)
	if err != nil {
		return nil, err
	}

	if as.tenantProviderFunc != nil {
		tenant, err := as.tenantProviderFunc(ctx)
		if err != nil {
			return nil, err
		}

		req.Header.Set(tenantHeader, tenant)
	} else if len(tnt) > 0 {
		req.Header.Set(tenantHeader, tnt)
	}

	return client.Do(req)
}
