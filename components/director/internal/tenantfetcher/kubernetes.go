package tenantfetcher

import (
	"errors"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeClient interface {
	GetTenantFetcherConfigMapData() (string, error)
	UpdateTenantFetcherConfigMapData(timestamp string) error
}

type noopK8sClient struct {
	cfg KubeConfig
}

func NewNoopK8sClient() KubeClient {
	return &noopK8sClient{
		cfg: KubeConfig{
			ConfigMapNamespace:      "namespace",
			ConfigMapName:           "name",
			ConfigMapTimestampField: "timestampField",
		},
	}
}

func (k *noopK8sClient) GetTenantFetcherConfigMapData() (string, error) {
	return "1", nil
}

func (k *noopK8sClient) UpdateTenantFetcherConfigMapData(_ string) error {
	return nil
}

type KubeConfig struct {
	ConfigMapNamespace      string `envconfig:"default=compass-system,APP_CONFIGMAP_NAMESPACE"`
	ConfigMapName           string `envconfig:"default=tenant-fetcher-config,APP_CONFIGMAP_NAME"`
	ConfigMapTimestampField string `envconfig:"default=lastConsumedTenantTimestamp,APP_CONFIGMAP_TIMESTAMP_FIELD"`
}

type k8sClient struct {
	client *kubernetes.Clientset

	cfg KubeConfig
}

func NewK8sClient(configMapNamespace, configMapName, configMapTimestampField string) (KubeClient, error) {
	k8sClientSetConfig := kube.K8sConfig{}
	K8sClientSet, err := kube.NewK8sClientSet(k8sClientSetConfig.PollInteval, k8sClientSetConfig.PollTimeout, k8sClientSetConfig.Timeout)
	if err != nil {
		return nil, err
	}

	cfg := KubeConfig{
		ConfigMapNamespace:      configMapNamespace,
		ConfigMapName:           configMapName,
		ConfigMapTimestampField: configMapTimestampField,
	}

	return &k8sClient{
		client: K8sClientSet,
		cfg:    cfg,
	}, nil
}

func (k *k8sClient) GetTenantFetcherConfigMapData() (string, error) {
	configMap, err := k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Get(k.cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if timestamp, ok := configMap.Data[k.cfg.ConfigMapTimestampField]; ok {
		return timestamp, nil
	}
	return "", errors.New("failed to find timestamp property in configMap")
}

func (k *k8sClient) UpdateTenantFetcherConfigMapData(timestamp string) error {
	configMap, err := k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Get(k.cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	configMap.Data[k.cfg.ConfigMapTimestampField] = timestamp
	_, err = k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Update(configMap)
	return err
}
