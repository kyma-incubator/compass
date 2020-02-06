package kyma

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client contains logic allowing to check if Kyma instance is accessible
type Client struct {
	client http.Client
	log    logrus.FieldLogger
}

func NewClient(clientHttp http.Client, log logrus.FieldLogger) *Client {
	clientHttp.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &Client{
		client: clientHttp,
		log:    log,
	}
}

func (c *Client) CallDashboardURL(dashboardURL string) error {
	targetURL := c.buildTargetURL(dashboardURL)

	c.log.Infof("Calling the dashboard URL: %s", targetURL)
	resp, err := c.client.Get(targetURL)
	if err != nil {
		return errors.Wrapf(err, "while calling dashboard '%s'", dashboardURL)
	}
	defer c.warnOnError(resp.Body.Close())

	if resp.StatusCode != http.StatusFound {
		return errors.Errorf("expected status code: %d, got %d", http.StatusFound, resp.StatusCode)
	}
	if location, err := resp.Location(); err != nil {
		return errors.Wrap(err, "while getting response location")
	} else if location.Path != "/auth/xsuaa" {
		return errors.Errorf("request was wrongly redirected: %s", location.String())
	}
	c.log.Info("Successful response from the dashboard URL")

	return nil
}

// Kyma console URL won't redirect us to the UUA logging page, to achieve that we must call dex with a set of parameters
// state and nonce params are faked
func (c *Client) buildTargetURL(dashboardURL string) string {
	domain := strings.Split(strings.Split(dashboardURL, "console.")[1], "/")[0]
	params := "&response_type=id_token%20token&scope=audience%3Aserver%3Aclient_id%3Akyma-client%20audience%3" +
		"Aserver%3Aclient_id%3Aconsole%20openid%20profile%20email%20groups&state=5a4f3d15&nonce=79bf9c29"

	return fmt.Sprintf("https://dex.%s/auth?client_id=console&redirect_uri=%s%s", domain, dashboardURL, params)
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn(err.Error())
	}
}
