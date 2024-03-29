package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/pkg/errors"
)

// ORDClient handles the communication with ord aggregator API
type ORDClient struct {
	cfg    OrdAggregatorClientConfig
	client *http.Client
}

// NewORDClient creates new ord aggregator client
func NewORDClient(cfg OrdAggregatorClientConfig) *ORDClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipSSLValidation,
		},
	}

	client := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), "Authorization")),
		Timeout:   cfg.ClientTimeout,
	}
	return &ORDClient{
		cfg:    cfg,
		client: client,
	}
}

// SetHTTPClient sets the underlying HTTP client
func (c *ORDClient) SetHTTPClient(client *http.Client) {
	c.client = client
}

// Aggregate call ord aggregator API
func (c *ORDClient) Aggregate(ctx context.Context, appID, appTemplateID string) error {
	ordData := ord.AggregationResources{
		ApplicationID:         appID,
		ApplicationTemplateID: appTemplateID,
	}
	marshalledOrdData, err := json.Marshal(ordData)
	if err != nil {
		return errors.Wrap(err, "while marshaling data for ord aggregator")
	}
	body := bytes.NewBuffer(marshalledOrdData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.OrdAggregatorAggregateAPI, body)
	if err != nil {
		return errors.Wrap(err, "while creating request to ord aggregator")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "while executing request to ord aggregator")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received unexpected status code %d while calling aggregate API with ApplicationID %q and ApplicationTemplateID %q", resp.StatusCode, appID, appTemplateID)
	}

	return nil
}
