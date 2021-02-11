package tenantfetcher

import (
	"context"
	"errors"
	"strconv"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockery --name=KubeClient --output=automock --outpkg=automock --case=underscore
type KubeClient interface {
	GetTenantFetcherConfigMapData(ctx context.Context) (string, error)
	UpdateTenantFetcherConfigMapData(ctx context.Context, timestamp string) error
}

func NewKubernetesClient(ctx context.Context, cfg KubeConfig) (KubeClient, error) {
	shouldUseKubernetes, err := strconv.ParseBool(cfg.UseKubernetes)
	if err != nil {
		return nil, err
	}

	if !shouldUseKubernetes {
		return newNoopKubernetesClient(), nil
	}
	return newKubernetesClient(ctx, cfg)
}

type noopKubernetesClient struct {
	cfg KubeConfig
}

func newNoopKubernetesClient() KubeClient {
	return &noopKubernetesClient{
		cfg: KubeConfig{
			ConfigMapNamespace:      "namespace",
			ConfigMapName:           "name",
			ConfigMapTimestampField: "timestampField",
		},
	}
}

func (k *noopKubernetesClient) GetTenantFetcherConfigMapData(_ context.Context) (string, error) {
	return "1", nil
}

func (k *noopKubernetesClient) UpdateTenantFetcherConfigMapData(_ context.Context, _ string) error {
	return nil
}

type KubeConfig struct {
	UseKubernetes string `envconfig:"default=true,APP_USE_KUBERNETES"`

	ConfigMapNamespace      string `envconfig:"default=compass-system,APP_CONFIGMAP_NAMESPACE"`
	ConfigMapName           string `envconfig:"default=tenant-fetcher-config,APP_LAST_EXECUTION_TIME_CONFIG_MAP_NAME"`
	ConfigMapTimestampField string `envconfig:"default=lastConsumedTenantTimestamp,APP_CONFIGMAP_TIMESTAMP_FIELD"`
}

type kubernetesClient struct {
	client *kubernetes.Clientset

	cfg KubeConfig
}

func newKubernetesClient(ctx context.Context, cfg KubeConfig) (KubeClient, error) {
	kubeClientSetConfig := kube.Config{}
	kubeClientSet, err := kube.NewKubernetesClientSet(ctx, kubeClientSetConfig.PollInterval, kubeClientSetConfig.PollTimeout, kubeClientSetConfig.Timeout)
	if err != nil {
		return nil, err
	}

	return &kubernetesClient{
		client: kubeClientSet,
		cfg:    cfg,
	}, nil
}

func (k *kubernetesClient) GetTenantFetcherConfigMapData(ctx context.Context) (string, error) {
	configMap, err := k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Get(k.cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if timestamp, ok := configMap.Data[k.cfg.ConfigMapTimestampField]; ok {
		return timestamp, nil
	}
	return "", errors.New("failed to find timestamp property in configMap")
}

func (k *kubernetesClient) UpdateTenantFetcherConfigMapData(ctx context.Context, timestamp string) error {
	configMap, err := k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Get(k.cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	configMap.Data[k.cfg.ConfigMapTimestampField] = timestamp
	_, err = k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Update(configMap)
	return err
}
