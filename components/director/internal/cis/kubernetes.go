package cis

import (
	"context"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KubeConfig missing godoc
type KubeConfig struct {
	SecretNamespace    string `envconfig:"default=compass-system,APP_SECRET_NAMESPACE"`
	SecretName         string `envconfig:"default=cis-tokens,APP_SECRET_NAMESPACE"`
	ConfigMapNamespace string `envconfig:"default=compass-system,APP_CONFIGMAP_NAMESPACE"`
	ConfigMapName      string `envconfig:"default=cis-config,APP_CONFIGMAP_NAME"`
}

// KubeClient missing godoc
type KubeClient interface {
	GetRegionToken(region string) string
	GetRegionURL(region string) string
}

type kubernetesClient struct {
	client    *kubernetes.Clientset
	configmap map[string]string
	secret    map[string][]byte
	cfg       KubeConfig
}

// NewKubernetesClient missing godoc
func NewKubernetesClient(ctx context.Context, cfg KubeConfig, kubeClientConfig kube.Config) (KubeClient, error) {
	kubeClientSet, err := kube.NewKubernetesClientSet(ctx, kubeClientConfig.PollInterval, kubeClientConfig.PollTimeout, kubeClientConfig.Timeout)
	if err != nil {
		return nil, err
	}

	secret, err := kubeClientSet.CoreV1().Secrets(cfg.SecretNamespace).Get(ctx, cfg.SecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	configMap, err := kubeClientSet.CoreV1().ConfigMaps(cfg.ConfigMapNamespace).Get(ctx, cfg.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &kubernetesClient{
		client:    kubeClientSet,
		secret:    secret.Data,
		configmap: configMap.Data,
		cfg:       cfg,
	}, nil
}

func (k *kubernetesClient) GetRegionToken(region string) string {
	return string(k.secret[region])
}

func (k *kubernetesClient) GetRegionURL(region string) string {
	return k.configmap[region]
}
