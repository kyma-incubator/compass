package gardener

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"k8s.io/client-go/rest"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
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

	client           client.Client
	workingNamespace string
	clusterName      string
}

func NewClient(config Config, clusterName, workingNamespace string, log logrus.FieldLogger) (*Client, error) {
	k8sCfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "cannot find Service Account in pod to build in-cluster kube config")
	}
	cli, err := client.New(k8sCfg, client.Options{})
	if err != nil {
		return nil, errors.Wrap(err, "while creating a new client")
	}
	return &Client{
		workingNamespace: workingNamespace,
		clusterName:      clusterName,
		config:           config,
		client:           cli,
		log:              log,
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

func (c *Client) setRuntimeConfig() error {
	// setting local cluster config
	err := c.setClientConfig("")
	if err != nil {
		return errors.Wrap(err, "while setting client config")
	}

	secret := v1.Secret{}
	if err = c.client.Get(context.Background(), client.ObjectKey{
		Namespace: c.workingNamespace,
		Name:      c.config.SecretName,
	}, &secret); err != nil {
		return errors.Wrapf(err, "while getting secret %s/%s", c.workingNamespace, c.config.SecretName)
	}
	gardenerK8sConfig, ok := secret.Data["credentials"]
	if !ok {
		return errors.New("credentials not exist inside secret")
	}
	gardenerConfigFile := "/configs/gardener.yaml"
	err = ioutil.WriteFile(gardenerConfigFile, gardenerK8sConfig, 0644)
	if err != nil {
		return errors.Wrap(err, "while creating gardener kubeconfig file")
	}
	// setting gardener cluster config
	err = c.setClientConfig(gardenerConfigFile)
	if err != nil {
		return errors.Wrap(err, "while setting client config")
	}

	runtimeSecretName := fmt.Sprintf("%s.kubeconfig", c.clusterName)
	if err = c.client.Get(context.Background(), client.ObjectKey{
		Namespace: "default",
		Name:      runtimeSecretName,
	}, &secret); err != nil {
		return errors.Wrapf(err, "while getting secret %s/%s", c.workingNamespace, runtimeSecretName)
	}
	runtimeK8sConfig, ok := secret.Data["kubeconfig"]
	if !ok {
		return errors.New("kubeconfig not exist inside secret")
	}
	runtimeConfigFile := "/configs/runtime.yaml"
	err = ioutil.WriteFile(runtimeConfigFile, runtimeK8sConfig, 0644)
	if err != nil {
		return errors.Wrap(err, "while creating runtime kubeconfig file")
	}
	// setting runtime cluster config
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
