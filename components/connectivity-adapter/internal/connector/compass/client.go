package compass

import (
	"context"
	"crypto/tls"
	"fmt"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
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
	SignCSR(csr string, headers map[string]string) (schema.CertificationResult, error)
}

func NewClient(graphqlEndpoint string, enableLogging, insecureConfigFetch bool) (Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureConfigFetch,
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

func (c client) Configuration(headers map[string]string) (schema.Configuration, error) {
	query := `query{
 		result: configuration()
        {
 			 token { token }
			 certificateSigningRequestInfo { subject keyAlgorithm }
			 managementPlaneInfo { directorURL certificateSecuredConnectorURL }
		}	
     }`

	var response ConfigurationResponse

	err := c.execute(headers, query, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}

	return response.Result, nil
}

func (c client) SignCSR(csr string, headers map[string]string) (schema.CertificationResult, error) {
	query := fmt.Sprintf(`mutation {
	result: signCertificateSigningRequest(csr: "%s")
  	{
	 	certificateChain
		caCertificate
		clientCertificate
	}
    }`, csr)

	var response CertificateResponse
	err := c.execute(headers, query, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to sign csr")
	}

	return response.Result, nil
}

type ConfigurationResponse struct {
	Result schema.Configuration `json:"result"`
}

type CertificateResponse struct {
	Result schema.CertificationResult `json:"result"`
}

func applyHeaders(req *graphql.Request, headers map[string]string) {
	for h, val := range headers {
		req.Header.Set(h, val)
	}
}

func (c *client) execute(headers map[string]string, query string, res interface{}) error {

	req := graphql.NewRequest(query)
	applyHeaders(req, headers)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.clearLogs()
	err := c.gqlClient.Run(ctx, req, res)
	if err != nil {
		for _, l := range c.logs {
			if l != "" {
				log.Println(l)
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
