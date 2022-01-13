package cis

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

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
	ClientCredentialsName      string `envconfig:"default=cis-client-creds,APP_CIS_CLIENT_CREDS_MAP_NAME"`
	ClientCredentialsNamespace string `envconfig:"default=compass-system"`
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

type oAuthDetails struct {
	ClientID     string // jsonTag
	ClientSecret string // jsonTag
	TokenURL     string // jsonTag
}

type kubernetesClient struct {
	client    *kubernetes.Clientset
	configmap map[string]string
	secret    map[string][]byte
	cfg       KubeConfig
	// for prod only
	clientCreds map[string]oAuthDetails
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
	clientCreds := make(map[string]oAuthDetails)
	clientIdsMap, err := kubeClientSet.CoreV1().ConfigMaps(cfg.ClientCredentialsNamespace).Get(ctx, cfg.ClientCredentialsName, metav1.GetOptions{})
	if err != nil {
		log.C(ctx).Warn("Could not find configmap with client ids")
	} else {
		for region, data := range clientIdsMap.Data {
			var details oAuthDetails
			if err := json.Unmarshal([]byte(data), &details); err != nil {
				log.C(ctx).Error(err)
			} else {
				clientCreds[region] = details
			}
		}
	}

	return &kubernetesClient{
		client:      kubeClientSet,
		secret:      secret.Data,
		configmap:   configMap.Data,
		cfg:         cfg,
		clientCreds: clientCreds,
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
	return k.clientCreds[region].ClientID
}

func (k *kubernetesClient) GetClientSecretForRegion(region string) string {
	return k.clientCreds[region].ClientSecret
}

func (k *kubernetesClient) GetTokenURLForRegion(region string) string {
	return k.clientCreds[region].TokenURL
}
