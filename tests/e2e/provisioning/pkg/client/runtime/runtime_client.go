package runtime

import (
	"context"
	"fmt"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"time"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/director"
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
	ProvisionerURL       string `default:"http://compass-provisioner.compass-system.svc.cluster.local:3000/graphql"`
	KubeconfigSecretName string `default:"e2e-runtime-kubeconfig"`

	UUAInstanceName      string `default:"uaa-issuer"`
	UUAInstanceNamespace string `default:"kyma-system"`
}

// Client allows to fetch runtime's config and execute the logic against it
type Client struct {
	config         Config
	httpClient     http.Client
	directorClient *director.Client
	log            logrus.FieldLogger

	configSecretName string
	instanceID       string
	tenantID         string
}

func NewClient(config Config, tenantID, instanceID, configSecretName string, clientHttp http.Client, directorClient *director.Client, log logrus.FieldLogger) *Client {
	return &Client{
		tenantID:         tenantID,
		instanceID:       instanceID,
		configSecretName: configSecretName,
		config:           config,
		httpClient:       clientHttp,
		directorClient:   directorClient,
		log:              log,
	}
}

type runtimeStatusResponse struct {
	Result schema.RuntimeStatus `json:"result"`
}

func (c *Client) ExposeConfigAsSecret() error {
	response, err := c.fetchRuntimeConfig()
	if err != nil {
		return errors.Wrap(err, "while fetching runtime config")
	}
	config := *response.Result.RuntimeConfiguration.Kubeconfig

	// client for host cluster
	cli, err := c.newClient("")
	if err != nil {
		return errors.Wrap(err, "while setting client config")
	}
	err = cli.Create(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.configSecretName,
			Namespace: "compass-system",
		},
		Data: map[string][]byte{
			"config": []byte(config),
		},
	})
	if err != nil {
		return errors.Wrapf(err, "while creating a secret %s", c.configSecretName)
	}
	return nil
}

func (c *Client) EnsureUAAInstanceRemoved() error {
	cli, err := c.newRuntimeClient()
	if err != nil {
		return errors.Wrap(err, "while setting runtime config")
	}
	err = c.ensureInstanceRemoved(cli)
	if err != nil {
		return errors.Wrap(err, "while removing UUA instance")
	}
	c.log.Info("Successfully ensured UAA instance was removed")
	return nil
}

func (c *Client) newRuntimeClient() (client.Client, error) {
	response, err := c.fetchRuntimeConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtime config")
	}
	tmpFile, err := c.writeConfigToFile(response)
	if err != nil {
		return nil, errors.Wrap(err, "while writing runtime config")
	}
	defer c.removeFile(tmpFile)

	cli, err := c.newClient(tmpFile)
	if err != nil {
		return nil, errors.Wrap(err, "while setting client config")
	}
	return cli, nil
}

func (c *Client) fetchRuntimeConfig() (*runtimeStatusResponse, error) {
	runtimeID, err := c.directorClient.GetRuntimeID(c.tenantID, c.instanceID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime id from director for instance ID %s", c.instanceID)
	}
	// setup graphql client
	gCli := graphCli.NewClient(c.config.ProvisionerURL, graphCli.WithHTTPClient(&c.httpClient))

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
	return res, nil
}

func (c *Client) writeConfigToFile(response *runtimeStatusResponse) (string, error) {
	runtimeConfig := *response.Result.RuntimeConfiguration.Kubeconfig
	content := []byte(runtimeConfig)
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

func (c *Client) newClient(configPath string) (client.Client, error) {
	err := v1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, errors.Wrap(err, "while adding schema")
	}
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting kubeconfig under path %s", configPath)
	}
	cli, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (c *Client) ensureInstanceRemoved(cli client.Client) error {
	c.log.Infof("Waiting for %s instance to be removed", c.config.UUAInstanceName)
	return wait.Poll(time.Second, 3*time.Minute, func() (bool, error) {
		if err := cli.Delete(context.Background(), &v1beta1.ServiceInstance{
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
