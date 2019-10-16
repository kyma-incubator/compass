package provisioning

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"testing"

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
	hydroformMock := &mocks.Client{}
	operationServiceMock := &persistenceMocks.OperationService{}
	runtimeServiceMock := &persistenceMocks.RuntimeService{}

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

	t.Run("Should start runtime provisioning and return operation ID", func(t *testing.T) {
		//given
		runtimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		runtimeServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{}, dberrors.NotFound("Not found"))
		runtimeServiceMock.On("SetProvisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		runtimeServiceMock.On("Update", runtimeID, "", "").Return(nil)
		operationServiceMock.On("SetAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: "", State: ""}, nil)

		service := NewProvisioningService(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		operationID, err, finished := service.ProvisionRuntime(runtimeID, gqlschema.ProvisionRuntimeInput{clusterConfig, &gqlschema.CredentialsInput{}, &gqlschema.KymaConfigInput{}})
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expOperationID, operationID)
	})

	t.Run("Should start runtime provisioning and return operation ID when previous provisioning failed", func(t *testing.T) {
		//given
		runtimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		runtimeServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{Type: model.Provision, State: model.Failed}, nil)
		runtimeServiceMock.On("SetProvisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		runtimeServiceMock.On("Update", runtimeID, "", "").Return(nil)
		operationServiceMock.On("SetAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: "", State: ""}, nil)

		service := NewProvisioningService(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		operationID, err, finished := service.ProvisionRuntime(runtimeID, gqlschema.ProvisionRuntimeInput{ClusterConfig: clusterConfig, Credentials: &gqlschema.CredentialsInput{}, KymaConfig: &gqlschema.KymaConfigInput{}})
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expOperationID, operationID)
	})

	t.Run("Should return error when cluster is already provisioned", func(t *testing.T) {
		//given
		runtimeID := "0ad91f16-d553-413f-aa27-4eefd9e5f1c6"
		runtimeServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{}, nil)

		service := NewProvisioningService(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		_, err, _ := service.ProvisionRuntime(runtimeID, gqlschema.ProvisionRuntimeInput{ClusterConfig: clusterConfig, Credentials: &gqlschema.CredentialsInput{}, KymaConfig: &gqlschema.KymaConfigInput{}})

		//then
		require.Error(t, err)
	})
}

func TestService_DeprovisionRuntime(t *testing.T) {
	operationServiceMock := &persistenceMocks.OperationService{}
	runtimeServiceMock := &persistenceMocks.RuntimeService{}
	hydroformMock := &mocks.Client{}

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

		runtimeServiceMock.On("GetStatus", runtimeID).Return(runtimeStatus, nil)
		runtimeServiceMock.On("SetDeprovisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		operationServiceMock.On("SetAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("DeprovisionCluster", mock.Anything, mock.Anything).Return(nil)

		resolver := NewProvisioningService(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		opt, err, finished := resolver.DeprovisionRuntime(runtimeID, gqlschema.CredentialsInput{})
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expOperationID, opt)
	})

	t.Run("Should not start deprovisioning when previous operation is in progress", func(t *testing.T) {
		//given
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"
		runtimeServiceMock := &persistenceMocks.RuntimeService{}
		lastOperation := model.Operation{State: model.InProgress}
		runtimeStatus := model.RuntimeStatus{LastOperationStatus: lastOperation}

		runtimeServiceMock.On("GetStatus", runtimeID).Return(runtimeStatus, nil)

		resolver := NewProvisioningService(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		_, err, _ := resolver.DeprovisionRuntime(runtimeID, gqlschema.CredentialsInput{})

		//then
		require.Error(t, err)
	})
}

func TestService_RuntimeOperationStatus(t *testing.T) {
	operationServiceMock := &persistenceMocks.OperationService{}
	runtimeServiceMock := &persistenceMocks.RuntimeService{}
	hydroformMock := &mocks.Client{}

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

		operationServiceMock.On("Get", operationID).Return(operation, nil)
		resolver := NewProvisioningService(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		status, err := resolver.RuntimeOperationStatus(operationID)

		//then
		require.NoError(t, err)
		assert.Equal(t, gqlschema.OperationTypeProvision, status.Operation)
		assert.Equal(t, gqlschema.OperationStateInProgress, status.State)
		assert.Equal(t, operation.ClusterID, status.RuntimeID)
	})
}

func waitUntilFinished(finished chan interface{}) {
	for {
		_, ok := <-finished
		if !ok {
			break
		}
	}
}
