package accessstrategy

import (
	"bytes"
	"context"
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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

var globalTr *http.Transport
var mutex sync.Mutex

// Execute performs the access strategy's specific execution logic
func (as *cmpMTLSAccessStrategyExecutor) Execute(ctx context.Context, baseClient *http.Client, documentURL, tnt string, additionalHeaders *sync.Map) (*http.Response, error) {
	clientCerts := as.certCache.Get()
	if clientCerts == nil {
		return nil, errors.New("did not find client certificate in the cache")
	}
	mutex.Lock()
	if globalTr == nil {
		log.C(ctx).Infof("Missing global transport - will construct new one for request %q", documentURL)
		if baseClient.Transport != nil {
			switch v := baseClient.Transport.(type) {
			case *http.Transport:
				log.C(ctx).Infof("Missing global transport - will clone *http.Transport for request %q", documentURL)
				globalTr = v.Clone()
			case HTTPRoundTripper:
				log.C(ctx).Infof("Missing global transport - will clone HTTPRoundTripper for request %q", documentURL)
				globalTr = v.GetTransport().Clone()
			default:
				return nil, errors.New("unsupported transport type")
			}
		}
	} else {
		log.C(ctx).Infof("Will reuse global transport for request %q", documentURL)
	}

	if len(globalTr.TLSClientConfig.Certificates) != 0 {
		latestCert := clientCerts[as.externalClientCertSecretName].Certificate
		existingCert := globalTr.TLSClientConfig.Certificates[0].Certificate

		hasCertBeenRotated := false
		if len(latestCert) == len(existingCert) {
			for i := range latestCert {
				if !bytes.Equal(latestCert[i], existingCert[i]) {
					log.C(ctx).Infof("Client certificate has been rotated (bytes mismatch), will rotate it in the global transport for request: %s", documentURL)
					hasCertBeenRotated = true
					break
				}
			}
		} else {
			log.C(ctx).Infof("Client certificate has been rotated (length mismatch), will rotate it in the global transport for request: %s", documentURL)
			hasCertBeenRotated = true
		}

		if hasCertBeenRotated {
			globalTr.TLSClientConfig.Certificates = []tls.Certificate{*clientCerts[as.externalClientCertSecretName]}
		}
	}
	client := &http.Client{
		Timeout:   baseClient.Timeout,
		Transport: globalTr,
	}
	mutex.Unlock()

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

	resp, err := client.Do(req)
	//if err != nil || resp.StatusCode >= http.StatusBadRequest {
	//	if len(clientCerts) != 2 {
	//		return nil, errors.Errorf("There must be exactly 2 certificates in the cert cache. Actual number of certificates: %d", len(clientCerts))
	//	}
	//	log.C(ctx).Infof("Failed to execute request %q with initial mtls certificate. Will retry with backup certificate...", req.URL.String())
	//	tr.TLSClientConfig.Certificates = []tls.Certificate{*clientCerts[as.extSvcClientCertSecretName]}
	//	client.Transport = tr
	//	return client.Do(req)
	//}
	return resp, err
}
