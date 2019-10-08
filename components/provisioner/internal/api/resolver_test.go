package api

import (
	"context"
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

func TestResolver_ProvisionRuntime(t *testing.T) {
	hydroformMock := mocks.Client{}
	operationServiceMock := persistenceMocks.OperationService{}
	runtimeServiceMock := persistenceMocks.RuntimeService{}
	ctx := context.Background()

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
		operation := model.Operation{OperationID: expOperationID}

		runtimeServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{}, dberrors.NotFound("Not found"))
		runtimeServiceMock.On("SetProvisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		operationServiceMock.On("SetAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(types.ClusterStatus{Phase: types.Provisioned}, nil)

		resolver := NewResolver(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		operationID, err := resolver.ProvisionRuntime(ctx, runtimeID, &gqlschema.ProvisionRuntimeInput{clusterConfig, &gqlschema.CredentialsInput{}, &gqlschema.KymaConfigInput{}})

		//then
		require.NoError(t, err)
		assert.Equal(t, expOperationID, operationID)
	})

	t.Run("Should return error when cluster is already provisioned", func(t *testing.T) {
		//given
		runtimeID := "0ad91f16-d553-413f-aa27-4eefd9e5f1c6"
		runtimeServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{}, nil)

		resolver := NewResolver(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		_, err := resolver.ProvisionRuntime(ctx, runtimeID, &gqlschema.ProvisionRuntimeInput{clusterConfig, &gqlschema.CredentialsInput{}, &gqlschema.KymaConfigInput{}})

		//then
		require.Error(t, err)
	})
}

func TestResolver_DeprovisionRuntime(t *testing.T) {
	hydroformMock := mocks.Client{}
	operationServiceMock := persistenceMocks.OperationService{}
	runtimeServiceMock := persistenceMocks.RuntimeService{}

	ctx := context.Background()

	t.Run("Should start runtime deprovisioning and return operation ID", func(t *testing.T) {
		//given
		lastOperation := model.Operation{State: model.Succeeded}
		runtimeStatus := model.RuntimeStatus{LastOperationStatus: lastOperation}

		runtimeID := "92a1c394-639a-424e-8578-ba1ca7501dc1"
		expOperationID := "c7241d2d-5ffd-434b-9a52-17ce9ee04578"
		operation := model.Operation{OperationID: expOperationID}

		runtimeServiceMock.On("GetStatus", runtimeID).Return(runtimeStatus, nil)
		runtimeServiceMock.On("SetDeprovisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		operationServiceMock.On("SetAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("DeprovisionCluster", mock.Anything, mock.Anything).Return(types.ClusterStatus{}, nil)

		resolver := NewResolver(operationServiceMock, runtimeServiceMock, hydroformMock)

		//when
		opt, err := resolver.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.NoError(t, err)
		assert.Equal(t, expOperationID, opt)
	})
}
