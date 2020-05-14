package gardener

import (
	"context"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes/scheme"

	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceCleanupTimeouts struct {
	ServiceInstance time.Duration `envconfig:"default=20m"`
}

//go:generate mockery -name=ResourcesJanitor
type ResourcesJanitor interface {
	CleanUpShootResources(shootName string) error
}

type resourcesJanitor struct {
	cleanupTimeouts ResourceCleanupTimeouts
	secretsClient   v1core.SecretInterface
	shootClient     gardener_apis.ShootInterface
}

func NewResourcesJanitor(
	cleanupTimeouts ResourceCleanupTimeouts,
	secretsClient v1core.SecretInterface,
	shootClient gardener_apis.ShootInterface,
) (ResourcesJanitor, error) {
	return &resourcesJanitor{
		cleanupTimeouts: cleanupTimeouts,
		secretsClient:   secretsClient,
		shootClient:     shootClient,
	}, nil
}

func (rj *resourcesJanitor) CleanUpShootResources(shootName string) error {
	kubeconfig, err := KubeconfigForShoot(rj.secretsClient, shootName)
	if err != nil {
		return errors.Wrap(err, "error fetching kubeconfig")
	}

	err = v1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return errors.Wrap(err, "while adding schema")
	}

	k8sCli, err := cli.New(kubeconfig, cli.Options{})
	if err != nil {
		return errors.Wrap(err, "while creating k8s client")
	}

	err = wait.Poll(20*time.Second, rj.cleanupTimeouts.ServiceInstance, func() (bool, error) {
		if err := k8sCli.Delete(context.Background(), &v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "uaa-issuer",
				Namespace: "kyma-system",
			},
		}); err != nil {
			if apiErrors.IsNotFound(err) {
				return true, nil
			}
			fmt.Println(fmt.Errorf("while deleting service instance: %s", err.Error())) //TODO: log failed SI name and info
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "while removing UAA instance")
	}
	return nil
}
