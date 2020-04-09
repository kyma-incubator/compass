package provisioning

import (
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	uuidMocks "github.com/kyma-incubator/compass/components/provisioner/internal/uuid/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

	mocks2 "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/mocks"
	sessionMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"

	releaseMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/release/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	directormock "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	kubeconfigFile = "kubeconfig data"
	runtimeID      = "184ccdf2-59e4-44b7-b553-6cb296af5ea0"
	operationID    = "223949ed-e6b6-4ab2-ab3e-8e19cd456dd40"
	runtimeName    = "test runtime"

	tenant       = "tenant"
	subAccountId = "sub-account"
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

	inputConverter := NewInputConverter(uuid.NewUUIDGenerator(), releaseRepo, gardenerProject)
	graphQLConverter := NewGraphQLConverter()
	uuidGenerator := uuid.NewUUIDGenerator()

	clusterConfig := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion: "1.16",
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				GcpConfig: &gqlschema.GCPProviderConfigInput{},
			},
		},
	}

	expectedCluster := model.Cluster{
		ID:         runtimeID,
		KymaConfig: fixKymaConfig(),
	}
	expectedOperation := model.Operation{
		ClusterID: runtimeID,
		State:     model.InProgress,
		Type:      model.Provision,
		Stage:     model.ShootProvisioning,
	}

	runtimeInput := &gqlschema.RuntimeInput{
		Name:        runtimeName,
		Description: new(string),
		Labels:      &gqlschema.Labels{},
	}

	provisionRuntimeInput := gqlschema.ProvisionRuntimeInput{
		RuntimeInput:  runtimeInput,
		ClusterConfig: clusterConfig,
		Credentials:   &gqlschema.CredentialsInput{},
		KymaConfig:    fixKymaGraphQLConfigInput(),
	}

	clusterMatcher := getClusterMatcher(expectedCluster)
	operationMatcher := getOperationMatcher(expectedOperation)

	t.Run("Should start runtime provisioning of Gardener cluster and return operation ID", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		directorServiceMock := &directormock.DirectorClient{}
		provisioner := &mocks2.Provisioner{}

		directorServiceMock.On("CreateRuntime", mock.Anything, tenant).Return(runtimeID, nil)
		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)
		writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGardenerConfig", mock.AnythingOfType("model.GardenerConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", mock.AnythingOfType("model.KymaConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("Commit").Return(nil)
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()
		provisioner.On("ProvisionCluster", mock.MatchedBy(clusterMatcher), mock.MatchedBy(notEmptyUUIDMatcher)).Return(nil)

		service := NewProvisioningService(inputConverter, graphQLConverter, directorServiceMock, sessionFactoryMock, provisioner, uuidGenerator, nil)

		//when
		operationStatus, err := service.ProvisionRuntime(provisionRuntimeInput, tenant, subAccountId)
		require.NoError(t, err)

		//then
		assert.Equal(t, runtimeID, *operationStatus.RuntimeID)
		assert.NotEmpty(t, operationStatus.ID)
		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
		provisioner.AssertExpectations(t)
		releaseRepo.AssertExpectations(t)
	})

	t.Run("Should start runtime provisioning of GCP cluster and return operation ID", func(t *testing.T) {
		//given
		provisionRuntimeInput := gqlschema.ProvisionRuntimeInput{
			RuntimeInput: runtimeInput,
			ClusterConfig: &gqlschema.ClusterConfigInput{
				GcpConfig: &gqlschema.GCPConfigInput{
					Name:        "cluster",
					ProjectName: "project",
				},
			},
			Credentials: &gqlschema.CredentialsInput{},
			KymaConfig:  fixKymaGraphQLConfigInput(),
		}

		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		directorServiceMock := &directormock.DirectorClient{}
		provisioner := &mocks2.Provisioner{}

		directorServiceMock.On("CreateRuntime", mock.Anything, tenant).Return(runtimeID, nil)
		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)
		writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGCPConfig", mock.AnythingOfType("model.GCPConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", mock.AnythingOfType("model.KymaConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("Commit").Return(nil)
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()
		provisioner.On("ProvisionCluster", mock.MatchedBy(clusterMatcher), mock.MatchedBy(notEmptyUUIDMatcher)).Return(nil)

		service := NewProvisioningService(inputConverter, graphQLConverter, directorServiceMock, sessionFactoryMock, provisioner, uuidGenerator, nil)

		//when
		operationStatus, err := service.ProvisionRuntime(provisionRuntimeInput, tenant, subAccountId)
		require.NoError(t, err)

		//then
		assert.Equal(t, runtimeID, *operationStatus.RuntimeID)
		assert.NotEmpty(t, operationStatus.ID)
		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
	})

	t.Run("Should return error and unregister Runtime when failed to commit transaction", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		directorServiceMock := &directormock.DirectorClient{}
		provisioner := &mocks2.Provisioner{}

		directorServiceMock.On("CreateRuntime", mock.Anything, tenant).Return(runtimeID, nil)
		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)
		writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGardenerConfig", mock.AnythingOfType("model.GardenerConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", mock.AnythingOfType("model.KymaConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("Commit").Return(dberrors.Internal("error"))
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()
		provisioner.On("ProvisionCluster", mock.MatchedBy(clusterMatcher), mock.MatchedBy(notEmptyUUIDMatcher)).Return(nil)
		directorServiceMock.On("DeleteRuntime", runtimeID, tenant).Return(nil)

		service := NewProvisioningService(inputConverter, graphQLConverter, directorServiceMock, sessionFactoryMock, provisioner, uuidGenerator, nil)

		//when
		_, err := service.ProvisionRuntime(provisionRuntimeInput, tenant, subAccountId)
		require.Error(t, err)

		//then
		assert.Contains(t, err.Error(), "Failed to commit transaction")
		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
		provisioner.AssertExpectations(t)
		releaseRepo.AssertExpectations(t)
	})

	t.Run("Should return error and unregister Runtime when failed to start provisioning", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		directorServiceMock := &directormock.DirectorClient{}
		provisioner := &mocks2.Provisioner{}

		directorServiceMock.On("CreateRuntime", mock.Anything, tenant).Return(runtimeID, nil)
		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)
		writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGardenerConfig", mock.AnythingOfType("model.GardenerConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", mock.AnythingOfType("model.KymaConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()
		provisioner.On("ProvisionCluster", mock.MatchedBy(clusterMatcher), mock.MatchedBy(notEmptyUUIDMatcher)).Return(errors.New("error"))
		directorServiceMock.On("DeleteRuntime", runtimeID, tenant).Return(nil)

		service := NewProvisioningService(inputConverter, graphQLConverter, directorServiceMock, sessionFactoryMock, provisioner, uuidGenerator, nil)

		//when
		_, err := service.ProvisionRuntime(provisionRuntimeInput, tenant, subAccountId)
		require.Error(t, err)

		//then
		assert.Contains(t, err.Error(), "Failed to start provisioning")
		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		directorServiceMock.AssertExpectations(t)
		provisioner.AssertExpectations(t)
		releaseRepo.AssertExpectations(t)
	})

	t.Run("Should return error when failed to register Runtime", func(t *testing.T) {
		//given
		directorServiceMock := &directormock.DirectorClient{}

		directorServiceMock.On("CreateRuntime", mock.Anything, tenant).Return("", errors.New("error"))

		service := NewProvisioningService(inputConverter, graphQLConverter, directorServiceMock, nil, nil, uuidGenerator, nil)

		//when
		_, err := service.ProvisionRuntime(provisionRuntimeInput, tenant, subAccountId)
		require.Error(t, err)

		//then
		assert.Contains(t, err.Error(), "Failed to register Runtime")
		directorServiceMock.AssertExpectations(t)
	})

}

