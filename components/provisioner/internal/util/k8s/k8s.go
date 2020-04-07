package k8s

import (
	"fmt"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func ParseToK8sConfig(kubeconfigRaw []byte) (*restclient.Config, error) {
	kubeconfig, err := clientcmd.NewClientConfigFromBytes(kubeconfigRaw)
	if err != nil {
		return nil, fmt.Errorf("error constructing kubeconfig from raw config: %s", err.Error())
	}

	clientConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client kubeconfig from parsed config: %s", err.Error())
	}

	return clientConfig, nil
}
