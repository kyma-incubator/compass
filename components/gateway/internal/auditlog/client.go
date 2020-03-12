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

type AuditlogConfig struct {
	User                 string `envconfig:"APP_AUDITLOG_USER"`
	Password             string `envconfig:"APP_AUDITLOG_PASSWORD"`
	AuditLogURL          string `envconfig:"APP_AUDITLOG_URL"`
	Tenant               string `envconfig:"APP_AUDITLOG_TENANT"`
	AuditlogConfigPath   string `envconfig:"APP_AUDITLOG_CONFIG_PATH,default=/audit-log/v2/configuration-changes"`
	AuditlogSecurityPath string `envconfig:"APP_AUDITLOG_SECURITY_PATH,default=/audit-log/v2/security-events"`
}

type Client struct {
	cfg              AuditlogConfig
	uuidSvc          UUIDService
	timeSvc          TimeService
	http             http.Client
	configChangeURL  string
	securityEventURL string
}

func NewClient(cfg AuditlogConfig, uuidSvc UUIDService, tsvc TimeService) (*Client, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}
	configChangeURL, err := createURL(cfg.AuditLogURL, cfg.AuditlogConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "while creating auditlog config change url")
	}

	securityEventURL, err := createURL(cfg.AuditLogURL, cfg.AuditlogSecurityPath)
	if err != nil {
		return nil, errors.Wrap(err, "while creating auditlog security event url")

	}

	return &Client{configChangeURL: configChangeURL.String(),
		securityEventURL: securityEventURL.String(),
		http:             client,
		uuidSvc:          uuidSvc,
		cfg:              cfg,
		timeSvc:          tsvc}, nil
}

//TODO: use basic auth, Currently JWT token is broken
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
	req.SetBasicAuth(c.cfg.User, c.cfg.Password)
	response, err := c.http.Do(req)
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

	message.Tenant = c.cfg.Tenant
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
