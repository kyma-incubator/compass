package k8s

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.ConfigMap, error)
	Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error)
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

func (c *ConfigmapCleaner) CleanRevocationList(ctx context.Context, hash string) error {
	configMap, err := c.configListManager.Get(ctx, c.configMapName, metav1.GetOptions{})
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
		_, err = c.configListManager.Update(ctx, updatedConfigMap, metav1.UpdateOptions{})
		return err
	})

	return err
}

func CreateClient(ctx context.Context) *kubernetes.Clientset {
	k8sClientSet, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing k8s client"))
	}
	return k8sClientSet
}
