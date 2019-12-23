package provisioning

import (
	"errors"
	"testing"

	releaseMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/release/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/converters"

	installationMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"
	uuidMocks "github.com/kyma-incubator/compass/components/provisioner/internal/uuid/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"

	directormock "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	persistenceMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	kubeconfigFile = "kubeconfig data"
	kymaVersion    = "1.5"
)

var (
	kymaRelease = model.Release{
		Id:            "releaseId",
		Version:       kymaVersion,
		TillerYAML:    "tiller yaml",
		InstallerYAML: "installer yaml",
	}
)

func TestService_ProvisionRuntime(t *testing.T) {
	releaseRepo := &releaseMocks.Repository{}
	releaseRepo.On("GetReleaseByVersion", kymaVersion).Return(kymaRelease, nil)

	inputConverter := converters.NewInputConverter(uuid.NewUUIDGenerator(), releaseRepo)
	graphQLConverter := converters.NewGraphQLConverter()

	clusterConfig := &gqlschema.ClusterConfigInput{
		GcpConfig: &gqlschema.GCPConfigInput{
			Name:              "Something",
			ProjectName:       "Project",
			NumberOfNodes:     3,
			BootDiskSizeGb:    256,
			MachineType:       "machine",
			Region:            "region",
			Zone:              new(string),
			KubernetesVersion: "version",
		},
	}

	kymaConfig := &gqlschema.KymaConfigInput{
		Version: kymaVersion,
		Modules: gqlschema.AllKymaModule,
	}

	runtimeInput := &gqlschema.RuntimeInput{
		Name:        "test runtime",
		Description: new(string),
		Labels:      &gqlschema.Labels{},
	}

	provisionRuntimeInput := gqlschema.ProvisionRuntimeInput{
		RuntimeInput:  runtimeInput,
		ClusterConfig: clusterConfig,
		Credentials:   &gqlschema.CredentialsInput{},
		KymaConfig:    kymaConfig,
	}

	t.Run("Should start runtime provisioning and return operation ID", func(t *testing.T) {
		//given
		hydroformMock := &mocks.Service{}
		persistenceServiceMock := &persistenceMocks.Service{}
		installationSvc := &installationMocks.Service{}

		expRuntimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything).Return(expRuntimeID, nil)

		persistenceServiceMock.On("GetLastOperation", expRuntimeID).Return(model.Operation{}, dberrors.NotFound("Not found"))
		persistenceServiceMock.On("SetProvisioningStarted", expRuntimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("UpdateClusterData", expRuntimeID, kubeconfigFile, []byte("")).Return(nil)
		persistenceServiceMock.On("SetOperationAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")}, nil)
		installationSvc.On("InstallKyma", expRuntimeID, kubeconfigFile, kymaRelease).Return(nil)

		service := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, installationSvc, directorServiceMock)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput)
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expRuntimeID, runtimeID)
		assert.Equal(t, expOperationID, operationID)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		installationSvc.AssertExpectations(t)
		releaseRepo.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
	})

	t.Run("Should fail runtime provisioning after failing Registering Runtime in Director", func(t *testing.T) {
		//given
		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything).Return("", errors.New("Some error"))

		service := NewProvisioningService(nil, nil, nil, nil, nil, directorServiceMock)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput)

		// then
		assert.Error(t, err)
		assert.Nil(t, finished)
		assert.Empty(t, operationID)
		assert.Empty(t, runtimeID)
		directorServiceMock.AssertExpectations(t)
	})

	t.Run("Should start runtime provisioning and return operation ID when previous provisioning failed", func(t *testing.T) {
		//given
		hydroformMock := &mocks.Service{}
		persistenceServiceMock := &persistenceMocks.Service{}
		installationSvc := &installationMocks.Service{}

		expRuntimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything).Return(expRuntimeID, nil)
		directorServiceMock.On("DeleteRuntime", expRuntimeID).Return(nil)
		persistenceServiceMock.On("GetLastOperation", expRuntimeID).Return(model.Operation{Type: model.Provision, State: model.Failed}, nil)
		persistenceServiceMock.On("CleanupClusterData", expRuntimeID).Return(nil)
		persistenceServiceMock.On("SetProvisioningStarted", expRuntimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("UpdateClusterData", expRuntimeID, kubeconfigFile, []byte("")).Return(nil)
		persistenceServiceMock.On("SetOperationAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")}, nil)
		installationSvc.On("InstallKyma", expRuntimeID, kubeconfigFile, kymaRelease).Return(nil)

		service := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, installationSvc, directorServiceMock)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput)
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expRuntimeID, runtimeID)
		assert.Equal(t, expOperationID, operationID)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		installationSvc.AssertExpectations(t)
		releaseRepo.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
	})

	t.Run("Should return error when Kyma installation failed", func(t *testing.T) {
		//given
		hydroformMock := &mocks.Service{}
		persistenceServiceMock := &persistenceMocks.Service{}
		installationSvc := &installationMocks.Service{}

		expRuntimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything).Return(expRuntimeID, nil)
		persistenceServiceMock.On("GetLastOperation", expRuntimeID).Return(model.Operation{}, dberrors.NotFound("Not found"))
		persistenceServiceMock.On("SetProvisioningStarted", expRuntimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("UpdateClusterData", expRuntimeID, kubeconfigFile, []byte("")).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")}, nil)
		installationSvc.On("InstallKyma", expRuntimeID, kubeconfigFile, kymaRelease).Return(errors.New("error"))
		persistenceServiceMock.On("SetOperationAsFailed", expOperationID, mock.AnythingOfType("string")).Return(nil)

		service := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, installationSvc, directorServiceMock)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput)
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expOperationID, operationID)
		assert.Equal(t, expRuntimeID, runtimeID)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		installationSvc.AssertExpectations(t)
		releaseRepo.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
	})

	t.Run("Should return error when cluster is already provisioned", func(t *testing.T) {
		//given
		hydroformMock := &mocks.Service{}
		persistenceServiceMock := &persistenceMocks.Service{}

		expRuntimeID := "0ad91f16-d553-413f-aa27-4eefd9e5f1c6"
		persistenceServiceMock.On("GetLastOperation", expRuntimeID).Return(model.Operation{}, nil)

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything).Return(expRuntimeID, nil)

		service := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, nil, directorServiceMock)

		//when
		operationID, runtimeID, _, err := service.ProvisionRuntime(provisionRuntimeInput)

		//then
		require.Error(t, err)
		persistenceServiceMock.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
		assert.Empty(t, operationID)
		assert.Empty(t, runtimeID)
	})
}

