package runtime

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	graphCli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// tenantHeaderName is a header key name for request send by graphQL client
const tenantHeaderName = "tenant"

type Config struct {
	ProvisionerURL string `default:"http://compass-provisioner.compass-system.svc.cluster.local:3000/graphql"`

	UUAInstanceName      string `default:"uaa-issuer"`
	UUAInstanceNamespace string `default:"kyma-system"`
}

// Client allows to fetch runtime's config and execute the logic against it
type Client struct {
	config Config
	log    logrus.FieldLogger

	client     client.Client
	httpClient http.Client

	runtimeID string
	tenantID  string
}

func NewClient(config Config, tenantID, runtimeID string, clientHttp http.Client, log logrus.FieldLogger) *Client {
	return &Client{
		tenantID:   tenantID,
		runtimeID:  runtimeID,
		config:     config,
		httpClient: clientHttp,
		log:        log,
	}
}

type runtimeStatusResponse struct {
	Result schema.RuntimeStatus `json:"result"`
}

func (c *Client) EnsureUAAInstanceRemoved() error {
	err := c.setRuntimeConfig() //TODO: get client and inject it later
	if err != nil {
		return errors.Wrap(err, "while setting runtime config")
	}
	err = c.ensureInstanceRemoved()
	if err != nil {
		return errors.Wrap(err, "while removing UUA instance")
	}
	return nil
}

func (c *Client) fetchRuntimeConfig() (*runtimeStatusResponse, error) {
	// setup graphql client
	gCli := graphCli.NewClient(c.config.ProvisionerURL, graphCli.WithHTTPClient(&c.httpClient))

	// create query
	q := fmt.Sprintf(`query {
	result: runtimeStatus(id: "%s") {
		runtimeConfiguration {
			kubeconfig
			}
		}
	}`, c.runtimeID)

	// prepare and run request
	req := graphCli.NewRequest(q)
	req.Header.Add(tenantHeaderName, c.tenantID)

	res := &runtimeStatusResponse{}
	err := gCli.Run(context.Background(), req, res)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime config")
	}
	return res, nil
}

func (c *Client) setRuntimeConfig() error {
	response, err := c.fetchRuntimeConfig()
	if err != nil {
		return errors.Wrap(err, "while fetching runtime config")
	}

	runtimeConfig := *response.Result.RuntimeConfiguration.Kubeconfig
	content := []byte(runtimeConfig)
	runtimeConfigTmpFile, err := ioutil.TempFile("", "runtime.*.yaml")
	if err != nil {
		return errors.Wrap(err, "while creating runtime config temp file")
	}
	defer func() {
		err = os.Remove(runtimeConfigTmpFile.Name())
		c.log.Fatal(err)
	}()

	if _, err := runtimeConfigTmpFile.Write(content); err != nil {
		return errors.Wrap(err, "while writing runtime config temp file")
	}
	if err := runtimeConfigTmpFile.Close(); err != nil {
		return errors.Wrap(err, "while closing runtime config temp file")
	}

	err = c.createClient(runtimeConfigTmpFile.Name())
	if err != nil {
		return errors.Wrap(err, "while setting client config")
	}
	return nil
}

func (c *Client) createClient(configPath string) error {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return errors.Wrapf(err, "while getting kubeconfig under path %s", configPath)
	}
	cli, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}
	c.client = cli
	return nil
}

func (c *Client) ensureInstanceRemoved() error {
	c.log.Infof("Waiting for %s instance to be removed", c.config.UUAInstanceName)
	return wait.Poll(time.Second, 3*time.Minute, func() (bool, error) {
		if err := c.client.Delete(context.Background(), &v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.config.UUAInstanceName,
				Namespace: c.config.UUAInstanceNamespace,
			},
		}); err != nil {
			if apiErrors.IsNotFound(err) {
				return true, nil
			}
			c.log.Warnf(errors.Wrap(err, "while removing instance").Error())
		}
		return false, nil
	})
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn(err.Error())
	}
}
