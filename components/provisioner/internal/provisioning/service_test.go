package provisioning

import (
	"errors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	runtimeConfigMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/runtimes/mocks"
	"testing"

	releaseMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/release/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	installationMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"
	uuidMocks "github.com/kyma-incubator/compass/components/provisioner/internal/uuid/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"

	directormock "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	persistenceMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	kubeconfigFile = "kubeconfig data"

	tenant = "tenant"
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

	inputConverter := NewInputConverter(uuid.NewUUIDGenerator(), releaseRepo)
	graphQLConverter := NewGraphQLConverter()

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

	kymaConfigInput := fixKymaGraphQLConfigInput()
	expectedGlobalConfig := fixGlobalConfig()

	runtimeInput := &gqlschema.RuntimeInput{
		Name:        "test runtime",
		Description: new(string),
		Labels:      &gqlschema.Labels{},
	}

	provisionRuntimeInput := gqlschema.ProvisionRuntimeInput{
		RuntimeInput:  runtimeInput,
		ClusterConfig: clusterConfig,
		Credentials:   &gqlschema.CredentialsInput{},
		KymaConfig:    kymaConfigInput,
	}

	t.Run("Should start runtime provisioning and return operation ID", func(t *testing.T) {
		//given
		hydroformMock := &mocks.Service{}
		persistenceServiceMock := &persistenceMocks.Service{}
		installationSvc := &installationMocks.Service{}
		configProvider := &runtimeConfigMocks.ConfigProvider{}

		token := graphql.OneTimeToken{}
		expRuntimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything, mock.Anything).Return(expRuntimeID, nil)
		directorServiceMock.On("GetConnectionToken", mock.Anything, mock.Anything).Return(token, nil)

		persistenceServiceMock.On("SetProvisioningStarted", expRuntimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("UpdateClusterData", expRuntimeID, kubeconfigFile, []byte("")).Return(nil)
		persistenceServiceMock.On("SetOperationAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")}, nil)
		installationSvc.On("InstallKyma", expRuntimeID, kubeconfigFile, kymaRelease, expectedGlobalConfig, mock.AnythingOfType("[]model.KymaComponentConfig")).Return(nil)
		configProvider.On("CreateConfigMapForRuntime", mock.Anything, mock.Anything).Return(nil, nil)

		service := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, installationSvc, directorServiceMock, configProvider)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput, tenant)
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
		configProvider.AssertExpectations(t)
	})

	t.Run("Should fail runtime provisioning after failing Registering Runtime in Director", func(t *testing.T) {
		//given
		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything, mock.Anything).Return("", errors.New("Some error"))

		service := NewProvisioningService(nil, nil, nil, nil, nil, directorServiceMock, nil)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput, tenant)

		// then
		assert.Error(t, err)
		assert.Nil(t, finished)
		assert.Empty(t, operationID)
		assert.Empty(t, runtimeID)
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
		directorServiceMock.On("CreateRuntime", mock.Anything, mock.Anything).Return(expRuntimeID, nil)
		persistenceServiceMock.On("SetProvisioningStarted", expRuntimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("UpdateClusterData", expRuntimeID, kubeconfigFile, []byte("")).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")}, nil)
		installationSvc.On("InstallKyma", expRuntimeID, kubeconfigFile, kymaRelease, expectedGlobalConfig, mock.AnythingOfType("[]model.KymaComponentConfig")).Return(errors.New("error"))
		persistenceServiceMock.On("SetOperationAsFailed", expOperationID, mock.AnythingOfType("string")).Return(nil)

		service := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, installationSvc, directorServiceMock, nil)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput, tenant)
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

	t.Run("Should return error when fails to get one time token from Director", func(t *testing.T) {
		//given
		hydroformMock := &mocks.Service{}
		persistenceServiceMock := &persistenceMocks.Service{}
		installationSvc := &installationMocks.Service{}
		configProvider := &runtimeConfigMocks.ConfigProvider{}

		expRuntimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything, mock.Anything).Return(expRuntimeID, nil)
		directorServiceMock.On("GetConnectionToken", mock.Anything, mock.Anything).Return(graphql.OneTimeToken{}, errors.New("Token error"))

		persistenceServiceMock.On("SetProvisioningStarted", expRuntimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("UpdateClusterData", expRuntimeID, kubeconfigFile, []byte("")).Return(nil)
		persistenceServiceMock.On("SetOperationAsFailed", expOperationID, mock.AnythingOfType("string")).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")}, nil)
		installationSvc.On("InstallKyma", expRuntimeID, kubeconfigFile, kymaRelease, expectedGlobalConfig, mock.AnythingOfType("[]model.KymaComponentConfig")).Return(nil)

		service := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, installationSvc, directorServiceMock, configProvider)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput, tenant)
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

	t.Run("Should return error when fails to create config map for Runtime", func(t *testing.T) {
		//given
		hydroformMock := &mocks.Service{}
		persistenceServiceMock := &persistenceMocks.Service{}
		installationSvc := &installationMocks.Service{}
		configProvider := &runtimeConfigMocks.ConfigProvider{}

		expRuntimeID := "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
		expOperationID := "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
		operation := model.Operation{ID: expOperationID}

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("CreateRuntime", mock.Anything, mock.Anything).Return(expRuntimeID, nil)
		directorServiceMock.On("GetConnectionToken", mock.Anything, mock.Anything).Return(graphql.OneTimeToken{}, nil)

		persistenceServiceMock.On("SetProvisioningStarted", expRuntimeID, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("UpdateClusterData", expRuntimeID, kubeconfigFile, []byte("")).Return(nil)
		persistenceServiceMock.On("SetOperationAsFailed", expOperationID, mock.AnythingOfType("string")).Return(nil)
		hydroformMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")}, nil)
		installationSvc.On("InstallKyma", expRuntimeID, kubeconfigFile, kymaRelease, expectedGlobalConfig, mock.AnythingOfType("[]model.KymaComponentConfig")).Return(nil)

		configProvider.On("CreateConfigMapForRuntime", mock.Anything, mock.Anything).Return(nil, errors.New("ConfigMap error"))

		service := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, installationSvc, directorServiceMock, configProvider)

		//when
		operationID, runtimeID, finished, err := service.ProvisionRuntime(provisionRuntimeInput, tenant)
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
		configProvider.AssertExpectations(t)
	})

}