func TestService_DeprovisionRuntime(t *testing.T) {
	persistenceServiceMock := &persistenceMocks.Service{}
	hydroformMock := &mocks.Service{}

	releaseRepo := &releaseMocks.Repository{}
	releaseRepo.On("GetReleaseByVersion", kymaVersion).Return(kymaRelease, nil)

	inputConverter := converters.NewInputConverter(uuid.NewUUIDGenerator(), releaseRepo)
	graphQLConverter := converters.NewGraphQLConverter()

	t.Run("Should start runtime deprovisioning and return operation ID", func(t *testing.T) {
		//given
		lastOperation := model.Operation{State: model.Succeeded}

		runtimeID := "92a1c394-639a-424e-8578-ba1ca7501dc1"
		expOperationID := "c7241d2d-5ffd-434b-9a52-17ce9ee04578"
		operation := model.Operation{ID: expOperationID}

		persistenceServiceMock.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		persistenceServiceMock.On("SetDeprovisioningStarted", runtimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("GetClusterData", runtimeID).Return(model.Cluster{TerraformState: []byte("{}"), ClusterConfig: model.GCPConfig{}}, nil)
		persistenceServiceMock.On("SetOperationAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("DeprovisionCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("DeleteRuntime", runtimeID).Return(nil)

		resolver := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, nil, directorServiceMock)

		//when
		opt, finished, err := resolver.DeprovisionRuntime(runtimeID)
		require.NoError(t, err)

		waitUntilFinished(finished)

		//then
		assert.Equal(t, expOperationID, opt)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
	})

	t.Run("Should not start deprovisioning when previous operation is in progress", func(t *testing.T) {
		//given
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"
		persistenceServiceMock := &persistenceMocks.Service{}

		persistenceServiceMock.On("GetLastOperation", runtimeID).Return(model.Operation{State: model.InProgress}, nil)

		resolver := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, nil, nil)

		//when
		_, _, err := resolver.DeprovisionRuntime(runtimeID)

		//then
		require.Error(t, err)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
	})

	t.Run("Should not start deprovisioning when Director fails to unregister Runtime", func(t *testing.T) {
		//given
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"
		persistenceServiceMock := &persistenceMocks.Service{}
		directorServiceMock := &directormock.DirectorClient{}

		lastOperation := model.Operation{State: model.Succeeded}

		persistenceServiceMock.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		persistenceServiceMock.On("GetClusterData", runtimeID).Return(model.Cluster{TerraformState: []byte("{}"), ClusterConfig: model.GCPConfig{}}, nil)
		directorServiceMock.On("DeleteRuntime", runtimeID).Return(errors.New("Some error!"))

		service := NewProvisioningService(persistenceServiceMock, nil, nil, nil, nil, directorServiceMock)

		//when
		_, _, err := service.DeprovisionRuntime(runtimeID)

		//then
		require.Error(t, err)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
	})
}

