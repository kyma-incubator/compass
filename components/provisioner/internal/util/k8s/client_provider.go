package k8s

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/apperrors"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockery -name=K8sClientProvider
type K8sClientProvider interface {
	CreateK8SClient(kubeconfigRaw string) (kubernetes.Interface, apperrors.AppError)
}

type k8sClientBuilder struct{}

func NewK8sClientProvider() K8sClientProvider {
	return &k8sClientBuilder{}
}

func (c *k8sClientBuilder) CreateK8SClient(kubeconfigRaw string) (kubernetes.Interface, apperrors.AppError) {
	k8sConfig, err := ParseToK8sConfig([]byte(kubeconfigRaw))

	if err != nil {
		return nil, apperrors.Internal("failed to parse kubeconfig: %s", err.Error())
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s core client: %s", err.Error())
	}

	return coreClientset, nil
}
