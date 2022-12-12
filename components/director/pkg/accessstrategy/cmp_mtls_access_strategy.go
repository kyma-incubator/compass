package accessstrategy

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
)

// HTTPRoundTripper missing godoc
type HTTPRoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
	Clone() HTTPRoundTripper
	GetTransport() *http.Transport
}

const tenantHeader = "Tenant_Id"

type cmpMTLSAccessStrategyExecutor struct {
	certCache                    certloader.Cache
	tenantProviderFunc           func(ctx context.Context) (string, error)
	externalClientCertSecretName string
	extSvcClientCertSecretName   string
}

// NewCMPmTLSAccessStrategyExecutor creates a new Executor for the CMP mTLS Access Strategy
func NewCMPmTLSAccessStrategyExecutor(certCache certloader.Cache, tenantProviderFunc func(ctx context.Context) (string, error), externalClientCertSecretName, extSvcClientCertSecretName string) *cmpMTLSAccessStrategyExecutor {
	return &cmpMTLSAccessStrategyExecutor{
		certCache:                    certCache,
		tenantProviderFunc:           tenantProviderFunc,
		externalClientCertSecretName: externalClientCertSecretName,
		extSvcClientCertSecretName:   extSvcClientCertSecretName,
	}
}

// Execute performs the access strategy's specific execution logic
func (as *cmpMTLSAccessStrategyExecutor) Execute(ctx context.Context, baseClient *http.Client, documentURL, tnt string) (*http.Response, error) {
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

	client := &http.Client{
		Timeout:   baseClient.Timeout,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", documentURL, nil)
	if err != nil {
		return nil, err
	}

	if as.tenantProviderFunc != nil {
		localTenantID, err := as.tenantProviderFunc(ctx)
		if err != nil {
			return nil, err
		}

		req.Header.Set(tenantHeader, localTenantID)
	} else if len(tnt) > 0 {
		req.Header.Set(tenantHeader, tnt)
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= http.StatusBadRequest {
		if len(clientCerts) != 2 {
			return nil, errors.Errorf("There must be exactly 2 certificates in the cert cache. Actual number of certificates: %d", len(clientCerts))
		}
		log.C(ctx).Info("Failed to execute request with initial mtls certificate. Will retry with backup certificate...")
		tr.TLSClientConfig.Certificates = []tls.Certificate{*clientCerts[as.extSvcClientCertSecretName]}
		client.Transport = tr
		return client.Do(req)
	}
	return resp, err
}
