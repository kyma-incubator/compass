package gardener

import (
	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Config(kubeconfig []byte) (*restclient.Config, error) {
	return clientcmd.RESTConfigFromKubeConfig(kubeconfig)
}

func NewClient(config *restclient.Config) (*gardener_apis.CoreV1beta1Client, error) {
	clientset, err := gardener_apis.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