func TestService_DeprovisionRuntime(t *testing.T) {

	inputConverter := NewInputConverter(uuid.NewUUIDGenerator(), nil, gardenerProject)
	graphQLConverter := NewGraphQLConverter()
	lastOperation := model.Operation{State: model.Succeeded}

	operation := model.Operation{
		ID:             operationID,
		Type:           model.Deprovision,
		State:          model.InProgress,
		StartTimestamp: time.Now(),
		Message:        "Deprovisioning started",
		ClusterID:      runtimeID,
	}

	cluster := model.Cluster{
		ID: runtimeID,
	}

	clusterMatcher := getClusterMatcher(cluster)
	operationMatcher := getOperationMatcher(operation)

	t.Run("Should start Runtime deprovisioning and return operation ID", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readWriteSession := &sessionMocks.ReadWriteSession{}
		provisioner := &mocks2.Provisioner{}

		sessionFactoryMock.On("NewReadWriteSession").Return(readWriteSession)
		readWriteSession.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		readWriteSession.On("GetCluster", runtimeID).Return(cluster, nil)
		provisioner.On("DeprovisionCluster", mock.MatchedBy(clusterMatcher), mock.MatchedBy(notEmptyUUIDMatcher)).Return(operation, nil)
		readWriteSession.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, provisioner, uuid.NewUUIDGenerator(), nil)

		//when
		opId, err := resolver.DeprovisionRuntime(runtimeID, tenant)
		require.NoError(t, err)

		//then
		assert.Equal(t, operationID, opId)
		sessionFactoryMock.AssertExpectations(t)
		readWriteSession.AssertExpectations(t)
		provisioner.AssertExpectations(t)
	})

	t.Run("Should return error when failed to start deprovisioning", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readWriteSession := &sessionMocks.ReadWriteSession{}
		provisioner := &mocks2.Provisioner{}

		sessionFactoryMock.On("NewReadWriteSession").Return(readWriteSession)
		readWriteSession.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		readWriteSession.On("GetCluster", runtimeID).Return(cluster, nil)
		provisioner.On("DeprovisionCluster", mock.MatchedBy(clusterMatcher), mock.MatchedBy(notEmptyUUIDMatcher)).Return(model.Operation{}, errors.New("error"))

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, provisioner, uuid.NewUUIDGenerator(), nil)

		//when
		_, err := resolver.DeprovisionRuntime(runtimeID, tenant)
		require.Error(t, err)

		//then
		assert.Contains(t, err.Error(), "Failed to start deprovisioning")
		sessionFactoryMock.AssertExpectations(t)
		readWriteSession.AssertExpectations(t)
		provisioner.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get cluster", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readWriteSession := &sessionMocks.ReadWriteSession{}

		sessionFactoryMock.On("NewReadWriteSession").Return(readWriteSession)
		readWriteSession.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		readWriteSession.On("GetCluster", runtimeID).Return(model.Cluster{}, dberrors.Internal("error"))

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuid.NewUUIDGenerator(), nil)

		//when
		_, err := resolver.DeprovisionRuntime(runtimeID, tenant)
		require.Error(t, err)

		//then
		assert.Contains(t, err.Error(), "Failed to get cluster")
		sessionFactoryMock.AssertExpectations(t)
		readWriteSession.AssertExpectations(t)
	})

	t.Run("Should return error when last operation in progress", func(t *testing.T) {
		//given
		operation := model.Operation{State: model.InProgress}

		sessionFactoryMock := &sessionMocks.Factory{}
		readWriteSession := &sessionMocks.ReadWriteSession{}

		sessionFactoryMock.On("NewReadWriteSession").Return(readWriteSession)
		readWriteSession.On("GetLastOperation", runtimeID).Return(operation, nil)

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuid.NewUUIDGenerator(), nil)

		//when
		_, err := resolver.DeprovisionRuntime(runtimeID, tenant)
		require.Error(t, err)

		//then
		assert.Contains(t, err.Error(), "previous one is in progress")
		sessionFactoryMock.AssertExpectations(t)
		readWriteSession.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get last operation", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readWriteSession := &sessionMocks.ReadWriteSession{}

		sessionFactoryMock.On("NewReadWriteSession").Return(readWriteSession)
		readWriteSession.On("GetLastOperation", runtimeID).Return(model.Operation{}, dberrors.Internal("error"))

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuid.NewUUIDGenerator(), nil)

		//when
		_, err := resolver.DeprovisionRuntime(runtimeID, tenant)
		require.Error(t, err)

		//then
		assert.Contains(t, err.Error(), "Failed to get last operation")
		sessionFactoryMock.AssertExpectations(t)
		readWriteSession.AssertExpectations(t)
	})
}

