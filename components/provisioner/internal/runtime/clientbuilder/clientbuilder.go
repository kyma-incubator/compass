package clientbuilder

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockery -name=ConfigMapClientBuilder
type ConfigMapClientBuilder interface {
	CreateK8SConfigMapClient(kubeconfigRaw, namespace string) (v1.ConfigMapInterface, error)
}

type configMapClientBuilder struct{}

func NewConfigMapClientBuilder() ConfigMapClientBuilder {
	return &configMapClientBuilder{}
}

func (c *configMapClientBuilder) CreateK8SConfigMapClient(kubeconfigRaw, namespace string) (v1.ConfigMapInterface, error) {
	kubeconfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeconfigRaw))
	if err != nil {
		return nil, fmt.Errorf("error constructing kubeconfig from raw config: %s", err.Error())
	}

	k8sConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client kubeconfig from parsed config: %s", err.Error())
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	return coreClientset.CoreV1().ConfigMaps(namespace), nil
}
