package gardener

import (
	"fmt"

	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	gardener_types "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/client/garden/clientset/versioned/typed/garden/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateAndDeleteShoot(client v1beta1.ShootInterface, shoot *gardener_types.Shoot) error {
	_, err := client.Update(shoot)
	if err != nil {
		return fmt.Errorf("error annotating Shoot with confirm deletion label: %s", err.Error())
	}

	err = client.Delete(shoot.Name, &v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting shoot cluster: %s", err.Error())
	}

	return nil
}

func AnnotateWithConfirmDeletion(shoot *gardener_types.Shoot) {
	if shoot.Annotations == nil {
		shoot.Annotations = map[string]string{}
	}

	shoot.Annotations["confirmation.garden.sapcloud.io/deletion"] = "true"
}

func FetchKubeconfigForShoot(secretsClient v12.SecretInterface, shootName string) ([]byte, error) {
	secret, err := secretsClient.Get(fmt.Sprintf("%s.kubeconfig", shootName), v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetching kubeconfig: %s", err.Error())
	}

	kubeconfig, found := secret.Data["kubeconfig"]
	if !found {
		return nil, fmt.Errorf("error fetching kubeconfig: secret does not contain kubeconfig")
	}

	return kubeconfig, nil
}

func KubeconfigForShoot(secretsClient v12.SecretInterface, shootName string) (*restclient.Config, error) {
	kubeconfigRaw, err := FetchKubeconfigForShoot(secretsClient, shootName)
	if err != nil {
		return nil, fmt.Errorf("error fetching kubeconfig: %s", err.Error())
	}

	return ParseToK8sConfig(kubeconfigRaw)

}

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
