package testkit

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

type Manager interface {
	Get(name string, options metav1.GetOptions) (*v1.ConfigMap, error)
	Update(configmap *v1.ConfigMap) (*v1.ConfigMap, error)
}

type ConfigmapCleaner struct {
	configListManager Manager
	configMapName     string
}

func NewConfigMapCleaner(configListManager Manager, configMapName string) *ConfigmapCleaner {
	return &ConfigmapCleaner{
		configListManager: configListManager,
		configMapName:     configMapName,
	}
}

func (c *ConfigmapCleaner) CleanRevocationList(hash string) error {
	configMap, err := c.configListManager.Get(c.configMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	revokedCerts := configMap.Data
	if revokedCerts == nil {
		return nil
	}
	delete(revokedCerts, hash)

	updatedConfigMap := configMap
	updatedConfigMap.Data = revokedCerts

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err = c.configListManager.Update(updatedConfigMap)
		return err
	})

	return err
}