func TestService_RuntimeOperationStatus(t *testing.T) {
	persistenceServiceMock := &persistenceMocks.Service{}
	hydroformMock := &mocks.Service{}
	uuidGenerator := &uuidMocks.UUIDGenerator{}

	releaseRepo := &releaseMocks.Repository{}
	releaseRepo.On("GetReleaseByVersion", kymaVersion).Return(kymaRelease, nil)

	inputConverter := converters.NewInputConverter(uuidGenerator, releaseRepo)
	graphQLConverter := converters.NewGraphQLConverter()

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

		persistenceServiceMock.On("GetOperation", operationID).Return(operation, nil)
		resolver := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, nil, nil)

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

func TestCleanUpRuntimeData(t *testing.T) {
	t.Run("Should fail to get Clean Up Runtime Data result when Runtime ID not found in database", func(t *testing.T) {
		// given
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"

		persistenceServiceMock := &persistenceMocks.Service{}
		persistenceServiceMock.On("CleanupClusterData", runtimeID).Return(dberrors.NotFound("Could not find given Runtime in database"))

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("DeleteRuntime", runtimeID).Return(nil)

		provisioningService := NewProvisioningService(persistenceServiceMock, nil, nil, nil, nil, directorServiceMock)

		// when
		result, err := provisioningService.CleanupRuntimeData(runtimeID)

		// then
		require.NoError(t, err)
		assert.Equal(t, runtimeID, result.ID)
		assert.NotEmpty(t, result.Message)
		persistenceServiceMock.AssertExpectations(t)
	})

	t.Run("Should fail to get Clean Up Runtime Data result when internal database error occurs", func(t *testing.T) {
		// given
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"

		persistenceServiceMock := &persistenceMocks.Service{}
		persistenceServiceMock.On("CleanupClusterData", runtimeID).Return(dberrors.Internal("Internal database error occurred"))

		provisioningService := NewProvisioningService(persistenceServiceMock, nil, nil, nil, nil, nil)

		// when
		result, err := provisioningService.CleanupRuntimeData(runtimeID)

		// then
		require.Error(t, err)
		require.Nil(t, result)
		persistenceServiceMock.AssertExpectations(t)
	})

	t.Run("Should pass and return Runtime ID and Clean Up Runtime Data result when data for given Runtime gets deleted", func(t *testing.T) {
		// given
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"

		persistenceServiceMock := &persistenceMocks.Service{}
		persistenceServiceMock.On("CleanupClusterData", runtimeID).Return(nil)

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("DeleteRuntime", runtimeID).Return(nil)

		provisioningService := NewProvisioningService(persistenceServiceMock, nil, nil, nil, nil, directorServiceMock)

		// when
		result, err := provisioningService.CleanupRuntimeData(runtimeID)

		// then
		require.NoError(t, err)
		assert.Equal(t, runtimeID, result.ID)
		assert.NotEmpty(t, result.Message)
		persistenceServiceMock.AssertExpectations(t)
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
