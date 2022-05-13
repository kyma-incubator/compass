package tenantfetcher

import (
	"context"
	"strconv"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KubeClient missing godoc
//go:generate mockery --name=KubeClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type KubeClient interface {
	GetTenantFetcherConfigMapData(ctx context.Context) (string, string, error)
	UpdateTenantFetcherConfigMapData(ctx context.Context, lastRunTimestamp, lastResyncTimestamp string) error
}

// NewKubernetesClient missing godoc
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

// GetTenantFetcherConfigMapData missing godoc
func (k *noopKubernetesClient) GetTenantFetcherConfigMapData(_ context.Context) (string, string, error) {
	return "1", "1", nil
}

// UpdateTenantFetcherConfigMapData missing godoc
func (k *noopKubernetesClient) UpdateTenantFetcherConfigMapData(_ context.Context, _, _ string) error {
	return nil
}

// KubeConfig missing godoc
type KubeConfig struct {
	UseKubernetes string `envconfig:"default=true,APP_USE_KUBERNETES"`

	ConfigMapNamespace            string `envconfig:"default=compass-system,APP_CONFIGMAP_NAMESPACE"`
	ConfigMapName                 string `envconfig:"default=tenant-fetcher-config,APP_LAST_EXECUTION_TIME_CONFIG_MAP_NAME"`
	ConfigMapTimestampField       string `envconfig:"default=lastConsumedTenantTimestamp,APP_CONFIGMAP_TIMESTAMP_FIELD"`
	ConfigMapResyncTimestampField string `envconfig:"default=lastFullResyncTimestamp,APP_CONFIGMAP_RESYNC_TIMESTAMP_FIELD"`
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

// GetTenantFetcherConfigMapData missing godoc
func (k *kubernetesClient) GetTenantFetcherConfigMapData(ctx context.Context) (string, string, error) {
	configMap, err := k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Get(ctx, k.cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	lastRunTimestamp, ok := configMap.Data[k.cfg.ConfigMapTimestampField]
	if !ok {
		lastRunTimestamp = "1"
	}

	lastResyncTimestamp, ok := configMap.Data[k.cfg.ConfigMapResyncTimestampField]
	if !ok {
		lastResyncTimestamp = "1"
	}
	return lastRunTimestamp, lastResyncTimestamp, nil
}

// UpdateTenantFetcherConfigMapData missing godoc
func (k *kubernetesClient) UpdateTenantFetcherConfigMapData(ctx context.Context, lastRunTimestamp, lastResyncTimestamp string) error {
	configMap, err := k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Get(ctx, k.cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	configMap.Data[k.cfg.ConfigMapTimestampField] = lastRunTimestamp
	configMap.Data[k.cfg.ConfigMapResyncTimestampField] = lastResyncTimestamp
	_, err = k.client.CoreV1().ConfigMaps(k.cfg.ConfigMapNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
	return err
}