func TestService_RuntimeOperationStatus(t *testing.T) {
	uuidGenerator := &uuidMocks.UUIDGenerator{}
	inputConverter := NewInputConverter(uuidGenerator, nil, gardenerProject)
	graphQLConverter := NewGraphQLConverter()

	operation := model.Operation{
		ID:        operationID,
		Type:      model.Provision,
		State:     model.InProgress,
		Message:   "Message",
		ClusterID: runtimeID,
	}

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSession := &sessionMocks.ReadSession{}

		sessionFactoryMock.On("NewReadSession").Return(readSession)
		readSession.On("GetOperation", operationID).Return(operation, nil)

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, nil)

		//when
		status, err := resolver.RuntimeOperationStatus(operationID)
		//then
		require.NoError(t, err)
		assert.Equal(t, gqlschema.OperationTypeProvision, status.Operation)
		assert.Equal(t, gqlschema.OperationStateInProgress, status.State)
		assert.Equal(t, operation.ClusterID, *status.RuntimeID)
		assert.Equal(t, operation.ID, *status.ID)
		assert.Equal(t, operation.Message, *status.Message)
		sessionFactoryMock.AssertExpectations(t)
		readSession.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get operation status", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSession := &sessionMocks.ReadSession{}

		sessionFactoryMock.On("NewReadSession").Return(readSession)
		readSession.On("GetOperation", operationID).Return(model.Operation{}, dberrors.Internal("error"))

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, nil)

		//when
		_, err := resolver.RuntimeOperationStatus(operationID)

		//then
		require.Error(t, err)
		sessionFactoryMock.AssertExpectations(t)
		readSession.AssertExpectations(t)
	})
}

