package deprovisioning

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
)

func getInstance(storage storage.Instances, operation internal.DeprovisioningOperation, log logrus.FieldLogger) (*internal.Instance, error) {
	instance, err := storage.GetByID(operation.InstanceID)
	switch {
	case err == nil:
	case dberr.IsNotFound(err):
		return nil, newInstanceNotFoundError("instance already deprovisioned")
	default:
		errorMessage := fmt.Sprintf("unable to get instance from storage: %s", err)
		log.Errorf(errorMessage)
		return nil, newInstanceGetError(errorMessage)
	}
	return instance, nil
}
