package accessstrategy

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
)

// HTTPRoundTripper missing godoc
type HTTPRoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
	Clone() HTTPRoundTripper
	GetTransport() *http.Transport
}

const tenantHeader = "Tenant_Id"

type cmpMTLSAccessStrategyExecutor struct {
	certCache                    credloader.CertCache
	tenantProviderFunc           func(ctx context.Context) (string, error)
	externalClientCertSecretName string
	extSvcClientCertSecretName   string
}

// NewCMPmTLSAccessStrategyExecutor creates a new Executor for the CMP mTLS Access Strategy
func NewCMPmTLSAccessStrategyExecutor(certCache credloader.CertCache, tenantProviderFunc func(ctx context.Context) (string, error), externalClientCertSecretName, extSvcClientCertSecretName string) *cmpMTLSAccessStrategyExecutor {
	return &cmpMTLSAccessStrategyExecutor{
		certCache:                    certCache,
		tenantProviderFunc:           tenantProviderFunc,
		externalClientCertSecretName: externalClientCertSecretName,
		extSvcClientCertSecretName:   extSvcClientCertSecretName,
	}
}

// Execute performs the access strategy's specific execution logic
func (as *cmpMTLSAccessStrategyExecutor) Execute(ctx context.Context, baseClient *http.Client, documentURL, tnt string, additionalHeaders *sync.Map) (*http.Response, error) {
	clientCerts := as.certCache.Get()
	if clientCerts == nil {
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

	tr.TLSClientConfig.Certificates = []tls.Certificate{*clientCerts[as.externalClientCertSecretName]}
	baseClient.Transport = tr

	req, err := http.NewRequest("GET", documentURL, nil)
	if err != nil {
		return nil, err
	}

	if additionalHeaders != nil {
		additionalHeaders.Range(func(key, value any) bool {
			req.Header.Set(str.CastOrEmpty(key), str.CastOrEmpty(value))
			return true
		})
	}

	// if it's not request to global registry && the webhook is associated with app template use the local tenant id as header
	if as.tenantProviderFunc != nil && len(tnt) > 0 {
		localTenantID, err := as.tenantProviderFunc(ctx)
		if err != nil {
			return nil, err
		}

		req.Header.Set(tenantHeader, localTenantID)
	} else {
		req.Header.Set(tenantHeader, tnt)
	}

	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			log.C(ctx).Infof("Connection reused: %+v\n", connInfo)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	resp, err := baseClient.Do(req)
	if err != nil || resp.StatusCode >= http.StatusBadRequest {
		if len(clientCerts) != 2 {
			return nil, errors.Errorf("There must be exactly 2 certificates in the cert cache. Actual number of certificates: %d", len(clientCerts))
		}
		log.C(ctx).Infof("Failed to execute request %q with initial mtls certificate. Will retry with backup certificate...", req.URL.String())
		tr.TLSClientConfig.Certificates = []tls.Certificate{*clientCerts[as.extSvcClientCertSecretName]}
		baseClient.Transport = tr
		return baseClient.Do(req)
	}
	return resp, err
}
