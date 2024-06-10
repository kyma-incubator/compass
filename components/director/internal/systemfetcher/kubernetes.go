package systemfetcher

import (
	"context"
	"strconv"
	"time"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KubeClient represents client responsible for kubernetes operations
//
//go:generate mockery --name=KubeClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type KubeClient interface {
	GetSystemFetcherSecretData(ctx context.Context, secretName string) ([]byte, error)
}

// KubeConfig represents Kubernetes configuration
type KubeConfig struct {
	UseKubernetes string `envconfig:"APP_USE_KUBERNETES"`

	SecretNamespace string `envconfig:"APP_SECRET_NAMESPACE" default:"compass-system"`

	PollInterval time.Duration `envconfig:"APP_KUBERNETES_POLL_INTERVAL" default:"2s"`
	PollTimeout  time.Duration `envconfig:"APP_KUBERNETES_POLL_TIMEOUT" default:"1m"`
	Timeout      time.Duration `envconfig:"APP_KUBERNETES_TIMEOUT" default:"2m"`
}

// NewKubernetesClient creates new Kubernetes client
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
		cfg: KubeConfig{SecretNamespace: "namespace"},
	}
}

// GetSystemFetcherSecretData Gets secret data from secret with given name
func (k *noopKubernetesClient) GetSystemFetcherSecretData(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
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

// GetSystemFetcherSecretData Gets secret data from secret with given name
func (k *kubernetesClient) GetSystemFetcherSecretData(ctx context.Context, secretName string) ([]byte, error) {
	secret, err := k.client.CoreV1().Secrets(k.cfg.SecretNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
		//return nil, errors.Wrapf(err, "while getting secret with name: %q in namespace: %q", secretName, k.cfg.SecretNamespace)
	}

	secretDataBytes, ok := secret.Data["data"]
	if !ok || len(secretDataBytes) == 0 {
		return nil, errors.Wrapf(err, "There is no data in secret with name: %q in namespace: %q. No credentials from secret will be set", secretName, k.cfg.SecretNamespace)
	}

	return secretDataBytes, nil
}
