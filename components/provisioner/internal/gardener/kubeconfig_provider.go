package gardener

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
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
