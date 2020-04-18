package gardener

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	gclientset "github.com/gardener/gardener/pkg/client/core/clientset/versioned"
)

type Client struct {
	Namespace           string
	GardenerClientset   gclientset.Interface
	KubernetesClientset kubernetes.Interface
}

func NewClient(kubeconfig string) (*Client, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	ns, _, err := kubeConfig.Namespace()
	if err != nil {
		return nil, err
	}

	gclient, err := gclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Namespace:           ns,
		GardenerClientset:   gclient,
		KubernetesClientset: kclient,
	}

	return client, nil
}
