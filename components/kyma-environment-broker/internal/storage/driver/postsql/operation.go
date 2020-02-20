package postsql

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/pivotal-cf/brokerapi/v7/domain"

	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/pkg/errors"
)

type operations struct {
	dbsession.Factory
}

func NewOperation(sess dbsession.Factory) *operations {
	return &operations{
		Factory: sess,
	}
}

// InsertProvisioningOperation insert new operation
func (s *operations) InsertProvisioningOperation(operation internal.ProvisioningOperation) error {
	session := s.NewWriteSession()
	dto, err := provisioningOperationToDTO(&operation)
	if err != nil {
		return errors.Wrapf(err, "while inserting provisioning operation (id: %s)", operation.ID)
	}

	return session.InsertOperation(dto)
}

// GetProvisioningOperationByID fetches the ProvisioningOperation by given ID, returns error if not found
func (s *operations) GetProvisioningOperationByID(operationID string) (*internal.ProvisioningOperation, error) {
	session := s.NewReadSession()
	dto, dbErr := session.GetOperationByID(operationID)
	if dbErr != nil {
		return nil, errors.Wrapf(dbErr, "while reading Operation from the storage")
	}

	ret, err := toProvisioningOperation(&dto)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting DTO to Operation")
	}

	return ret, nil
}

// UpdateProvisioningOperation updates ProvisioningOperation, fails if not exists or optimistic locking failure occurs.
func (s *operations) UpdateProvisioningOperation(op internal.ProvisioningOperation) (*internal.ProvisioningOperation, error) {
	session := s.NewWriteSession()
	op.UpdatedAt = time.Now()
	dto, err := provisioningOperationToDTO(&op)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Operation to DTO")
	}

	dErr := session.UpdateOperation(dto)
	if dErr != nil && dErr.Code() == dberr.CodeNotFound {
		_, err := s.NewReadSession().GetOperationByID(op.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting Operation")
		}

		// the operation exists but the version is different
		return nil, dberr.Conflict("operation update conflict, operation ID: %s", op.ID)
	}
	op.Version = op.Version + 1
	return &op, err
}

// GetOperation returns Operation with given ID. Returns an error if the operation does not exists.
func (s *operations) GetOperation(operationID string) (*internal.Operation, error) {
	session := s.NewReadSession()
	dto, err := session.GetOperationByID(operationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading Operation from the storage")
	}
	op := toOperation(&dto)
	return &op, nil
}

func toOperation(op *dbsession.OperationDTO) internal.Operation {
	return internal.Operation{
		ID:                     op.ID,
		CreatedAt:              op.CreatedAt,
		UpdatedAt:              op.UpdatedAt,
		ProvisionerOperationID: op.TargetOperationID,
		State:                  domain.LastOperationState(op.State),
		InstanceID:             op.InstanceID,
		Description:            op.Description,
		Version:                op.Version,
	}
}

func toProvisioningOperation(op *dbsession.OperationDTO) (*internal.ProvisioningOperation, error) {
	if op.Type != dbsession.OperationTypeProvision {
		return nil, errors.New(fmt.Sprintf("expected operation type Provisioning, but was %s", op.Type))
	}
	var operation internal.ProvisioningOperation
	err := json.Unmarshal([]byte(op.Data), &operation)
	if err != nil {
		return nil, errors.New("unable to unmarshall provisioning data")
	}
	operation.Operation = toOperation(op)

	return &operation, nil
}

func provisioningOperationToDTO(op *internal.ProvisioningOperation) (dbsession.OperationDTO, error) {
	serialized, err := json.Marshal(op)
	if err != nil {
		return dbsession.OperationDTO{}, errors.Wrapf(err, "while serializing provisioning data %v", op)
	}

	ret := operationToDB(&op.Operation)
	ret.Data = string(serialized)
	ret.Type = dbsession.OperationTypeProvision
	return ret, nil
}

func operationToDB(op *internal.Operation) dbsession.OperationDTO {
	return dbsession.OperationDTO{
		ID:                op.ID,
		TargetOperationID: op.ProvisionerOperationID,
		State:             string(op.State),
		Description:       op.Description,
		UpdatedAt:         op.UpdatedAt,
		CreatedAt:         op.CreatedAt,
		Version:           op.Version,
		InstanceID:        op.InstanceID,
	}
}
