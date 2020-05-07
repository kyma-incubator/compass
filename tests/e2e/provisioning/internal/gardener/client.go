package gardener

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewGardenerClusterConfig returns REST config for Gardener cluster
func NewGardenerClusterConfig(kubeconfigPath string) (*restclient.Config, error) {
	rawKubeconfig, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read Gardener Kubeconfig from path: %s", kubeconfigPath)
	}

	gardenerClusterConfig, err := RESTConfig(rawKubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the REST config from Kubeconfig")
	}

	return gardenerClusterConfig, nil
}

// NewGardenerSecretsInterface returns k8s secrets client(interface) for Gardener cluster
func NewGardenerSecretsInterface(gardenerClusterCfg *restclient.Config, gardenerProjectName string) (corev1.SecretInterface, error) {
	gardenerNamespace := fmt.Sprintf("garden-%s", gardenerProjectName)
	gardenerClusterClient, err := kubernetes.NewForConfig(gardenerClusterCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch gardener cluster clienset")
	}

	return gardenerClusterClient.CoreV1().Secrets(gardenerNamespace), nil
}

// RESTConfig returns REST config
func RESTConfig(kubeconfig []byte) (*restclient.Config, error) {
	return clientcmd.RESTConfigFromKubeConfig(kubeconfig)
}
