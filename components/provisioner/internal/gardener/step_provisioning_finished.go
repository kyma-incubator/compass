package gardener

import (
	"fmt"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ProvisioningOperator) ProceedToFinishedStep(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) error {
	log.Infof("Kyma is installed. Setting operation as succeeded.")

	session := r.dbsFactory.NewWriteSession()
	dberr := session.UpdateOperationState(operationId, "Operation succeeded.", model.Succeeded)
	if dberr != nil {
		return fmt.Errorf("error setting operation as succeeded: %s", dberr.Error())
	}

	err := r.updateShoot(shoot, func(shootToUpdate *gardener_types.Shoot) {
		annotate(shootToUpdate, installationAnnotation, Installed.String())
		annotate(shootToUpdate, provisioningStepAnnotation, ProvisioningFinishedStep.String())
		removeAnnotation(shootToUpdate, operationIdAnnotation)
	})
	if err != nil {
		return fmt.Errorf("Error updating Shoot with retries: %s", err.Error())
	}

	return nil
}

func (r *ProvisioningOperator) ProvisioningFinished(log *logrus.Entry, shoot gardener_types.Shoot) (ctrl.Result, error) {
	log.Infof("Cluster is provisioned and Kyma is installed")

	// TODO: here we can ensure Kyma is installed properly

	return ctrl.Result{}, nil
}
