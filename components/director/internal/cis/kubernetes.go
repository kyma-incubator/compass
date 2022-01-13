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
	SecretName         string `envconfig:"default=cis-tokens,APP_SECRET_NAME"`
	ConfigMapNamespace string `envconfig:"default=compass-system,APP_CONFIGMAP_NAMESPACE"`
	ConfigMapName      string `envconfig:"default=cis-endpoints,APP_CONFIGMAP_NAME"`
	// for prod only
	ClientIDsNamespace     string `envconfig:"default=compass-system,APP_CLIENT_IDS_NAMESPACE"`
	ClientIDsMapName       string `envconfig:"default=client-ids,APP_CLIENT_IDS_MAP_NAME"`
	ClientSecretsNamespace string `envconfig:"default=compass-system,APP_CLIENT_SECRETS_NAMESPACE"`
	ClientSecretsMapName   string `envconfig:"default=client-secrets,APP_CLIENT_SECRETS_MAP_NAME"`
	TokenURLsNamespace     string `envconfig:"default=compass-system,APP_TOKEN_URLS_NAMESPACE"`
	TokenURLsMapName       string `envconfig:"default=token-urls,APP_TOKEN_URLS_MAP_NAME"`
}

// KubeClient missing godoc
type KubeClient interface {
	GetRegionToken(region string) string
	GetRegionURL(region string) string
	SetRegionToken(region string, newToken string)
	// for prod
	GetClientIDForRegion(region string) string
	GetClientSecretForRegion(region string) string
	GetTokenURLForRegion(region string) string
}

type kubernetesClient struct {
	client    *kubernetes.Clientset
	configmap map[string]string
	secret    map[string][]byte
	cfg       KubeConfig
	// for prod only
	clientIDs     map[string]string
	clientSecrets map[string]string
	tokenURLs     map[string]string
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

	// for prod
	clientIDs, err := kubeClientSet.CoreV1().ConfigMaps(cfg.ClientIDsNamespace).Get(ctx, cfg.ClientIDsMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	clientSecrets, err := kubeClientSet.CoreV1().ConfigMaps(cfg.ClientSecretsNamespace).Get(ctx, cfg.ClientSecretsMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	tokenURLs, err := kubeClientSet.CoreV1().ConfigMaps(cfg.TokenURLsNamespace).Get(ctx, cfg.TokenURLsMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &kubernetesClient{
		client:        kubeClientSet,
		secret:        secret.Data,
		configmap:     configMap.Data,
		cfg:           cfg,
		clientIDs:     clientIDs.Data,
		clientSecrets: clientSecrets.Data,
		tokenURLs:     tokenURLs.Data,
	}, nil
}

func (k *kubernetesClient) GetRegionToken(region string) string {
	return string(k.secret[region])
}

// used when we renew the token
func (k *kubernetesClient) SetRegionToken(region string, newToken string) {
	k.secret[region] = []byte(newToken)
}

func (k *kubernetesClient) GetRegionURL(region string) string {
	return k.configmap[region]
}

func (k *kubernetesClient) GetClientIDForRegion(region string) string {
	return k.clientIDs[region]
}

func (k *kubernetesClient) GetClientSecretForRegion(region string) string {
	return k.clientSecrets[region]
}

func (k *kubernetesClient) GetTokenURLForRegion(region string) string {
	return k.tokenURLs[region]
}
