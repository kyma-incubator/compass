package provisioning

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	persistenceMocks "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_ProvisionRuntime(t *testing.T) {
	hydroformMock := &mocks.Service{}
	persistenceServiceMock := &persistenceMocks.Service{}
	uuidGenerator := &persistenceMocks.UUIDGenerator{}

	clusterConfig := &gqlschema.ClusterConfigInput{
		GcpConfig: &gqlschema.GCPConfigInput{
			Name:              "Something",
			ProjectName:       "Project",
			NumberOfNodes:     3,
			BootDiskSize:      "256",
			MachineType:       "machine",
			Region:            "region",
			Zone:              new(string),
			KubernetesVersion: "version",
		},
	}

	kymaConfig := &gqlschema.KymaConfigInput{
		Version: "1.5",
		Modules: gqlschema.AllKymaModule,
	}

	t.Run("Should start runtime provisioning and return operation ID", func(t *testing.T) {
		//given
		runtimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		uuidGenerator.On("New").Return("id", nil)

		persistenceServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{}, dberrors.NotFound("Not found"))
		persistenceServiceMock.On("SetProvisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("Update", runtimeID, "", "").Return(nil)
		persistenceServiceMock.On("SetAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: "", State: ""}, nil)

		service := NewProvisioningService(persistenceServiceMock, uuidGenerator, hydroformMock)

		//when
		operationID, finished, err := service.ProvisionRuntime(runtimeID, gqlschema.ProvisionRuntimeInput{ClusterConfig: clusterConfig, Credentials: &gqlschema.CredentialsInput{}, KymaConfig: kymaConfig})
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expOperationID, operationID)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})

	t.Run("Should start runtime provisioning and return operation ID when previous provisioning failed", func(t *testing.T) {
		//given
		runtimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		persistenceServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{Type: model.Provision, State: model.Failed}, nil)
		persistenceServiceMock.On("SetProvisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("Update", runtimeID, "", "").Return(nil)
		persistenceServiceMock.On("SetAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: "", State: ""}, nil)

		service := NewProvisioningService(persistenceServiceMock, uuidGenerator, hydroformMock)

		//when
		operationID, finished, err := service.ProvisionRuntime(runtimeID, gqlschema.ProvisionRuntimeInput{ClusterConfig: clusterConfig, Credentials: &gqlschema.CredentialsInput{}, KymaConfig: kymaConfig})
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expOperationID, operationID)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})

	t.Run("Should return error when cluster is already provisioned", func(t *testing.T) {
		//given
		runtimeID := "0ad91f16-d553-413f-aa27-4eefd9e5f1c6"
		persistenceServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{}, nil)
		uuidGenerator := &persistenceMocks.UUIDGenerator{}

		service := NewProvisioningService(persistenceServiceMock, uuidGenerator, hydroformMock)

		//when
		_, _, err := service.ProvisionRuntime(runtimeID, gqlschema.ProvisionRuntimeInput{ClusterConfig: clusterConfig, Credentials: &gqlschema.CredentialsInput{}, KymaConfig: kymaConfig})

		//then
		require.Error(t, err)
		persistenceServiceMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})
}

func TestService_DeprovisionRuntime(t *testing.T) {
	persistenceServiceMock := &persistenceMocks.Service{}
	hydroformMock := &mocks.Service{}
	uuidGenerator := &persistenceMocks.UUIDGenerator{}

	runtimeConfig := model.RuntimeConfig{
		ClusterConfig: model.GCPConfig{},
	}

	t.Run("Should start runtime deprovisioning and return operation ID", func(t *testing.T) {
		//given
		lastOperation := model.Operation{State: model.Succeeded}
		runtimeStatus := model.RuntimeStatus{LastOperationStatus: lastOperation, RuntimeConfiguration: runtimeConfig}

		runtimeID := "92a1c394-639a-424e-8578-ba1ca7501dc1"
		expOperationID := "c7241d2d-5ffd-434b-9a52-17ce9ee04578"
		operation := model.Operation{ID: expOperationID}

		persistenceServiceMock.On("GetStatus", runtimeID).Return(runtimeStatus, nil)
		persistenceServiceMock.On("SetDeprovisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("GetClusterData", runtimeID).Return(model.Cluster{TerraformState: "{}"}, nil)
		persistenceServiceMock.On("SetAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("DeprovisionCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		resolver := NewProvisioningService(persistenceServiceMock, uuidGenerator, hydroformMock)

		//when
		opt, finished, err := resolver.DeprovisionRuntime(runtimeID)
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expOperationID, opt)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})

	t.Run("Should not start deprovisioning when previous operation is in progress", func(t *testing.T) {
		//given
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"
		persistenceServiceMock := &persistenceMocks.Service{}
		lastOperation := model.Operation{State: model.InProgress}
		runtimeStatus := model.RuntimeStatus{LastOperationStatus: lastOperation}

		persistenceServiceMock.On("GetStatus", runtimeID).Return(runtimeStatus, nil)

		resolver := NewProvisioningService(persistenceServiceMock, uuidGenerator, hydroformMock)

		//when
		_, _, err := resolver.DeprovisionRuntime(runtimeID)

		//then
		require.Error(t, err)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})
}

func TestService_RuntimeOperationStatus(t *testing.T) {
	persistenceServiceMock := &persistenceMocks.Service{}
	hydroformMock := &mocks.Service{}
	uuidGenerator := &persistenceMocks.UUIDGenerator{}

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		operationID := "999f2260-8cae-4367-a38a-2355b71bf054"
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"

		operation := model.Operation{
			ID:        operationID,
			Type:      model.Provision,
			State:     model.InProgress,
			ClusterID: runtimeID,
			Message:   "some message",
		}

		persistenceServiceMock.On("Get", operationID).Return(operation, nil)
		resolver := NewProvisioningService(persistenceServiceMock, uuidGenerator, hydroformMock)

		//when
		status, err := resolver.RuntimeOperationStatus(operationID)

		//then
		require.NoError(t, err)
		assert.Equal(t, gqlschema.OperationTypeProvision, status.Operation)
		assert.Equal(t, gqlschema.OperationStateInProgress, status.State)
		assert.Equal(t, operation.ClusterID, *status.RuntimeID)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})
}

func waitUntilFinished(finished <-chan struct{}) {
	for {
		_, ok := <-finished
		if !ok {
			break
		}
	}
}
