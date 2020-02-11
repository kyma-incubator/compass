package gardener

import (
	"fmt"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
)

func (r *ProvisioningOperator) ProceedToFailedStep(log *logrus.Entry, shoot gardener_types.Shoot, operationId, failMessage string) error {
	session := r.dbsFactory.NewWriteSession()
	dberr := session.UpdateOperationState(operationId, failMessage, model.Failed)
	if dberr != nil {
		return fmt.Errorf("error: failed to set operation as failed: %s", dberr.Error())
	}

	err := r.updateShoot(shoot, func(s *gardener_types.Shoot) {
		removeAnnotation(s, operationIdAnnotation)
		annotate(s, installationAnnotation, InstallationFailed.String())
		annotate(s, provisioningStepAnnotation, ProvisioningFailedStep.String())
	})
	if err != nil {
		return fmt.Errorf("error: failed to update shoot, after installation timeout")
	}

	// TODO: we can delete Shoot if timeout occurred - maybe switched with some flag (probably not desired for debugging)

	return nil
}
