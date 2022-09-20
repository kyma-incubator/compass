package resync

import (
	"context"
	"strconv"
	"time"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
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
	UseKubernetes string `envconfig:"USE_KUBERNETES"`

	ConfigMapNamespace            string `envconfig:"CONFIGMAP_NAMESPACE" default:"compass-system"`
	ConfigMapName                 string `envconfig:"LAST_EXECUTION_TIME_CONFIG_MAP_NAME" default:"tenant-fetcher-config"`
	ConfigMapTimestampField       string `envconfig:"CONFIGMAP_TIMESTAMP_FIELD" default:"lastConsumedTenantTimestamp"`
	ConfigMapResyncTimestampField string `envconfig:"CONFIGMAP_RESYNC_TIMESTAMP_FIELD" default:"lastFullResyncTimestamp"`

	PollInterval time.Duration `envconfig:"APP_KUBERNETES_POLL_INTERVAL" default:"2s"`
	PollTimeout  time.Duration `envconfig:"APP_KUBERNETES_POLL_TIMEOUT" default:"1m"`
	Timeout      time.Duration `envconfig:"APP_KUBERNETES_TIMEOUT" default:"2m"`
}

type kubernetesClient struct {
	client *kubernetes.Clientset

	cfg KubeConfig
}

func newKubernetesClient(ctx context.Context, cfg KubeConfig) (KubeClient, error) {
	kubeClientSet, err := kube.NewKubernetesClientSet(ctx, cfg.PollInterval, cfg.PollTimeout, cfg.Timeout)
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

func resyncTimestamps(ctx context.Context, client KubeClient, fullResyncInterval time.Duration) (*time.Time, string, string, error) {
	startTime := time.Now()

	lastConsumedTenantTimestamp, lastFullResyncTimestamp, err := client.GetTenantFetcherConfigMapData(ctx)
	if err != nil {
		return nil, "", "", err
	}

	shouldFullResync, err := shouldFullResync(lastFullResyncTimestamp, fullResyncInterval)
	if err != nil {
		return nil, "", "", err
	}

	if shouldFullResync {
		log.C(ctx).Infof("Last full resync was %s ago. Will perform a full resync.", fullResyncInterval)
		lastConsumedTenantTimestamp = "1"
		lastFullResyncTimestamp = convertTimeToUnixMilliSecondString(startTime)
	}
	return &startTime, lastConsumedTenantTimestamp, lastFullResyncTimestamp, nil
}

func shouldFullResync(lastFullResyncTimestamp string, fullResyncInterval time.Duration) (bool, error) {
	i, err := strconv.ParseInt(lastFullResyncTimestamp, 10, 64)
	if err != nil {
		return false, err
	}
	ts := time.Unix(i/1000, 0)
	return time.Now().After(ts.Add(fullResyncInterval)), nil
}

func convertTimeToUnixMilliSecondString(timestamp time.Time) string {
	return strconv.FormatInt(timestamp.UnixNano()/int64(time.Millisecond), 10)
}
