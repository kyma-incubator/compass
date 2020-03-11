package auditlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"

	"github.com/pkg/errors"
)

const LogFormatDate = "2006-01-02T15:04:05.999Z"

const ConfigChangeURLPath = "/audit-log/v2/configuration-changes"
const SecurityEventURLPath = "/audit-log/v2/security-events"

//go:generate mockery -name=UUIDService -output=automock -outpkg=automock -case=underscore
type UUIDService interface {
	Generate() string
}

//go:generate mockery -name=TimeService -output=automock -outpkg=automock -case=underscore
type TimeService interface {
	Now() time.Time
}

type AuditlogConfig struct {
	User        string `envconfig:"APP_AUDITLOG_USER"`
	Password    string `envconfig:"APP_AUDITLOG_PASSWORD"`
	AuditLogURL string `envconfig:"APP_AUDITLOG_URL"`
	Tenant      string `envconfig:"APP_AUDITLOG_TENANT"`
}

type Client struct {
	cfg              AuditlogConfig
	uuidSvc          UUIDService
	timeSvc          TimeService
	http             http.Client
	configChangeURL  string
	securityEventURL string
}

func NewClient(cfg AuditlogConfig, uuidSvc UUIDService, tsvc TimeService) *Client {
	client := http.Client{
		Timeout: time.Second * 5,
	}
	configChangeURL := cfg.AuditLogURL + ConfigChangeURLPath
	securityEventURL := cfg.AuditLogURL + SecurityEventURLPath
	return &Client{configChangeURL: configChangeURL,
		securityEventURL: securityEventURL,
		http:             client,
		uuidSvc:          uuidSvc,
		cfg:              cfg,
		timeSvc:          tsvc}
}

//TODO: use basic auth, Currently JWT token is broken
func (c *Client) LogConfigurationChange(change model.ConfigurationChange) error {
	t := c.timeSvc.Now()
	logTime := t.Format(LogFormatDate)
	change.Time = logTime

	change.Tenant = c.cfg.Tenant
	change.UUID = c.uuidSvc.Generate()

	payload, err := json.Marshal(&change)
	if err != nil {
		return errors.Wrap(err, "while marshalling auditlog payload")
	}

	fmt.Printf("PAYLOAD TO CONFIGURATION CHANGE AUDITLOG:\n %s\n", payload)
	req, err := http.NewRequest("POST", c.configChangeURL, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}
	req.SetBasicAuth(c.cfg.User, c.cfg.Password)

	response, err := c.http.Do(req)
	if err != nil {
		return errors.Wrapf(err, "while sending auditlog to: %s", c.configChangeURL)
	}
	defer c.closeBody(response.Body)

	if response.StatusCode != http.StatusCreated {
		fmt.Printf("Got different status code: %d\n", response.StatusCode)
		output, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "while reading response from auditlog")
		}
		fmt.Print(string(output))
		return errors.Errorf("Write to auditlog failed with status code: %d", response.StatusCode)
	}

	return nil
}

func (c *Client) LogSecurityEvent(event model.SecurityEvent) error {
	t := c.timeSvc.Now()
	logTime := t.Format(LogFormatDate)
	event.Time = logTime

	event.Tenant = c.cfg.Tenant
	event.UUID = c.uuidSvc.Generate()

	payload, err := json.Marshal(&event)
	if err != nil {
		return errors.Wrap(err, "while marshalling auditlog payload")
	}

	req, err := http.NewRequest("POST", c.securityEventURL, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}

	return c.sendAuditLog(req)
}

func (c *Client) sendAuditLog(req *http.Request) error {
	req.SetBasicAuth(c.cfg.User, c.cfg.Password)
	response, err := c.http.Do(req)
	if err != nil {
		return errors.Wrapf(err, "while sending auditlog to: %s", c.configChangeURL)
	}
	defer c.closeBody(response.Body)

	if response.StatusCode != http.StatusCreated {
		fmt.Printf("Got different status code: %d\n", response.StatusCode)
		output, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "while reading response from auditlog")
		}
		fmt.Print(string(output))
		return errors.Errorf("Write to auditlog failed with status code: %d", response.StatusCode)
	}
	return nil
}

func (c *Client) closeBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.Printf("while closing body %+v\n", err)
	}
}
