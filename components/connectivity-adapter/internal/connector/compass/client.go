package compass

import (
	"context"
	"crypto/tls"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"
)

const (
	timeout = 30 * time.Second
)

type client struct {
	gqlClient *graphql.Client
	logs      []string
	logging   bool
}

type Client interface {
	Configuration(headers map[string]string) (schema.Configuration, error)
}

func New(certificate *tls.Certificate, graphqlEndpoint string, enableLogging, insecureConfigFetch bool) (Client, error) {
	var certificates []tls.Certificate
	if certificate != nil {
		certificates = []tls.Certificate{*certificate}
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureConfigFetch,
				Certificates:       certificates,
			},
		},
	}

	gqlClient := graphql.NewClient(graphqlEndpoint, graphql.WithHTTPClient(httpClient))

	client := &client{
		gqlClient: gqlClient,
		logging:   enableLogging,
		logs:      []string{},
	}

	client.gqlClient.Log = client.addLog

	return client, nil
}

func NewConnectorClient(graphqlEndpoint string) Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	gqlClient := graphql.NewClient(graphqlEndpoint, graphql.WithHTTPClient(httpClient))

	return &client{
		gqlClient: gqlClient,
	}
}

func (c client) Configuration(headers map[string]string) (schema.Configuration, error) {
	query := `query{
 		configuration
        {
 			 token { token }
			 certificateSigningRequestInfo { subject keyAlgorithm }
			 managementPlaneInfo { directorURL certificateSecuredConnectorURL }
		}	
     }`

	req := graphql.NewRequest(query)

	applyHeaders(req, headers)

	var response ConfigurationResponse

	err := c.execute(req, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}

	log.Printf("SUUUCCCESSS: %v", response)

	return response.Result.Configuration, nil
}

type Data struct {
	Configuration schema.Configuration `json:"configuration"`
}

type ConfigurationResponse struct {
	Result Data `json:"data"`
}

func applyHeaders(req *graphql.Request, headers map[string]string) {
	for h, val := range headers {
		req.Header.Set(h, val)
	}
}

func (c *client) execute(req *graphql.Request, res interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.clearLogs()
	err := c.gqlClient.Run(ctx, req, res)
	if err != nil {
		for _, l := range c.logs {
			if l != "" {
				logrus.Info(l)
			}
		}
	}
	return err
}

func (c *client) addLog(log string) {
	if !c.logging {
		return
	}

	c.logs = append(c.logs, log)
}

func (c *client) clearLogs() {
	c.logs = []string{}
}
