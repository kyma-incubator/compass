package auditlog

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"

	"github.com/pkg/errors"
)

const LogFormatDate = "2006-01-02T15:04:05.999Z"

//go:generate mockery -name=UUIDService -output=automock -outpkg=automock -case=underscore
type UUIDService interface {
	Generate() string
}

//go:generate mockery -name=TimeService -output=automock -outpkg=automock -case=underscore
type TimeService interface {
	Now() time.Time
}

//go:generate mockery -name=HttpClient -output=automock -outpkg=automock -case=underscore
type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type Client struct {
	uuidSvc          UUIDService
	timeSvc          TimeService
	httpClient       HttpClient
	configChangeURL  string
	securityEventURL string
	tenant           *string
}

func NewClient(cfg Config, httpClient HttpClient, uuidSvc UUIDService, tsvc TimeService, tenant *string) (*Client, error) {
	configChangeURL, err := createURL(cfg.URL, cfg.ConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "while creating auditlog config change url")
	}

	securityEventURL, err := createURL(cfg.URL, cfg.SecurityPath)
	if err != nil {
		return nil, errors.Wrap(err, "while creating auditlog security event url")
	}

	return &Client{configChangeURL: configChangeURL.String(),
		securityEventURL: securityEventURL.String(),
		tenant:           tenant,
		uuidSvc:          uuidSvc,
		timeSvc:          tsvc,
		httpClient:       httpClient}, nil
}

func (c *Client) LogConfigurationChange(change model.ConfigurationChange) error {
	c.fillMessage(&change.AuditlogMetadata)
	payload, err := json.Marshal(&change)
	if err != nil {
		return errors.Wrap(err, "while marshaling auditlog payload")
	}

	req, err := http.NewRequest("POST", c.configChangeURL, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}

	return c.sendAuditLog(req)
}

func (c *Client) LogSecurityEvent(event model.SecurityEvent) error {
	c.fillMessage(&event.AuditlogMetadata)
	payload, err := json.Marshal(&event)
	if err != nil {
		return errors.Wrap(err, "while marshaling auditlog payload")
	}

	req, err := http.NewRequest("POST", c.securityEventURL, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}

	return c.sendAuditLog(req)
}

func (c *Client) sendAuditLog(req *http.Request) error {
	response, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "while sending auditlog to: %s", req.URL.String())
	}
	defer httpcommon.CloseBody(response.Body)

	if response.StatusCode != http.StatusCreated {
		log.Printf("Got different status code: %d\n", response.StatusCode)
		output, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "while reading response from auditlog")
		}
		log.Println(string(output))
		return errors.Errorf("Write to auditlog failed with status code: %d", response.StatusCode)
	}
	return nil
}

func (c *Client) fillMessage(message *model.AuditlogMetadata) {
	t := c.timeSvc.Now()
	logTime := t.Format(LogFormatDate)
	message.Time = logTime

	if c.tenant != nil {
		message.Tenant = *c.tenant
	}
	message.UUID = c.uuidSvc.Generate()
}

func createURL(auditlogURL, urlPath string) (url.URL, error) {
	parsedURL, err := url.Parse(auditlogURL)
	if err != nil {
		return url.URL{}, errors.Wrap(err, "while creating auditlog URL")
	}
	parsedURL.Path = path.Join(parsedURL.Path, urlPath)
	return *parsedURL, nil
}
