package gardener

import (
	"fmt"

	"github.com/kyma-project/control-plane/components/provisioner/internal/provisioning/persistence/dbsession"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/sirupsen/logrus"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
)

func NewShootController(
	mgr manager.Manager,
	dbsFactory dbsession.Factory,
	auditLogTenantConfigPath string) (*ShootController, error) {

	err := gardener_types.AddToScheme(mgr.GetScheme())
	if err != nil {
		return nil, fmt.Errorf("failed to add Gardener types to scheme: %s", err.Error())
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&gardener_types.Shoot{}).
		Complete(NewReconciler(mgr, dbsFactory, auditLogTenantConfigPath))
	if err != nil {
		return nil, fmt.Errorf("unable to create controller: %w", err)
	}

	return &ShootController{
		controllerManager: mgr,
		log:               logrus.WithField("Component", "ShootController"),
	}, nil
}

type ShootController struct {
	namespace         string
	controllerManager ctrl.Manager
	log               *logrus.Entry
}

func (sc *ShootController) StartShootController() error {
	// Start Controller
	if err := sc.controllerManager.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("error starting shoot controller: %w", err)
	}

	return nil
}
