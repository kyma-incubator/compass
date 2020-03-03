package gardener

import (
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ProvisioningOperator) ProvisioningInProgress(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) (ctrl.Result, error) {
	lastOperation := shoot.Status.LastOperation

	if lastOperation.State == gardencorev1alpha1.LastOperationStateSucceeded {
		err := r.ProceedToInstallation(log, shoot, operationId)
		if err != nil {
			log.Errorf("Error proceeding to installation: %s", err.Error())
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if isShootFailed(shoot) {
		log.Infof("Provisioning failed! Last state: %s, Description: %s", lastOperation.State, lastOperation.Description)
		err := r.ProceedToFailedStep(log, shoot, operationId, "Provisioning failed.")
		if err != nil {
			log.Errorf("error proceeding to provisioning failed step: %s", err.Error())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	log.Infof("Provisioning in progress. Last state: %s, Description: %s", lastOperation.State, lastOperation.Description)
	return ctrl.Result{}, nil
}
