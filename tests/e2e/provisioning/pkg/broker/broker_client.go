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
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	gcpPlanID   = "ca6e5357-707f-4565-bbbd-b3ab732597c6"
	azurePlanID = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
)

type Config struct {
	URL            string
	ServiceClassID string
	Auth           struct {
		Username string
		Password string
	}
}

type Client struct {
	brokerConfig     Config
	provisionGCP     bool
	provisionTimeout time.Duration

	client http.Client
	log    logrus.FieldLogger
}

func NewClient(config Config, provisionGCP bool, provisionTimeout time.Duration, log logrus.FieldLogger) *Client {
	return &Client{
		client:           http.Client{},
		brokerConfig:     config,
		provisionGCP:     provisionGCP,
		provisionTimeout: provisionTimeout,
		log:              log,
	}
}

const (
	instancesURL = "/v2/service_instances"
)

func (c *Client) ProvisionRuntime() (string, string, error) {
	instanceUUID, err := uuid.NewRandom()
	if err != nil {
		return "", "", errors.Wrap(err, "while generating instanceID")
	}
	instanceID := instanceUUID.String()
	clusterName := fmt.Sprintf("%s-%s", "e2e-provisioning", strings.ToLower(randstr.String(10)))

	requestByte, err := c.prepareProvisionDetails(clusterName)
	if err != nil {
		return "", "", errors.Wrap(err, "while marshalling request body")
	}

	c.log.Infof("Provisioning Runtime [ID: %s, NAME: %s]", instanceID, clusterName)
	result := &http.Response{}
	err = wait.Poll(time.Second, time.Minute, func() (bool, error) {
		request, err := c.prepareRequest(http.MethodPut, fmt.Sprintf("%s%s/%s", c.brokerConfig.URL, instancesURL, instanceID), bytes.NewReader(requestByte))
		if err != nil {
			return false, errors.Wrap(err, "while creating request for provisioning")
		}
		resp, err := c.client.Do(request)
		if err != nil {
			c.log.Warn(errors.Wrap(err, "while sending a request").Error())
			return false, nil
		}
		result = resp
		defer c.warnOnError(resp.Body.Close())

		if resp.StatusCode != http.StatusAccepted {
			c.log.Warn(errors.Errorf("expected status code is %d, got %d", http.StatusAccepted, resp.StatusCode).Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", "", errors.Wrap(err, "while waiting for successful provision call to KEB")
	}
	response := ProvisionResponse{}
	err = json.NewDecoder(result.Body).Decode(&response)
	if err != nil {
		return "", "", errors.Wrapf(err, "while decoding body")
	}
	err = result.Body.Close()
	if err != nil {
		return "", "", errors.New("cannot close body")
	}

	c.log.Infof("Successfully send provision request, got operation ID %s", response.Operation)

	return instanceID, response.Operation, nil
}

func (c *Client) AwaitProvisioningSucceeded(instanceID, operationID string) error {
	err := wait.Poll(time.Second*15, c.provisionTimeout, func() (bool, error) {
		request, err := c.prepareRequest(http.MethodGet, fmt.Sprintf("%s%s/%s/last_operation?operation=%s", c.brokerConfig.URL, instancesURL, instanceID, operationID), nil)
		if err != nil {
			return false, errors.Wrap(err, "while creating request for provisioning")
		}
		resp, err := c.client.Do(request)
		if err != nil {
			c.log.Warn(errors.Wrapf(err, "while getting last operation of instance %s", instanceID))
			return false, nil
		}
		defer c.warnOnError(resp.Body.Close())

		response := struct {
			State string `json:"state"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return false, errors.Wrapf(err, "while decoding body")
		}
		if response.State != string(gqlschema.OperationStateSucceeded) {
			c.log.Infof("Waiting 15s for operation to success...  state: %s", response.State)
			return false, nil
		}
		return true, nil
	})
	return err
}

func (c *Client) GetInstanceDetails(instanceID string) (string, error) {
	request, err := c.prepareRequest(http.MethodGet, fmt.Sprintf("%s%s/%s", c.brokerConfig.URL, instancesURL, instanceID), nil)
	if err != nil {
		return "", errors.Wrap(err, "while creating request for provisioning")
	}
	resp, err := c.client.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "while getting instance %s", instanceID)
	}
	defer c.warnOnError(resp.Body.Close())

	response := struct {
		DashboardURL string `json:"dashboard_url"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", errors.Wrapf(err, "while decoding body")
	}
	if response.DashboardURL == "" {
		return "", errors.New("got empty dashboardURL")
	}
	c.log.Info("Successfully fetched dashboard URL: %s", response.DashboardURL)

	return response.DashboardURL, nil
}

func (c *Client) prepareProvisionDetails(clusterName string) ([]byte, error) {
	parameters := struct {
		Name       string   `json:"name"`
		Components []string `json:"components"`
	}{
		Name:       clusterName,
		Components: []string{""}, // fill with optional components
	}
	ersContext := ERSContext{
		TenantID:        "e2e-provisioning",
		SubAccountID:    "e2e-provisioning",
		GlobalAccountID: "e2e-provisioning",
	}
	rawContext, err := json.Marshal(ersContext)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling parameters body")
	}
	rawParams, err := json.Marshal(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling parameters body")
	}
	requestBody := domain.ProvisionDetails{
		ServiceID:        c.brokerConfig.ServiceClassID,
		OrganizationGUID: "e2e-provisioning",
		SpaceGUID:        "e2e-provisioning",
		MaintenanceInfo: &domain.MaintenanceInfo{
			Version:     "0.1.0",
			Description: "Kyma environment broker e2e-provisioning test",
		},
		RawParameters: rawParams,
		RawContext:    rawContext,
	}
	requestBody.PlanID = azurePlanID
	if c.provisionGCP {
		requestBody.PlanID = gcpPlanID
	}
	requestByte, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling request body")
	}
	return requestByte, nil
}

func (c *Client) prepareRequest(method, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "while creating request for provisioning")
	}
	request.SetBasicAuth(c.brokerConfig.Auth.Username, c.brokerConfig.Auth.Password)
	request.Header.Set("X-Broker-API-Version", "2.14")

	return request, nil
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn("couldn't close the response body")
	}
}
