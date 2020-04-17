package gardener

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ProvisioningOperator) DeprovisioningInProgress(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) (ctrl.Result, error) {
	log.Debug("Shoot is on deprovisioning in progress step")

	if uninstallTriggered(shoot) {
		return ctrl.Result{}, nil
	}

	log.Infof("Starting Uninstall")
	k8sConfig, err := KubeconfigForShoot(r.secretsClient, shoot.Name)
	if err != nil {
		log.Errorf("error fetching kubeconfig: %s", err.Error())
		return ctrl.Result{}, err
	}

	err = r.installationSvc.TriggerUninstall(k8sConfig)
	if err != nil {
		log.Errorf("error triggering uninstalling: %s", err.Error())
		return ctrl.Result{}, err
	}

	err = r.updateShoot(shoot, func(s *gardener_types.Shoot) {
		annotate(s, uninstallingAnnotation, "true")
	})
	if err != nil {
		log.Errorf("error updating Shoot with retries: %s", err.Error())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