func TestService_RuntimeStatus(t *testing.T) {
	uuidGenerator := &uuidMocks.UUIDGenerator{}
	inputConverter := NewInputConverter(uuidGenerator, nil, gardenerProject)
	graphQLConverter := NewGraphQLConverter()

	operation := model.Operation{
		ID:        operationID,
		Type:      model.Provision,
		State:     model.Succeeded,
		Message:   "Message",
		ClusterID: runtimeID,
	}

	cluster := model.Cluster{
		ID:         runtimeID,
		Kubeconfig: util.StringPtr("kubeconfig"),
	}

	t.Run("Should return runtime status", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSession := &sessionMocks.ReadSession{}

		sessionFactoryMock.On("NewReadSession").Return(readSession)
		readSession.On("GetLastOperation", operationID).Return(operation, nil)
		readSession.On("GetCluster", operationID).Return(cluster, nil)

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, nil)

		//when
		status, err := resolver.RuntimeStatus(operationID)

		//then
		require.NoError(t, err)
		assert.Equal(t, cluster.ID, *status.LastOperationStatus.RuntimeID)
		assert.Equal(t, cluster.Kubeconfig, status.RuntimeConfiguration.Kubeconfig)
		sessionFactoryMock.AssertExpectations(t)
		readSession.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get cluster", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSession := &sessionMocks.ReadSession{}

		sessionFactoryMock.On("NewReadSession").Return(readSession)
		readSession.On("GetLastOperation", operationID).Return(operation, nil)
		readSession.On("GetCluster", operationID).Return(model.Cluster{}, dberrors.Internal("error"))

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, nil)

		//when
		_, err := resolver.RuntimeStatus(operationID)

		//then
		require.Error(t, err)
		sessionFactoryMock.AssertExpectations(t)
		readSession.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get operation status", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSession := &sessionMocks.ReadSession{}

		sessionFactoryMock.On("NewReadSession").Return(readSession)
		readSession.On("GetLastOperation", operationID).Return(model.Operation{}, dberrors.Internal("error"))

		resolver := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, nil)

		//when
		_, err := resolver.RuntimeStatus(operationID)

		//then
		require.Error(t, err)
		sessionFactoryMock.AssertExpectations(t)
		readSession.AssertExpectations(t)
	})
}

