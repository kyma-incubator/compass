package tenantfetcher

import (
	"errors"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockery --name=KubeClient --output=automock --outpkg=automock --case=underscore
type KubeClient interface {
	GetTenantFetcherConfigMapData() (string, error)
	UpdateTenantFetcherConfigMapData(timestamp string) error
}

type noopKubernetesClient struct {
	cfg KubeConfig
}

func NewNoopKubernetesClient() KubeClient {
	return &noopKubernetesClient{
		cfg: KubeConfig{
			ConfigMapNamespace:      "namespace",
			ConfigMapName:           "name",
			ConfigMapTimestampField: "timestampField",
		},
	}
}

func (k *noopKubernetesClient) GetTenantFetcherConfigMapData() (string, error) {
	return "1", nil
}

func (k *noopKubernetesClient) UpdateTenantFetcherConfigMapData(_ string) error {
	return nil
}

type KubeConfig struct {
	ConfigMapNamespace      string `envconfig:"default=compass-system,APP_CONFIGMAP_NAMESPACE"`
	ConfigMapName           string `envconfig:"default=tenant-fetcher-config,APP_CONFIGMAP_NAME"`
	ConfigMapTimestampField string `envconfig:"default=lastConsumedTenantTimestamp,APP_CONFIGMAP_TIMESTAMP_FIELD"`
}

type kubernetesClient struct {
	client *kubernetes.Clientset

	cfg KubeConfig
}

func NewKubernetesClient(configMapNamespace, configMapName, configMapTimestampField string) (KubeClient, error) {
	k8sClientSetConfig := kube.K8sConfig{}
	K8sClientSet, err := kube.NewK8sClientSet(k8sClientSetConfig.PollInterval, k8sClientSetConfig.PollTimeout, k8sClientSetConfig.Timeout)
	if err != nil {
		return nil, err
	}

	cfg := KubeConfig{
		ConfigMapNamespace:      configMapNamespace,
		ConfigMapName:           configMapName,
		ConfigMapTimestampField: configMapTimestampField,
	}

	return &kubernetesClient{
		client: K8sClientSet,
		cfg:    cfg,
	}, nil
}

func (k *kubernetesClient) GetTenantFetcherConfigMapData() (string, error) {
	configMap, err := k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Get(k.cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if timestamp, ok := configMap.Data[k.cfg.ConfigMapTimestampField]; ok {
		return timestamp, nil
	}
	return "", errors.New("failed to find timestamp property in configMap")
}

func (k *kubernetesClient) UpdateTenantFetcherConfigMapData(timestamp string) error {
	configMap, err := k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Get(k.cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	configMap.Data[k.cfg.ConfigMapTimestampField] = timestamp
	_, err = k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Update(configMap)
	return err
}
