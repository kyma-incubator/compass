package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/pkg/errors"
)

// SystemFetcherClient handles the communication with system fetcher on demand API
type SystemFetcherClient struct {
	cfg    SystemFetcherSyncClientConfig
	client *http.Client
}

type aggregationResource struct {
	TenantIDs      []string `json:"tenantIDs"`
	SkipReschedule bool     `json:"skipReschedule"`
}

// NewSystemFetcherClient creates new system fetcher client
func NewSystemFetcherClient(cfg SystemFetcherSyncClientConfig) *SystemFetcherClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipSSLValidation,
		},
	}

	client := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), "Authorization")),
		Timeout:   cfg.ClientTimeout,
	}
	return &SystemFetcherClient{
		cfg:    cfg,
		client: client,
	}
}

// SetHTTPClient sets the underlying HTTP client
func (c *SystemFetcherClient) SetHTTPClient(client *http.Client) {
	c.client = client
}

// Sync call to system fetcher on demand API
func (c *SystemFetcherClient) Sync(ctx context.Context, tenantIDs []string, skipReschedule bool) error {
	log.C(ctx).Debugf("Call to sync systems API for Tenants %v started", tenantIDs)

	syncData := aggregationResource{
		TenantIDs:      tenantIDs,
		SkipReschedule: skipReschedule,
	}

	marshalledSyncData, err := json.Marshal(syncData)
	if err != nil {
		return errors.Wrap(err, "while marshaling data for system fetcher")
	}
	body := bytes.NewBuffer(marshalledSyncData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.SystemFetcherSyncAPI, body)
	if err != nil {
		return errors.Wrap(err, "while creating request to system fetcher sync API")
	}
	req.Header.Set("Content-Type", "application/json")

	log.C(ctx).Debugf("Executing remote request to sync systems API for Tenants %v", tenantIDs)
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "while executing request to system fetcher")
	}

	log.C(ctx).Debugf("Remote request to sync systems API for Tenants %v completed with status code %d", tenantIDs, resp.StatusCode)
	if resp.StatusCode != http.StatusAccepted {
		return errors.Errorf("received unexpected status code %d while calling sync API for Tenants %v", resp.StatusCode, tenantIDs)
	}

	log.C(ctx).Debugf("Call to sync systems API for Tenants %v completed", tenantIDs)
	return nil
}
