package runtime

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	graphCli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// accountIDKey is a header key name for request send by graphQL client
const accountIDKey = "tenant"

type Config struct {
	ProvisionerURL string

	UUAInstanceName      string
	UUAInstanceNamespace string
}

type Client struct {
	config Config
	log    logrus.FieldLogger

	client     client.Client
	httpClient http.Client

	runtimeID       string
	globalAccountID string
}

func NewClient(config Config, runtimeID, globalAccountID string, log logrus.FieldLogger) (*Client, error) {
	return &Client{
		globalAccountID: globalAccountID,
		runtimeID:       runtimeID,
		config:          config,
		httpClient:      http.Client{},
		client:          nil,
		log:             log,
	}, nil
}

func (c *Client) TearDown() error {
	err := c.setRuntimeConfig()
	if err != nil {
		return errors.Wrap(err, "while setting runtime config")
	}
	err = c.ensureInstanceRemoved()
	if err != nil {
		return errors.Wrap(err, "while removing UUA instance")
	}
	return nil
}

type graphQLResponseWrapper struct {
	Result schema.RuntimeStatus `json:"result"`
}

func (c *Client) setRuntimeConfig() error {
	// setup graphql client
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	gCli := graphCli.NewClient(c.config.ProvisionerURL, graphCli.WithHTTPClient(httpClient))

	// setup logging in graphql client
	logger := logrus.WithField("Client", "GraphQL")
	gCli.Log = func(s string) {
		logger.Info(s)
	}

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
	req.Header.Add(accountIDKey, c.globalAccountID)

	res := &graphQLResponseWrapper{}
	err := gCli.Run(context.Background(), req, res)
	if err != nil {
		return errors.Wrapf(err, "while getting runtime config")
	}

	runtimeConfig := *res.Result.RuntimeConfiguration.Kubeconfig
	gCli.Log(runtimeConfig)

	runtimeConfigFile := "/configs/runtime.yaml"
	err = ioutil.WriteFile(runtimeConfigFile, []byte(runtimeConfig), 0644)
	if err != nil {
		return errors.Wrap(err, "while creating runtime kubeconfig file")
	}
	err = c.setClientConfig(runtimeConfigFile)
	if err != nil {
		return errors.Wrap(err, "while setting client config")
	}
	return nil
}

func (c *Client) setClientConfig(configPath string) error {
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
		c.log.Warn("couldn't close the response body")
	}
}
