package v1_client

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeout = time.Second * 10
)

type ConfigMaps interface {
	Get(name, namespace string) (*v1.ConfigMap, error)
	Create(configMap v1.ConfigMap) error
	Update(configMap v1.ConfigMap) error
	Delete(configMap v1.ConfigMap) error
}

type ConfigMapClient struct {
	client client.Client
	log    logrus.FieldLogger
}

func NewConfigMapClient(client client.Client, log logrus.FieldLogger) *ConfigMapClient {
	return &ConfigMapClient{client: client, log: log}
}

func (c *ConfigMapClient) Get(name, namespace string) (*v1.ConfigMap, error) {
	configMap := v1.ConfigMap{}
	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		err := c.client.Get(context.Background(), client.ObjectKey{Name: name, Namespace: namespace}, &configMap)
		if err != nil {
			if apiErrors.IsNotFound(err) {
				return false, err
			}
			c.log.Errorf("while creating config map: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		c.log.Errorf("while getting config map: %v", err)
		return nil, err
	}
	return &configMap, nil
}

func (c *ConfigMapClient) Create(configMap v1.ConfigMap) error {
	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		err := c.client.Create(context.Background(), &configMap)
		if err != nil {
			if apiErrors.IsAlreadyExists(err) {
				err = c.Update(configMap)
				if err != nil {
					return false, errors.Wrap(err, "while updating a config map")
				}
			}
			c.log.Errorf("while creating config map: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		c.log.Errorf("while creating secret: %v", err)
		return err
	}
	return nil
}

func (c *ConfigMapClient) Update(configMap v1.ConfigMap) error {
	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		err := c.client.Update(context.Background(), &configMap)
		if err != nil {
			c.log.Errorf("while creating config map: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return errors.Wrap(err, "while waiting for secret update")
	}
	return nil
}

func (c *ConfigMapClient) Delete(configMap v1.ConfigMap) error {
	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		err := c.client.Delete(context.Background(), &configMap)
		if err != nil {
			if apiErrors.IsNotFound(err) {
				c.log.Warn("config map not found")
				return true, nil
			}
			c.log.Errorf("while creating config map: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return errors.Wrap(err, "while waiting for secret update")
	}
	return nil
}
