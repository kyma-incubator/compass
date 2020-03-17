package postsql

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/pivotal-cf/brokerapi/v7/domain"

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
	var lastErr error
	_ = wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		lastErr = session.InsertOperation(dto)
		if lastErr != nil {
			log.Warn(errors.Wrap(err, "while insert operation"))
			return false, nil
		}
		return true, nil
	})
	return lastErr
}

// GetProvisioningOperationByID fetches the ProvisioningOperation by given ID, returns error if not found
func (s *operations) GetProvisioningOperationByID(operationID string) (*internal.ProvisioningOperation, error) {
	session := s.NewReadSession()
	operation := dbsession.OperationDTO{}
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		dto, dbErr := session.GetOperationByID(operationID)
		if dbErr != nil {
			log.Warn(errors.Wrapf(dbErr, "while reading Operation from the storage"))
			return false, nil
		}
		operation = dto
		return true, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "while getting operation by ID")
	}
	ret, err := toProvisioningOperation(&operation)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting DTO to Operation")
	}

	return ret, nil
}

// GetProvisioningOperationByInstanceID fetches the ProvisioningOperation by given instanceID, returns error if not found
func (s *operations) GetProvisioningOperationByInstanceID(instanceID string) (*internal.ProvisioningOperation, error) {
	session := s.NewReadSession()
	operation := dbsession.OperationDTO{}
	var lastErr dberr.Error
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		operation, lastErr = session.GetOperationByInstanceID(instanceID)
		if lastErr != nil {
			if lastErr.Code() == dberr.CodeNotFound {
				lastErr = dberr.NotFound("operation does not exist")
				return false, lastErr
			}
			log.Warn(errors.Wrapf(lastErr, "while reading Operation from the storage").Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, lastErr
	}
	ret, err := toProvisioningOperation(&operation)
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

	var lastErr error
	err = wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		dErr := session.UpdateOperation(dto)
		if dErr != nil && dErr.Code() == dberr.CodeNotFound {
			_, lastErr = s.NewReadSession().GetOperationByID(op.ID)
			if lastErr != nil {
				log.Warn(errors.Wrapf(lastErr, "while getting Operation").Error())
				return false, nil
			}

			// the operation exists but the version is different
			lastErr = dberr.Conflict("operation update conflict, operation ID: %s", op.ID)
			log.Warn(lastErr.Error())
			return false, lastErr
		}
		return true, nil
	})
	op.Version = op.Version + 1
	return &op, lastErr
}

// GetOperation returns Operation with given ID. Returns an error if the operation does not exists.
func (s *operations) GetOperation(operationID string) (*internal.Operation, error) {
	session := s.NewReadSession()
	operation := dbsession.OperationDTO{}
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		dto, err := session.GetOperationByID(operationID)
		if err != nil {
			log.Warn(errors.Wrapf(err, "while reading Operation from the storage").Error())
			return false, nil
		}
		operation = dto
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	op := toOperation(&operation)
	return &op, nil
}

func (s *operations) GetOperationsInProgressByType(operationType dbsession.OperationType) ([]internal.Operation, error) {
	session := s.NewReadSession()
	operations := make([]dbsession.OperationDTO, 0)
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		dto, err := session.GetOperationsInProgressByType(operationType)
		if err != nil {
			log.Warn(errors.Wrapf(err, "while getting Operations from the storage").Error())
			return false, nil
		}
		operations = dto
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return toOperations(operations), nil
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

func toOperations(op []dbsession.OperationDTO) []internal.Operation {
	operations := make([]internal.Operation, 0)
	for _, o := range op {
		operations = append(operations, toOperation(&o))
	}
	return operations
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
