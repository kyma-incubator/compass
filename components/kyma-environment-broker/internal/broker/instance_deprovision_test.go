package broker

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker/automock"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	instanceID  = "instance-001"
	operationID = "1234"
)

func TestDeprovisionEndpoint_DeprovisionNotExistingInstance(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	queue := &automock.Queue{}
	queue.On("Add", mock.AnythingOfType("string"))

	svc := NewDeprovision(memoryStorage.Instances(), memoryStorage.Operations(), queue, logrus.StandardLogger())

	// when
	_, err := svc.Deprovision(context.TODO(), "inst-0001", domain.DeprovisionDetails{}, true)

	// then
	assert.Equal(t, apiresponses.ErrInstanceDoesNotExist, err)
}

func TestDeprovisionEndpoint_DeprovisionExistingInstance(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	memoryStorage.Instances().Insert(internal.Instance{
		InstanceID: instanceID,
	})

	queue := &automock.Queue{}
	queue.On("Add", mock.AnythingOfType("string"))

	svc := NewDeprovision(memoryStorage.Instances(), memoryStorage.Operations(), queue, logrus.StandardLogger())

	// when
	_, err := svc.Deprovision(context.TODO(), instanceID, domain.DeprovisionDetails{}, true)

	// then
	require.NoError(t, err)
	operation, err := memoryStorage.Operations().GetDeprovisioningOperationByInstanceID(instanceID)
	require.NoError(t, err)
	assert.Equal(t, domain.InProgress, operation.State)
}

func TestDeprovisionEndpoint_DeprovisionExistingOperationInProgress(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	err := memoryStorage.Instances().Insert(internal.Instance{
		InstanceID: instanceID,
	})
	require.NoError(t, err)

	err = memoryStorage.Operations().InsertDeprovisioningOperation(fixDeprovisioningOperation(domain.InProgress))
	require.NoError(t, err)

	queue := &automock.Queue{}
	queue.On("Add", mock.AnythingOfType("string"))

	svc := NewDeprovision(memoryStorage.Instances(), memoryStorage.Operations(), queue, logrus.StandardLogger())

	// when
	res, err := svc.Deprovision(context.TODO(), instanceID, domain.DeprovisionDetails{}, true)

	// then
	require.NoError(t, err)
	assert.Equal(t, operationID, res.OperationData)

	operation, err := memoryStorage.Operations().GetDeprovisioningOperationByInstanceID(instanceID)
	require.NoError(t, err)
	assert.Equal(t, domain.InProgress, operation.State)
}

func TestDeprovisionEndpoint_DeprovisionExistingOperationFailed(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	err := memoryStorage.Instances().Insert(internal.Instance{
		InstanceID: instanceID,
	})
	require.NoError(t, err)

	err = memoryStorage.Operations().InsertDeprovisioningOperation(fixDeprovisioningOperation(domain.Failed))
	require.NoError(t, err)

	queue := &automock.Queue{}
	queue.On("Add", mock.AnythingOfType("string"))

	svc := NewDeprovision(memoryStorage.Instances(), memoryStorage.Operations(), queue, logrus.StandardLogger())

	// when
	res, err := svc.Deprovision(context.TODO(), instanceID, domain.DeprovisionDetails{}, true)

	// then
	require.NoError(t, err)
	assert.Equal(t, operationID, res.OperationData)

	operation, err := memoryStorage.Operations().GetDeprovisioningOperationByInstanceID(instanceID)
	require.NoError(t, err)
	assert.Equal(t, domain.InProgress, operation.State)
}

func fixDeprovisioningOperation(state domain.LastOperationState) internal.DeprovisioningOperation {
	return internal.DeprovisioningOperation{
		Operation: internal.Operation{
			ID:         operationID,
			InstanceID: instanceID,
			State:      state,
		},
	}
}
