package broker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

type LastOperationEndpoint struct {
	operationStorage storage.Operations
	dumper           StructDumper
}

func NewLastOperation(os storage.Operations, d StructDumper) *LastOperationEndpoint {
	return &LastOperationEndpoint{
		operationStorage: os,
		dumper:           d,
	}
}

// LastOperation fetches last operation state for a service instance
//   GET /v2/service_instances/{instance_id}/last_operation
func (b *LastOperationEndpoint) LastOperation(ctx context.Context, instanceID string, details domain.PollDetails) (domain.LastOperation, error) {
	operation, err := b.operationStorage.GetProvisioningOperationByID(details.OperationData)
	if err != nil {
		b.dumper.Dump(fmt.Sprintf("cannot get operation from storage: %s", err))
		return domain.LastOperation{}, errors.Wrapf(err, "while getting operation from storage")
	}

	if operation.InstanceID != instanceID {
		err := errors.Errorf("operation %s with instance ID %s not exist", operation.ID, instanceID)
		return domain.LastOperation{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, err.Error())
	}

	return domain.LastOperation{
		State:       operation.State,
		Description: operation.Description,
	}, nil
}