func TestService_UpgradeRuntime(t *testing.T) {
	releaseRepo := &releaseMocks.Repository{}
	releaseRepo.On("GetReleaseByVersion", kymaVersion).Return(kymaRelease, nil)
	inputConverter := NewInputConverter(uuid.NewUUIDGenerator(), releaseRepo, gardenerProject)
	graphQLConverter := NewGraphQLConverter()
	uuidGenerator := uuid.NewUUIDGenerator()

	lastOperation := model.Operation{State: model.Succeeded}

	oldKymaConfigId := "old-kyma-config-id"

	cluster := model.Cluster{
		ID: runtimeID,
		KymaConfig: model.KymaConfig{
			ID: oldKymaConfigId,
		},
	}

	expectedOperation := model.Operation{
		ClusterID: runtimeID,
		State:     model.InProgress,
		Type:      model.Upgrade,
		Stage:     model.StartingUpgrade,
	}

	runtimeUpgradeMatcher := func(rUp model.RuntimeUpgrade) bool {
		return rUp.OperationId != "" && rUp.PreUpgradeKymaConfigId == oldKymaConfigId && rUp.PostUpgradeKymaConfigId != oldKymaConfigId
	}

	upgradeInput := gqlschema.UpgradeRuntimeInput{
		KymaConfig: fixKymaGraphQLConfigInput(),
	}

	operationMatcher := getOperationMatcher(expectedOperation)

	t.Run("Should start runtime provisioning of Gardener cluster and return operation ID", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		readSessionMock := &sessionMocks.ReadSession{}
		upgradeQueue := &mocks.OperationQueue{}

		sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)
		readSessionMock.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		readSessionMock.On("GetCluster", runtimeID).Return(cluster, nil)
		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", mock.AnythingOfType("model.KymaConfig")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertRuntimeUpgrade", mock.MatchedBy(runtimeUpgradeMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("SetActiveKymaConfig", runtimeID, mock.AnythingOfType("string")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("Commit").Return(nil)
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()
		upgradeQueue.On("Add", mock.AnythingOfType("string")).Return(nil)

		service := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, upgradeQueue)

		//when
		operationStatus, err := service.UpgradeRuntime(runtimeID, upgradeInput)
		require.NoError(t, err)

		//then
		assert.Equal(t, runtimeID, *operationStatus.RuntimeID)
		assert.NotEmpty(t, operationStatus.ID)
		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		readSessionMock.AssertExpectations(t)
		upgradeQueue.AssertExpectations(t)
		releaseRepo.AssertExpectations(t)
	})

	for _, testCase := range []struct {
		description string
		mockFunc    func(sessionFactory *sessionMocks.Factory, writeSession *sessionMocks.WriteSessionWithinTransaction, readSession *sessionMocks.ReadSession)
	}{
		{
			description: "should fail to upgrade Runtime when failed to commit transaction",
			mockFunc: func(sessionFactory *sessionMocks.Factory, writeSession *sessionMocks.WriteSessionWithinTransaction, readSession *sessionMocks.ReadSession) {
				sessionFactory.On("NewReadSession").Return(readSession, nil)
				readSession.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
				readSession.On("GetCluster", runtimeID).Return(cluster, nil)
				sessionFactory.On("NewSessionWithinTransaction").Return(writeSession, nil)
				writeSession.On("InsertKymaConfig", mock.AnythingOfType("model.KymaConfig")).Return(nil)
				writeSession.On("InsertRuntimeUpgrade", mock.MatchedBy(runtimeUpgradeMatcher)).Return(nil)
				writeSession.On("SetActiveKymaConfig", runtimeID, mock.AnythingOfType("string")).Return(nil)
				writeSession.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)
				writeSession.On("Commit").Return(dberrors.Internal("error"))
				writeSession.On("RollbackUnlessCommitted").Return()
			},
		},
		{
			description: "should fail to upgrade Runtime when failed to insert new Kyma Config",
			mockFunc: func(sessionFactory *sessionMocks.Factory, writeSession *sessionMocks.WriteSessionWithinTransaction, readSession *sessionMocks.ReadSession) {
				sessionFactory.On("NewReadSession").Return(readSession, nil)
				readSession.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
				readSession.On("GetCluster", runtimeID).Return(cluster, nil)
				sessionFactory.On("NewSessionWithinTransaction").Return(writeSession, nil)
				writeSession.On("InsertKymaConfig", mock.AnythingOfType("model.KymaConfig")).Return(dberrors.Internal("error"))
				writeSession.On("RollbackUnlessCommitted").Return()
			},
		},
		{
			description: "should fail to upgrade Runtime when last operation is in progress",
			mockFunc: func(sessionFactory *sessionMocks.Factory, writeSession *sessionMocks.WriteSessionWithinTransaction, readSession *sessionMocks.ReadSession) {
				sessionFactory.On("NewReadSession").Return(readSession, nil)
				readSession.On("GetLastOperation", runtimeID).Return(model.Operation{State: model.InProgress}, nil)
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			//given
			sessionFactoryMock := &sessionMocks.Factory{}
			writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
			readSessionMock := &sessionMocks.ReadSession{}

			testCase.mockFunc(sessionFactoryMock, writeSessionWithinTransactionMock, readSessionMock)

			service := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, nil)

			//when
			_, err := service.UpgradeRuntime(runtimeID, upgradeInput)
			require.Error(t, err)

			// then
			sessionFactoryMock.AssertExpectations(t)
			writeSessionWithinTransactionMock.AssertExpectations(t)
			readSessionMock.AssertExpectations(t)
			releaseRepo.AssertExpectations(t)
		})
	}
}

