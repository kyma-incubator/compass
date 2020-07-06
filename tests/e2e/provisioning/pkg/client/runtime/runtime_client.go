package runtime

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	schema "github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/control-plane/tests/e2e/provisioning/internal/director"
	graphCli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// tenantHeaderName is a header key name for request send by graphQL client
const tenantHeaderName = "tenant"

// Client allows to fetch runtime's config and execute the logic against it
type Client struct {
	httpClient     http.Client
	directorClient *director.Client
	log            logrus.FieldLogger

	provisionerURL string
	instanceID     string
	tenantID       string
}

func NewClient(provisionerURL, tenantID, instanceID string, clientHttp http.Client, directorClient *director.Client, log logrus.FieldLogger) *Client {
	return &Client{
		tenantID:       tenantID,
		instanceID:     instanceID,
		provisionerURL: provisionerURL,
		httpClient:     clientHttp,
		directorClient: directorClient,
		log:            log,
	}
}

type runtimeStatusResponse struct {
	Result schema.RuntimeStatus `json:"result"`
}

func (c *Client) FetchRuntimeConfig() (*string, error) {
	runtimeID, err := c.directorClient.GetRuntimeID(c.tenantID, c.instanceID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime id from director for instance ID %s", c.instanceID)
	}
	// setup graphql client
	gCli := graphCli.NewClient(c.provisionerURL, graphCli.WithHTTPClient(&c.httpClient))

	// create query
	q := fmt.Sprintf(`query {
	result: runtimeStatus(id: "%s") {
		runtimeConfiguration {
			kubeconfig
			}
		}
	}`, runtimeID)

	// prepare and run request
	req := graphCli.NewRequest(q)
	req.Header.Add(tenantHeaderName, c.tenantID)

	res := &runtimeStatusResponse{}
	err = gCli.Run(context.Background(), req, res)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime config")
	}
	if res.Result.RuntimeConfiguration != nil {
		return res.Result.RuntimeConfiguration.Kubeconfig, nil
	}
	return nil, errors.New("kubeconfig shouldn't be nil")
}

func (c *Client) writeConfigToFile(config string) (string, error) {
	content := []byte(config)
	runtimeConfigTmpFile, err := ioutil.TempFile("", "runtime.*.yaml")
	if err != nil {
		return "", errors.Wrap(err, "while creating runtime config temp file")
	}

	if _, err := runtimeConfigTmpFile.Write(content); err != nil {
		return "", errors.Wrap(err, "while writing runtime config temp file")
	}
	if err := runtimeConfigTmpFile.Close(); err != nil {
		return "", errors.Wrap(err, "while closing runtime config temp file")
	}

	return runtimeConfigTmpFile.Name(), nil
}

func (c *Client) removeFile(fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		c.log.Fatal(err)
	}
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn(err.Error())
	}
}
