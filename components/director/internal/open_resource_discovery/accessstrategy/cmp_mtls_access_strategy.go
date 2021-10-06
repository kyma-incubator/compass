package accessstrategy

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const cmpMTLSConfigPrefix = "ACCESS_STRATEGY_SAP_CMP_MTLS_V1"

type cmpMTLSAccessStrategyExecutor struct {
	lock sync.RWMutex

	client *http.Client

	config *cert.CertSvcConfig
}

// NewCMPmTLSAccessStrategyExecutor creates a new Executor for the CMP mTLS Access Strategy
func NewCMPmTLSAccessStrategyExecutor() *cmpMTLSAccessStrategyExecutor {
	return &cmpMTLSAccessStrategyExecutor{
		lock: sync.RWMutex{},
	}
}

// Execute performs the access strategy's specific execution logic
func (as *cmpMTLSAccessStrategyExecutor) Execute(ctx context.Context, baseClient *http.Client, documentURL string) (*http.Response, error) {
	if !as.isInitialized() {
		if err := as.initialize(ctx, baseClient); err != nil {
			return nil, errors.Wrap(err, "while initializing access strategy sap:cmp-mtls:v1")
		}
	}
	return as.client.Get(documentURL)
}

// SetConfig sets the Access Strategy config. This is used when reading from environment is not suitable in that context.
func (as *cmpMTLSAccessStrategyExecutor) SetConfig(config cert.CertSvcConfig) {
	as.config = &config
}

func (as *cmpMTLSAccessStrategyExecutor) isInitialized() bool {
	as.lock.RLock()
	defer as.lock.RUnlock()

	return as.config != nil && as.client != nil
}

func (as *cmpMTLSAccessStrategyExecutor) initialize(ctx context.Context, baseClient *http.Client) error {
	as.lock.Lock()
	defer as.lock.Unlock()

	if as.config == nil {
		cfg := cert.CertSvcConfig{}
		if err := envconfig.InitWithPrefix(&cfg, cmpMTLSConfigPrefix); err != nil {
			return err
		}
		as.config = &cfg
	}

	if as.client == nil {
		certSvcClient := cert.NewCertSvcClient(baseClient, *as.config)
		clientCert, err := certSvcClient.IssueClientCert(ctx)
		if err != nil {
			return err
		}

		tr := &http.Transport{}
		if baseClient.Transport != nil {
			tr = baseClient.Transport.(*http.Transport).Clone()
		}

		tr.TLSClientConfig = &tls.Config{
			Certificates: []tls.Certificate{*clientCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
		}

		as.client = &http.Client{
			Timeout:   baseClient.Timeout,
			Transport: tr,
		}
	}

	return nil
}
