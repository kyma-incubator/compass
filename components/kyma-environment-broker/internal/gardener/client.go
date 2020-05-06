package gardener

import (
	"fmt"
	"io/ioutil"

	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewGardenerClusterConfig(kubeconfigPath string) (*restclient.Config, error) {

	rawKubeconfig, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Gardener Kubeconfig from path %s: %s", kubeconfigPath, err.Error())
	}

	gardenerClusterConfig, err := RESTConfig(rawKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	return gardenerClusterConfig, nil
}

func NewGardenerSecretsInterface(gardenerClusterCfg *restclient.Config, gardenerProjectName string) (corev1.SecretInterface, error) {

	gardenerNamespace := fmt.Sprintf("garden-%s", gardenerProjectName)

	gardenerClusterClient, err := kubernetes.NewForConfig(gardenerClusterCfg)
	if err != nil {
		return nil, err
	}

	return gardenerClusterClient.CoreV1().Secrets(gardenerNamespace), nil
}

func RESTConfig(kubeconfig []byte) (*restclient.Config, error) {
	return clientcmd.RESTConfigFromKubeConfig(kubeconfig)
}

func NewClient(config *restclient.Config) (*gardener_apis.CoreV1beta1Client, error) {
	clientset, err := gardener_apis.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