func TestService_RollBackLastUpgrade(t *testing.T) {
	releaseRepo := &releaseMocks.Repository{}
	inputConverter := NewInputConverter(uuid.NewUUIDGenerator(), releaseRepo, gardenerProject)
	graphQLConverter := NewGraphQLConverter()
	uuidGenerator := uuid.NewUUIDGenerator()

	lastOperation := model.Operation{ID: operationID, State: model.Succeeded, Type: model.Upgrade}

	oldKymaConfigId := "old-kyma-config-id"

	runtimeUpgrade := model.RuntimeUpgrade{
		State:                   model.UpgradeSucceeded,
		OperationId:             operationID,
		PreUpgradeKymaConfigId:  oldKymaConfigId,
		PostUpgradeKymaConfigId: "new-id",
	}

	cluster := model.Cluster{
		ID: runtimeID,
		KymaConfig: model.KymaConfig{
			ID: oldKymaConfigId,
		},
	}

	t.Run("Should start runtime provisioning of Gardener cluster and return operation ID", func(t *testing.T) {
		//given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		readSessionMock := &sessionMocks.ReadSession{}

		sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)
		readSessionMock.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
		readSessionMock.On("GetRuntimeUpgrade", operationID).Return(runtimeUpgrade, nil)
		readSessionMock.On("GetCluster", runtimeID).Return(cluster, nil)
		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)
		writeSessionWithinTransactionMock.On("SetActiveKymaConfig", runtimeID, oldKymaConfigId).Return(nil)
		writeSessionWithinTransactionMock.On("UpdateUpgradeState", operationID, model.UpgradeRolledBack).Return(nil)
		writeSessionWithinTransactionMock.On("Commit").Return(nil)
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()

		service := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, nil)

		//when
		runtimeStatus, err := service.RollBackLastUpgrade(runtimeID)
		require.NoError(t, err)

		//then
		assert.NotEmpty(t, runtimeStatus)
		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		readSessionMock.AssertExpectations(t)
	})

	for _, testCase := range []struct {
		description string
		mockFunc    func(sessionFactory *sessionMocks.Factory, writeSession *sessionMocks.WriteSessionWithinTransaction, readSession *sessionMocks.ReadSession)
	}{
		{
			description: "should fail to roll back upgrade when failed to commit transaction",
			mockFunc: func(sessionFactory *sessionMocks.Factory, writeSession *sessionMocks.WriteSessionWithinTransaction, readSession *sessionMocks.ReadSession) {
				sessionFactory.On("NewReadSession").Return(readSession, nil)
				readSession.On("GetLastOperation", runtimeID).Return(lastOperation, nil)
				readSession.On("GetRuntimeUpgrade", operationID).Return(runtimeUpgrade, nil)
				sessionFactory.On("NewSessionWithinTransaction").Return(writeSession, nil)
				writeSession.On("SetActiveKymaConfig", runtimeID, oldKymaConfigId).Return(nil)
				writeSession.On("UpdateUpgradeState", operationID, model.UpgradeRolledBack).Return(nil)
				writeSession.On("Commit").Return(dberrors.Internal("error"))
				writeSession.On("RollbackUnlessCommitted").Return()
			},
		},
		{
			description: "should fail to roll back upgrade when failed to get last operation",
			mockFunc: func(sessionFactory *sessionMocks.Factory, writeSession *sessionMocks.WriteSessionWithinTransaction, readSession *sessionMocks.ReadSession) {
				sessionFactory.On("NewReadSession").Return(readSession, nil)
				readSession.On("GetLastOperation", runtimeID).Return(model.Operation{}, dberrors.Internal("error"))
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			//given
			sessionFactoryMock := &sessionMocks.Factory{}
			writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
			readSessionMock := &sessionMocks.ReadSession{}

			testCase.mockFunc(sessionFactoryMock, writeSessionWithinTransactionMock, readSessionMock)

			service := NewProvisioningService(inputConverter, graphQLConverter, nil, sessionFactoryMock, nil, uuidGenerator, nil)

			//when
			_, err := service.RollBackLastUpgrade(runtimeID)
			require.Error(t, err)

			//then
			sessionFactoryMock.AssertExpectations(t)
			writeSessionWithinTransactionMock.AssertExpectations(t)
			readSessionMock.AssertExpectations(t)
		})
	}
}

func getOperationMatcher(expected model.Operation) func(model.Operation) bool {
	return func(op model.Operation) bool {
		return op.Type == expected.Type && op.ClusterID == expected.ClusterID &&
			op.State == expected.State && op.Stage == expected.Stage
	}
}

func getClusterMatcher(expected model.Cluster) func(model.Cluster) bool {
	return func(cluster model.Cluster) bool {
		return cluster.ID == expected.ID
	}
}

func notEmptyUUIDMatcher(id string) bool {
	return len(id) > 0
}