func TestService_DeprovisionRuntime(t *testing.T) {
	persistenceServiceMock := &persistenceMocks.Service{}
	hydroformMock := &mocks.Service{}

	releaseRepo := &releaseMocks.Repository{}
	releaseRepo.On("GetReleaseByVersion", kymaVersion).Return(kymaRelease, nil)

	inputConverter := NewInputConverter(uuid.NewUUIDGenerator(), releaseRepo)
	graphQLConverter := NewGraphQLConverter()

	t.Run("Should start Runtime deprovisioning and return operation ID", func(t *testing.T) {
		//given
		lastOperation := model.Operation{State: model.Succeeded}

		runtimeID := "92a1c394-639a-424e-8578-ba1ca7501dc1"
		expOperationID := "c7241d2d-5ffd-434b-9a52-17ce9ee04578"
		operation := model.Operation{ID: expOperationID}

		persistenceServiceMock.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		persistenceServiceMock.On("SetDeprovisioningStarted", runtimeID, mock.Anything, mock.Anything).Return(operation, nil)
		persistenceServiceMock.On("GetClusterData", runtimeID).Return(model.Cluster{ID: runtimeID, TerraformState: []byte("{}"), ClusterConfig: model.GCPConfig{}}, nil)
		persistenceServiceMock.On("SetOperationAsSucceeded", expOperationID).Return(nil)
		hydroformMock.On("DeprovisionCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		directorServiceMock := &directormock.DirectorClient{}
		directorServiceMock.On("DeleteRuntime", runtimeID, mock.Anything).Return(nil)

		resolver := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, nil, directorServiceMock, nil)

		//when
		opt, finished, err := resolver.DeprovisionRuntime(runtimeID, tenant)
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

		resolver := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, nil, nil, nil)

		//when
		_, _, err := resolver.DeprovisionRuntime(runtimeID, tenant)

		//then
		require.Error(t, err)
		hydroformMock.AssertExpectations(t)
		persistenceServiceMock.AssertExpectations(t)
	})

	t.Run("Deprovisioning should fail when Director fails to unregister Runtime", func(t *testing.T) {
		//given
		runtimeID := "a24142da-1111-4ec2-93e3-e47ccaa6973f"
		expOperationID := "c7241d2d-5ffd-434b-9a52-17ce9ee04578"
		operation := model.Operation{ID: expOperationID}

		persistenceServiceMock := &persistenceMocks.Service{}
		directorServiceMock := &directormock.DirectorClient{}

		lastOperation := model.Operation{State: model.Succeeded}

		persistenceServiceMock.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		persistenceServiceMock.On("GetClusterData", runtimeID).Return(model.Cluster{ID: runtimeID, TerraformState: []byte("{}"), ClusterConfig: model.GCPConfig{}}, nil)
		persistenceServiceMock.On("SetDeprovisioningStarted", runtimeID, mock.Anything, mock.Anything).Return(operation, nil)
		hydroformMock.On("DeprovisionCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		directorServiceMock.On("DeleteRuntime", runtimeID, mock.Anything).Return(errors.New("Some error!"))
		persistenceServiceMock.On("SetOperationAsFailed", expOperationID, "Some error!").Return(nil)

		service := NewProvisioningService(persistenceServiceMock, nil, nil, hydroformMock, nil, directorServiceMock, nil)

		//when
		_, doneChan, err := service.DeprovisionRuntime(runtimeID, tenant)
		require.NoError(t, err)

		_ = <-doneChan //

		//then
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

	inputConverter := NewInputConverter(uuidGenerator, releaseRepo)
	graphQLConverter := NewGraphQLConverter()

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
		resolver := NewProvisioningService(persistenceServiceMock, inputConverter, graphQLConverter, hydroformMock, nil, nil, nil)

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
