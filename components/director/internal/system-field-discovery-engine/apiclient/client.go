package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	systemfielddiscoveryengine "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/pkg/errors"
)

// SystemFieldDiscoveryEngineClient handles the communication with system field discovery engine API
type SystemFieldDiscoveryEngineClient struct {
	cfg    SystemFieldDiscoveryEngineClientConfig
	client *http.Client
}

// NewSystemFieldDiscoveryEngineClient creates new system field discovery engine client
func NewSystemFieldDiscoveryEngineClient(cfg SystemFieldDiscoveryEngineClientConfig) *SystemFieldDiscoveryEngineClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipSSLValidation,
		},
	}

	client := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), "Authorization")),
		Timeout:   cfg.ClientTimeout,
	}
	return &SystemFieldDiscoveryEngineClient{
		cfg:    cfg,
		client: client,
	}
}

// SetHTTPClient sets the underlying HTTP client
func (c *SystemFieldDiscoveryEngineClient) SetHTTPClient(client *http.Client) {
	c.client = client
}

// Discover calls system field discovery API
func (c *SystemFieldDiscoveryEngineClient) Discover(ctx context.Context, appID, tenantID, registry string) error {
	data := systemfielddiscoveryengine.SystemFieldDiscoveryResources{
		ApplicationID: appID,
		TenantID:      tenantID,
	}
	marshalledData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "while marshaling data for system field discovery")
	}

	body := bytes.NewBuffer(marshalledData)
	req := &http.Request{}
	if registry == "saas-registry" {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.SystemFieldDiscoveryEngineSaaSRegistryAPI, body)
	}
	if err != nil {
		return errors.Wrap(err, "while creating request to system field discovery")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "while executing request to system field discovery")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received unexpected status code %d while calling discover API with ApplicationID %q and Registry %q", resp.StatusCode, appID, registry)
	}

	return nil
}
