package gardener

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ProvisioningOperator) DeprovisioningInProgress(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) (ctrl.Result, error) {
	log.Infof("Shoot is on deprovisioning in progress step")

	installationState := getInstallationState(shoot)

	// TODO: consider spliting to two steps with deinstallation
	if installationState == Uninstalling {
		// TODO: we can check status here
		return ctrl.Result{}, nil
	}

	log.Infof("Starting Uninstall")
	k8sConfig, err := KubeconfigForShoot(r.secretsClient, shoot.Name)
	if err != nil {
		log.Errorf("error fetching kubeconfig: %s", err.Error())
		return ctrl.Result{}, err
	}

	err = installationSDK.TriggerUninstall(k8sConfig)
	if err != nil {
		log.Errorf("error triggering uninstalling: %s", err.Error())
		return ctrl.Result{}, err
	}

	err = r.updateShoot(shoot, func(s *gardener_types.Shoot) {
		annotate(s, installationAnnotation, Uninstalling.String())
	})
	if err != nil {
		log.Errorf("error updating Shoot with retries: %s", err.Error())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
