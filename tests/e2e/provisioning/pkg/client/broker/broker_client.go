package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Config struct {
	URL  string
	Auth struct {
		Username string
		Password string
	}
	ProvisionGCP     bool
	ProvisionTimeout time.Duration
}

type Client struct {
	brokerConfig Config
	clusterName  string
	instanceID   string

	client http.Client
	log    logrus.FieldLogger
}

func NewClient(config Config, clusterName, instanceID string, log logrus.FieldLogger) *Client {
	return &Client{
		brokerConfig: config,
		clusterName:  clusterName,
		instanceID:   instanceID,
		client:       http.Client{},
		log:          log,
	}
}

const (
	kymaClassID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
	gcpPlanID   = "ca6e5357-707f-4565-bbbd-b3ab732597c6"
	azurePlanID = "4deee563-e5ec-4731-b9b1-53b42d855f0c"

	instancesURL = "/v2/service_instances"
)

func (c *Client) ProvisionRuntime() (string, error) {
	requestByte, err := c.prepareProvisionDetails(c.clusterName)
	if err != nil {
		return "", errors.Wrap(err, "while marshalling request body")
	}
	provisionURL := fmt.Sprintf("%s%s/%s", c.brokerConfig.URL, instancesURL, c.instanceID)

	response := provisionResponse{}
	c.log.Infof("Provisioning Runtime [ID: %s, NAME: %s]", c.instanceID, c.clusterName)
	err = wait.Poll(time.Second, time.Second*5, func() (bool, error) {
		err := c.executeRequest(http.MethodPut, provisionURL, bytes.NewReader(requestByte), response)
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

func (c *Client) DeprovisionRuntime() error {
	deprovisionURL := fmt.Sprintf("%s%s/%s", c.brokerConfig.URL, instancesURL, c.instanceID)

	response := provisionResponse{}
	c.log.Infof("Deprovisioning Runtime [ID: %s, NAME: %s]", c.instanceID, c.clusterName)
	err := wait.Poll(time.Second, time.Second*5, func() (bool, error) {
		err := c.executeRequest(http.MethodDelete, deprovisionURL, nil, response)
		if err != nil {
			c.log.Warn(errors.Wrap(err, "while executing request").Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return errors.Wrap(err, "while waiting for successful deprovision call")
	}
	c.log.Infof("Successfully send deprovision request, got operation ID %s", response.Operation)
	return nil
}

func (c *Client) AwaitProvisioningSucceeded(operationID string) error {
	lastOperationURL := fmt.Sprintf("%s%s/%s/last_operation?operation=%s", c.brokerConfig.URL, instancesURL, c.instanceID, operationID)

	response := lastOperationResponse{}
	err := wait.Poll(time.Second*15, c.brokerConfig.ProvisionTimeout, func() (bool, error) {
		err := c.executeRequest(http.MethodGet, lastOperationURL, nil, response)
		if err != nil {
			return false, errors.Wrap(err, "while executing request")
		}
		if response.State != string(gqlschema.OperationStateSucceeded) {
			c.log.Infof("Waiting 15s for provisioning succeeded...  state: %s", response.State)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return errors.Wrap(err, "while waiting for succeeded last operation")
	}
	c.log.Infof("Runtime Provisioning completed")
	return nil
}

func (c *Client) FetchDashboardURL() (string, error) {
	instanceDetailsURL := fmt.Sprintf("%s%s/%s", c.brokerConfig.URL, instancesURL, c.instanceID)

	c.log.Info("Fetching the Runtime's dashboard URL")
	response := instanceDetailsResponse{}
	err := wait.Poll(time.Second, time.Second*5, func() (bool, error) {
		err := c.executeRequest(http.MethodGet, instanceDetailsURL, nil, response)
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
	c.log.Info("Successfully fetched dashboard URL: %s", response.DashboardURL)

	return response.DashboardURL, nil
}

func (c *Client) prepareProvisionDetails(clusterName string) ([]byte, error) {
	parameters := provisionParameters{
		Name:       clusterName,
		Components: []string{""}, // fill with optional components
	}
	context := ersContext{
		TenantID:        "e2e-provisioning",
		SubAccountID:    "e2e-provisioning",
		GlobalAccountID: "e2e-provisioning",
	}
	rawParameters, err := json.Marshal(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling parameters body")
	}
	rawContext, err := json.Marshal(context)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling parameters body")
	}
	requestBody := domain.ProvisionDetails{
		ServiceID: kymaClassID,
		PlanID:    azurePlanID,
		MaintenanceInfo: &domain.MaintenanceInfo{
			Version:     "0.1.0",
			Description: "Kyma environment broker e2e-provisioning test",
		},
		RawParameters: rawParameters,
		RawContext:    rawContext,
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

	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return errors.Wrapf(err, "while decoding body")
	}
	c.warnOnError(resp.Body.Close())
	return nil
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn("couldn't close the response body")
	}
}
