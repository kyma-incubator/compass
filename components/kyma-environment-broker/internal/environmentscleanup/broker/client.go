package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/apimachinery/pkg/util/wait"

	log "github.com/sirupsen/logrus"
)

const (
	kymaClassID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
	gcpPlanID   = "ca6e5357-707f-4565-bbbd-b3ab732597c6"
	azurePlanID = "4deee563-e5ec-4731-b9b1-53b42d855f0c"

	instancesURL    = "/oauth/v2/service_instances"
	deprovisionTmpl = "%s%s/%s?service_id=%s&plan_id=%s"
)

type Config struct {
	URL          string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
}

type Client struct {
	brokerConfig Config
	httpClient   *http.Client
}

func NewClient(ctx context.Context, config Config) *Client {
	cfg := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     config.TokenURL,
		Scopes:       []string{config.Scope},
	}
	httpClientOAuth := cfg.Client(ctx)
	httpClientOAuth.Timeout = 30 * time.Second

	return &Client{
		brokerConfig: config,
		httpClient:   httpClientOAuth,
	}
}

type DeprovisionDetails struct {
	InstanceID       string
	CloudProfileName string
}

type deprovisionResponse struct {
	Operation string `json:"operation"`
}

// Deprovision requests Runtime deprovisioning in KEB with given details
func (c *Client) Deprovision(details DeprovisionDetails) (string, error) {
	deprovisionURL := c.formatDeprovisionUrl(details)

	response := deprovisionResponse{}
	log.Infof("Requesting deprovisioning of the environment with instance id: %q", details.InstanceID)
	err := wait.Poll(time.Second, time.Second*5, func() (bool, error) {
		err := c.executeRequest(http.MethodDelete, deprovisionURL, http.StatusAccepted, nil, &response)
		if err != nil {
			log.Warn(errors.Wrap(err, "while executing request").Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", errors.Wrap(err, "while waiting for successful deprovision call")
	}
	return response.Operation, nil
}

func (c *Client) formatDeprovisionUrl(details DeprovisionDetails) string {
	switch details.CloudProfileName {
	case "az":
		return fmt.Sprintf(deprovisionTmpl, c.brokerConfig.URL, instancesURL, details.InstanceID, kymaClassID, azurePlanID)
	case "gcp":
		return fmt.Sprintf(deprovisionTmpl, c.brokerConfig.URL, instancesURL, details.InstanceID, kymaClassID, gcpPlanID)
	default:
		return ""
	}
}

func (c *Client) executeRequest(method, url string, expectedStatus int, body io.Reader, responseBody interface{}) error {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err, "while creating request for provisioning")
	}
	request.Header.Set("X-Broker-API-Version", "2.14")

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "while executing request URL: %s", url)
	}
	defer c.warnOnError(resp.Body.Close)
	if resp.StatusCode != expectedStatus {
		return errors.Errorf("got unexpected status code while calling Kyma Environment Broker: want: %d, got: %d", expectedStatus, resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(responseBody)
	if err != nil {
		return errors.Wrapf(err, "while decoding body")
	}

	return nil
}

func (c *Client) warnOnError(do func() error) {
	if err := do(); err != nil {
		log.Warn(err.Error())
	}
}

// setHttpClient auxiliary method of testing to get rid of oAuth client wrapper
func (c *Client) setHttpClient(httpClient *http.Client) {
	c.httpClient = httpClient
}
