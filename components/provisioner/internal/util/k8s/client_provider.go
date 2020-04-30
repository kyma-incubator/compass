package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

//go:generate mockery -name=K8sClientProvider
type K8sClientProvider interface {
	CreateK8SClient(kubeconfigRaw string) (kubernetes.Interface, error)
}

type k8sClientBuilder struct{}

func NewK8sClientProvider() K8sClientProvider {
	return &k8sClientBuilder{}
}

func (c *k8sClientBuilder) CreateK8SClient(kubeconfigRaw string) (kubernetes.Interface, error) {
	k8sConfig, err := ParseToK8sConfig([]byte(kubeconfigRaw))

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s core client, %s", err.Error())
	}

	return coreClientset, nil
}
