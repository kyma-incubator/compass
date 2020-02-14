package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Config struct {
	URL  string
	Auth struct {
		Username string
		Password string
	}
	ProvisionGCP bool
}

type Client struct {
	brokerConfig    Config
	clusterName     string
	instanceID      string
	globalAccountID string

	client *http.Client
	log    logrus.FieldLogger
}

func NewClient(config Config, globalAccountID, instanceID string, clientHttp http.Client, log logrus.FieldLogger) *Client {
	return &Client{
		brokerConfig:    config,
		instanceID:      instanceID,
		clusterName:     fmt.Sprintf("%s-%s", "e2e-provisioning", strings.ToLower(randstr.String(10))),
		globalAccountID: globalAccountID,
		client:          &clientHttp,
		log:             log,
	}
}

const (
	kymaClassID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
	gcpPlanID   = "ca6e5357-707f-4565-bbbd-b3ab732597c6"
	azurePlanID = "4deee563-e5ec-4731-b9b1-53b42d855f0c"

	instancesURL = "/v2/service_instances"
)

type inputContext struct {
	TenantID        string `json:"tenant_id"`
	SubAccountID    string `json:"subaccount_id"`
	GlobalAccountID string `json:"globalaccount_id"`
}

type provisionResponse struct {
	Operation string `json:"operation"`
}

type lastOperationResponse struct {
	State string `json:"state"`
}

type instanceDetailsResponse struct {
	DashboardURL string `json:"dashboard_url"`
}

type provisionParameters struct {
	Name       string   `json:"name"`
	Components []string `json:"components"`
}

func (c *Client) ProvisionRuntime() (string, error) {
	c.log.Infof("Provisioning Runtime [ID: %s, NAME: %s]", c.instanceID, c.clusterName)
	requestByte, err := c.prepareProvisionDetails()
	if err != nil {
		return "", errors.Wrap(err, "while preparing provision details")
	}
	c.log.Infof("Provisioning parameters: %v", string(requestByte))

	provisionURL := fmt.Sprintf("%s%s/%s", c.brokerConfig.URL, instancesURL, c.instanceID)
	response := provisionResponse{}
	err = wait.Poll(time.Second, time.Second*5, func() (bool, error) {
		err := c.executeRequest(http.MethodPut, provisionURL, bytes.NewReader(requestByte), &response)
		if err != nil {
			c.log.Warn(errors.Wrap(err, "while executing request").Error())
			return false, nil
		}
		if response.Operation == "" {
			c.log.Warn("Got empty operation ID")
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", errors.Wrap(err, "while waiting for successful provision call")
	}
	c.log.Infof("Successfully send provision request, got operation ID %s", response.Operation)

	return response.Operation, nil
}

func (c *Client) DeprovisionRuntime() (string, error) {
	deprovisionURL := fmt.Sprintf("%s%s/%s", c.brokerConfig.URL, instancesURL, c.instanceID)

	response := provisionResponse{}
	c.log.Infof("Deprovisioning Runtime [ID: %s, NAME: %s]", c.instanceID, c.clusterName)
	err := wait.Poll(time.Second, time.Second*5, func() (bool, error) {
		err := c.executeRequest(http.MethodDelete, deprovisionURL, nil, &response)
		if err != nil {
			c.log.Warn(errors.Wrap(err, "while executing request").Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", errors.Wrap(err, "while waiting for successful deprovision call")
	}
	c.log.Infof("Successfully send deprovision request, got operation ID %s", response.Operation)
	return response.Operation, nil
}

func (c *Client) AwaitOperationSucceeded(operationID string, timeout time.Duration) error {
	lastOperationURL := fmt.Sprintf("%s%s/%s/last_operation?operation=%s", c.brokerConfig.URL, instancesURL, c.instanceID, operationID)
	c.log.Infof("Waiting for operation at most %s", timeout.String())

	response := lastOperationResponse{}
	err := wait.Poll(5*time.Minute, timeout, func() (bool, error) {
		err := c.executeRequest(http.MethodGet, lastOperationURL, nil, &response)
		if err != nil {
			c.log.Warn(errors.Wrap(err, "while executing request").Error())
			return false, nil
		}
		switch domain.LastOperationState(response.State) {
		case domain.Failed:
			c.log.Info("Operation failed")
			return true, errors.New("operation failed")
		case domain.Succeeded:
			c.log.Infof("Operation succeeded!")
			return true, nil
		case domain.InProgress:
			return false, nil
		default:
			if response.State == "" {
				c.log.Infof("Got empty last operation response")
				return false, nil
			}
			return false, nil
		}
	})
	if err != nil {
		return errors.Wrap(err, "while waiting for succeeded last operation")
	}
	return nil
}

func (c *Client) FetchDashboardURL() (string, error) {
	instanceDetailsURL := fmt.Sprintf("%s%s/%s", c.brokerConfig.URL, instancesURL, c.instanceID)

	c.log.Info("Fetching the Runtime's dashboard URL")
	response := instanceDetailsResponse{}
	err := wait.Poll(time.Second, time.Second*5, func() (bool, error) {
		err := c.executeRequest(http.MethodGet, instanceDetailsURL, nil, &response)
		if err != nil {
			c.log.Warn(errors.Wrap(err, "while executing request").Error())
			return false, nil
		}
		if response.DashboardURL == "" {
			c.log.Warn("got empty dashboardURL")
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", errors.Wrap(err, "while waiting for dashboardURL")
	}
	c.log.Infof("Successfully fetched dashboard URL: %s", response.DashboardURL)

	return response.DashboardURL, nil
}

func (c *Client) prepareProvisionDetails() ([]byte, error) {
	parameters := provisionParameters{
		Name:       c.clusterName,
		Components: []string{""}, // fill with optional components
	}
	context := inputContext{
		TenantID:        "1eba80dd-8ff6-54ee-be4d-77944d17b10b",
		SubAccountID:    "39ba9a66-2c1a-4fe4-a28e-6e5db434084e",
		GlobalAccountID: c.globalAccountID,
	}
	rawParameters, err := json.Marshal(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling parameters body")
	}
	rawContext, err := json.Marshal(context)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling context body")
	}
	requestBody := domain.ProvisionDetails{
		ServiceID:        kymaClassID,
		PlanID:           azurePlanID,
		OrganizationGUID: uuid.New().String(),
		SpaceGUID:        uuid.New().String(),
		RawContext:       rawContext,
		RawParameters:    rawParameters,
		MaintenanceInfo: &domain.MaintenanceInfo{
			Version:     "0.1.0",
			Description: "Kyma environment broker e2e-provisioning test",
		},
	}
	if c.brokerConfig.ProvisionGCP {
		requestBody.PlanID = gcpPlanID
	}
	requestByte, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling request body")
	}
	return requestByte, nil
}

func (c *Client) executeRequest(method, url string, body io.Reader, responseBody interface{}) error {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err, "while creating request for provisioning")
	}
	request.SetBasicAuth(c.brokerConfig.Auth.Username, c.brokerConfig.Auth.Password)
	request.Header.Set("X-Broker-API-Version", "2.14")

	resp, err := c.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "while executing request URL: %s", url)
	}

	err = json.NewDecoder(resp.Body).Decode(responseBody)
	if err != nil {
		return errors.Wrapf(err, "while decoding body")
	}
	c.warnOnError(resp.Body.Close())
	return nil
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn(err.Error())
	}
}
