package auditlog

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/avast/retry-go"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"

	"github.com/pkg/errors"
)

//go:generate mockery --name=HttpClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient       HttpClient
	configChangeURL  string
	securityEventURL string
}

const (
	retryAttempts          = 2
	retryDelayMilliseconds = 100
)

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

	return c.sendAuditLogWithRetry(ctx, c.configChangeURL, payload)
}

func (c *Client) LogSecurityEvent(ctx context.Context, event model.SecurityEvent) error {
	payload, err := json.Marshal(&event)
	if err != nil {
		return errors.Wrap(err, "while marshaling auditlog payload")
	}

	return c.sendAuditLogWithRetry(ctx, c.securityEventURL, payload)
}

func (c *Client) sendAuditLogWithRetry(ctx context.Context, url string, payload []byte) error {
	logger := log.C(ctx)
	err := retry.Do(func() error {
		buf := bytes.NewBuffer(payload)
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
		if err != nil {
			return errors.Wrap(err, "while creating request")
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return errors.Wrapf(err, "while sending auditlog to: %s", request.URL.String())
		}

		defer httpcommon.CloseBody(ctx, response.Body)

		if response.StatusCode != http.StatusCreated {
			logger.Infof("Got different status code: %d\n", response.StatusCode)
			output, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return errors.Wrap(err, "while reading response from auditlog")
			}
			logger.Infoln(string(output))
			return errors.Errorf("Write to auditlog failed with status code: %d", response.StatusCode)
		}
		return nil
	}, retry.Attempts(retryAttempts), retry.Delay(retryDelayMilliseconds*time.Millisecond))

	if err != nil {
		return err
	}
	return nil
}

func createURL(auditlogURL, urlPath string) (url.URL, error) {
	parsedURL, err := url.Parse(auditlogURL)
	if err != nil {
		return url.URL{}, errors.Wrap(err, "while creating auditlog URL")
	}
	parsedURL.Path = path.Join(parsedURL.Path, urlPath)
	return *parsedURL, nil
}
