package auditlog

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"

	"github.com/pkg/errors"
)

//go:generate mockery --name=HttpClient --output=automock --outpkg=automock --case=underscore
type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient       HttpClient
	configChangeURL  string
	securityEventURL string
}

func NewClient(cfg Config, httpClient HttpClient) (*Client, error) {
	configChangeURL, err := createURL(cfg.URL, cfg.ConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "while creating auditlog config change url")
	}

	securityEventURL, err := createURL(cfg.URL, cfg.SecurityPath)
	if err != nil {
		return nil, errors.Wrap(err, "while creating auditlog security event url")
	}

	return &Client{
		configChangeURL:  configChangeURL.String(),
		securityEventURL: securityEventURL.String(),
		httpClient:       httpClient,
	}, nil
}

func (c *Client) LogConfigurationChange(ctx context.Context, change model.ConfigurationChange) error {
	payload, err := json.Marshal(&change)
	if err != nil {
		return errors.Wrap(err, "while marshaling auditlog payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.configChangeURL, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}

	return c.sendAuditLog(ctx, req)
}

func (c *Client) LogSecurityEvent(ctx context.Context, event model.SecurityEvent) error {
	payload, err := json.Marshal(&event)
	if err != nil {
		return errors.Wrap(err, "while marshaling auditlog payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.securityEventURL, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}

	return c.sendAuditLog(ctx, req)
}

func (c *Client) sendAuditLog(ctx context.Context, req *http.Request) error {
	logger := log.C(ctx)
	responseStatusCode, err := c.sendAuditLogInternal(ctx, req)
	if responseStatusCode != http.StatusCreated || err != nil {
		logger.Infof("Will retry auditlog request.")
		responseStatusCode, err = c.sendAuditLogInternal(ctx, req)
		if responseStatusCode != http.StatusCreated || err != nil {
			return errors.Wrapf(err, "Write to auditlog failed with status code: %d", responseStatusCode)
		}
	}
	return nil
}

func (c *Client) sendAuditLogInternal(ctx context.Context, req *http.Request) (int, error) {
	logger := log.C(ctx)
	var err error = nil
	response, err := c.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrapf(err, "while sending auditlog to: %s", req.URL.String())
	}
	defer httpcommon.CloseBody(ctx, response.Body)

	if response.StatusCode != http.StatusCreated {
		var output []byte = nil
		logger.Infof("Got different status code: %d\n", response.StatusCode)
		output, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return response.StatusCode, errors.Wrap(err, "while reading response from auditlog")
		}
		logger.Infoln(string(output))
		err = errors.Errorf("Write to auditlog failed with status code: %d", response.StatusCode)
	}
	return response.StatusCode, err
}

func createURL(auditlogURL, urlPath string) (url.URL, error) {
	parsedURL, err := url.Parse(auditlogURL)
	if err != nil {
		return url.URL{}, errors.Wrap(err, "while creating auditlog URL")
	}
	parsedURL.Path = path.Join(parsedURL.Path, urlPath)
	return *parsedURL, nil
}
