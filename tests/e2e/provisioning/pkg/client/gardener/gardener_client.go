package gardener

import (
	"context"
	"net/http"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	SecretName string

	UUAInstanceName      string
	UUAInstanceNamespace string
}

type Client struct {
	config Config
	log    logrus.FieldLogger

	client     client.Client
	httpClient http.Client
	runtimeID  string
}

func NewClient(config Config, runtimeID string, log logrus.FieldLogger) (*Client, error) {
	return &Client{
		runtimeID:  runtimeID,
		config:     config,
		httpClient: http.Client{},
		client:     nil,
		log:        log,
	}, nil
}

func (c *Client) RuntimeTearDown() error {
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

// fetch directly from provisioner runtime status endpoint
func (c *Client) setRuntimeConfig() error {
	//runtimeConfigFile := "/configs/runtime.yaml"
	//err = ioutil.WriteFile(runtimeConfigFile, runtimeK8sConfig, 0644)
	//if err != nil {
	//	return errors.Wrap(err, "while creating runtime kubeconfig file")
	//}
	//err = c.setClientConfig(runtimeConfigFile)
	//if err != nil {
	//	return errors.Wrap(err, "while setting client config")
	//}

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
			c.log.Warnf(errors.Wrap(err, "while getting instance").Error())
		}
		return false, nil
	})
}

func (c *Client) warnOnError(err error) {
	if err != nil {
		c.log.Warn("couldn't close the response body")
	}
}
