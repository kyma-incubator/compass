package kyma

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Client struct {
	client http.Client
	log    logrus.FieldLogger
}

func NewClient(log logrus.FieldLogger) *Client {
	return &Client{
		client: http.Client{},
		log:    log,
	}
}

func (c *Client) CallDashboard(dashboardURL string) error {
	c.log.Info("Calling the dashboard URL")
	resp, err := c.client.Get(dashboardURL)
	if err != nil {
		return errors.Wrapf(err, "while calling dashboard '%s'", dashboardURL)
	}
	defer c.warnOnError(resp.Body.Close())

	fmt.Println(resp, "DUPA2")
	fmt.Println(resp.StatusCode, "DUPA")
	fmt.Println(resp.Body, "DUPA1")

	return nil
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn("couldn't close the response body")
	}
}
