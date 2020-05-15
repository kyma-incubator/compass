package gardener

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeconfigProvider struct {
	secretsClient v12.SecretInterface
}

func NewKubeconfigProvider(secretsClient v12.SecretInterface) KubeconfigProvider {
	return KubeconfigProvider{
		secretsClient: secretsClient,
	}
}

func (kp KubeconfigProvider) FetchRaw(shootName string) ([]byte, error) {
	secret, err := kp.secretsClient.Get(fmt.Sprintf("%s.kubeconfig", shootName), v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetching kubeconfig: %s", err.Error())
	}

	kubeconfig, found := secret.Data["kubeconfig"]
	if !found {
		return nil, fmt.Errorf("error fetching kubeconfig: secret does not contain kubeconfig")
	}

	return kubeconfig, nil
}

func (kp KubeconfigProvider) Fetch(shootName string) (*restclient.Config, error) {
	kubeconfigRaw, err := kp.FetchRaw(shootName)
	if err != nil {
		return nil, fmt.Errorf("error fetching kubeconfig: %s", err.Error())
	}

	return parseToK8sConfig(kubeconfigRaw)
}

func parseToK8sConfig(kubeconfigRaw []byte) (*restclient.Config, error) {
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
