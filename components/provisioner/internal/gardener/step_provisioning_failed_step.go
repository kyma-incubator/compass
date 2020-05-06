package gardener

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (r *ProvisioningOperator) ProceedToFailedStep(log *logrus.Entry, shoot gardener_types.Shoot, operationId, runtimeId, failMessage string) error {
	session := r.dbsFactory.NewReadWriteSession()
	dberr := session.UpdateOperationState(operationId, failMessage, model.Failed, time.Now())
	if dberr != nil {
		return fmt.Errorf("error: failed to set operation as failed: %s", dberr.Error())
	}

	tenant, dberr := session.GetTenant(runtimeId)
	if dberr != nil {
		return errors.Wrap(dberr, "failed to get Gardener cluster by name")
	}
	if err := r.directorClient.SetRuntimeStatusCondition(runtimeId, graphql.RuntimeStatusConditionFailed, tenant); err != nil {
		return errors.Wrap(dberr, fmt.Sprintf("failed to set runtime %s status condition", graphql.RuntimeStatusConditionFailed.String()))
	}

	err := r.updateShoot(shoot, func(s *gardener_types.Shoot) {
		removeAnnotation(s, operationIdAnnotation)
		annotate(s, provisioningAnnotation, ProvisioningFailed.String())
	})
	if err != nil {
		return fmt.Errorf("error: failed to update shoot, after installation timeout")
	}

	// TODO: we can delete Shoot if timeout occurred - maybe switched with some flag (probably not desired for debugging)

	return nil
}
