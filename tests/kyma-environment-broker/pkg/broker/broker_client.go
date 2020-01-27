package broker

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"time"
)

type Client struct {
	BrokerURL           string
	ProvisioningTimeout time.Duration

	client http.Client
	log    logrus.FieldLogger
}

func NewClient(brokerURL string, log logrus.Logger) *Client {
	return &Client{
		BrokerURL: brokerURL,
		client:    http.Client{},
		log:       &log,
	}
}

const (
	instancesURL = "/v2/service_instances"
)

func (c *Client) ProvisionRuntime() (string, error) {
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s%s", c.BrokerURL, instancesURL), nil)
	if err != nil {
		return "", errors.Wrap(err, "while creating request for provisioning")
	}
	resp, err := c.client.Do(request)
	if err != nil {
		return "", errors.Wrap(err, "while sending a request")
	}
	defer c.warnOnError(resp.Body.Close())

	fmt.Println(resp.StatusCode, "DUPA")
	fmt.Println(resp.Body, "DUPA1")

	return "", nil
}

func (c *Client) AwaitProvisioningSucceeded(instanceID string) error {
	err := wait.Poll(time.Second*5, c.ProvisioningTimeout, func() (bool, error) {
		resp, err := c.client.Get(fmt.Sprintf("%s%s/%s%s", c.BrokerURL, instancesURL, instanceID, "/last_operation"))
		if err != nil {
			c.log.Warn(errors.Wrapf(err, "while getting last operation of instance %s", instanceID))
			return false, nil
		}
		defer c.warnOnError(resp.Body.Close())

		fmt.Println(resp.StatusCode, "DUPA")
		fmt.Println(resp.Body, "DUPA1")

		return true, nil
	})
	return err
}

func (c *Client) GetInstanceDetails(instanceID string) (string, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s%s/%s", c.BrokerURL, instancesURL, instanceID))
	if err != nil {
		return "", errors.Wrapf(err, "while getting instance %s", instanceID)
	}
	defer c.warnOnError(resp.Body.Close())

	fmt.Println(resp.StatusCode, "DUPA")
	fmt.Println(resp.Body, "DUPA1")

	return "", nil
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn("couldn't close the response body")
	}
}
