package gardener

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operation"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	v1core "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"

	"github.com/sirupsen/logrus"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
)

func NewShootController(
	mgr manager.Manager,
	shootClient gardener_apis.ShootInterface,
	secretsClient v1core.SecretInterface,
	dbsFactory dbsession.Factory,
	directorClient director.DirectorClient,
	installQueue operation.OperationQueue) (*ShootController, error) {

	err := gardener_types.AddToScheme(mgr.GetScheme())
	if err != nil {
		return nil, fmt.Errorf("failed to add Gardener types to scheme: %s", err.Error())
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&gardener_types.Shoot{}).
		Complete(NewReconciler(mgr, dbsFactory, secretsClient, shootClient, directorClient, installQueue))
	if err != nil {
		return nil, fmt.Errorf("unable to create controller: %w", err)
	}

	return &ShootController{
		controllerManager: mgr,
		shootClient:       shootClient,
		log:               logrus.WithField("Component", "ShootController"),
	}, nil
}

type ShootController struct {
	namespace         string
	controllerManager ctrl.Manager
	shootClient       gardener_apis.ShootInterface
	log               *logrus.Entry
}

func (sc *ShootController) StartShootController() error {
	// Start Controller
	if err := sc.controllerManager.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("error starting shoot controller: %w", err)
	}

	return nil
}
