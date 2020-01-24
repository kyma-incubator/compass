package persistence

import (
	"testing"
	"time"

	uuidMocks "github.com/kyma-incubator/compass/components/provisioner/internal/uuid/mocks"

	sessionMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetProvisioning(t *testing.T) {

	runtimeID := "runtimeId"
	runtimeName := "runtimeName"

	gcpConfig := model.GCPConfig{
		Name:              "name",
		ProjectName:       "projectName",
		KubernetesVersion: "1.15",
		NumberOfNodes:     3,
		BootDiskSizeGB:    1024,
		MachineType:       "big_one",
		Region:            "region",
		Zone:              "zone",
	}

	gardenerConfig := model.GardenerConfig{
		Name:              "uuid",
		ProjectName:       "projectName",
		KubernetesVersion: "1.15",
		NodeCount:         3,
		VolumeSizeGB:      1024,
		DiskType:          "SSD",
		MachineType:       "big_one",
		Provider:          "GCP",
		Seed:              "gcp-eu1",
		TargetSecret:      "secret",
		WorkerCidr:        "cidr",
		Region:            "region",
		AutoScalerMin:     1,
		AutoScalerMax:     10,
		MaxSurge:          2,
		MaxUnavailable:    2,
		GardenerProviderConfig: &model.GCPGardenerConfig{
			ProviderSpecificConfig: "{\"zone\":\"zone\"}",
		},
	}

	kymaConfig := model.KymaConfig{
		Release: model.Release{
			Id:            "releaseID",
			Version:       "1.7",
			TillerYAML:    "tiller: yaml",
			InstallerYAML: "installer: yaml",
		},
		Components: []model.KymaComponentConfig{
			{ID: "id1", Component: model.KymaComponent("core")},
			{ID: "id2", Component: model.KymaComponent("monitoring")},
		},
	}

	runtimeGCPConfig := model.Cluster{
		ID:            runtimeID,
		RuntimeName:   runtimeName,
		KymaConfig:    kymaConfig,
		ClusterConfig: gcpConfig,
	}

	runtimeGardenerConfig := model.Cluster{
		ID:            runtimeID,
		RuntimeName:   runtimeName,
		KymaConfig:    kymaConfig,
		ClusterConfig: gardenerConfig,
	}

	timestamp := time.Now()

	operationID := "OperationID"

	operation := model.Operation{
		ID:             operationID,
		Type:           model.Provision,
		StartTimestamp: timestamp,
		State:          model.InProgress,
		Message:        "Provisioning started",
		ClusterID:      runtimeID,
	}

	operationMatcher := getOperationMatcher(operation)

	runtimeConfigurations := []struct {
		config                        model.Cluster
		description                   string
		insertClusterConfigMethodName string
	}{
		{config: runtimeGardenerConfig, insertClusterConfigMethodName: "InsertGardenerConfig", description: "Should set provisioning on Gardener started"},
		{config: runtimeGCPConfig, insertClusterConfigMethodName: "InsertGCPConfig", description: "Should set provisioning on GCP started"},
	}

	cluster := model.Cluster{
		ID:                runtimeID,
		RuntimeName:       runtimeName,
		TerraformState:    []byte("state"),
		CreationTimestamp: timestamp,
	}

	clusterMatcher := getClusterMatcher(cluster)

	for _, cfg := range runtimeConfigurations {
		t.Run(cfg.description, func(t *testing.T) {
			// given
			sessionFactoryMock := &sessionMocks.Factory{}
			writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
			uuidGenerator := &uuidMocks.UUIDGenerator{}

			uuidGenerator.On("New").Return(operationID, nil)

			writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
			writeSessionWithinTransactionMock.On(cfg.insertClusterConfigMethodName, cfg.config.ClusterConfig).Return(nil)
			writeSessionWithinTransactionMock.On("InsertKymaConfig", kymaConfig).Return(nil)
			writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)
			writeSessionWithinTransactionMock.On("Commit").Return(nil)
			writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()

			sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

			runtimeService := NewService(sessionFactoryMock, uuidGenerator)

			// when
			provisioningOperation, err := runtimeService.SetProvisioningStarted(runtimeID, cfg.config)

			// then
			assert.NoError(t, err)
			assert.Equal(t, provisioningOperation.ID, operationID)
			assert.Equal(t, provisioningOperation.Type, model.Provision)
			assert.Equal(t, provisioningOperation.State, model.InProgress)
			assert.Equal(t, provisioningOperation.ClusterID, runtimeID)
			assert.NotEmpty(t, provisioningOperation.Message)

			sessionFactoryMock.AssertExpectations(t)
			writeSessionWithinTransactionMock.AssertExpectations(t)
			uuidGenerator.AssertExpectations(t)
		})
	}

	t.Run("Should rollback transaction when failed to insert record to Cluster table", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		uuidGenerator := &uuidMocks.UUIDGenerator{}

		writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(dberrors.Internal("some error"))
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewService(sessionFactoryMock, uuidGenerator)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})

	t.Run("Should rollback transaction when failed to insert record to KymaConfig or KymaComponentConfig table", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		uuidGenerator := &uuidMocks.UUIDGenerator{}

		writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGCPConfig", gcpConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", kymaConfig).Return(dberrors.Internal("some error"))
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewService(sessionFactoryMock, uuidGenerator)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})

	clusterConfigurations := []struct {
		config model.Cluster

		description                   string
		insertClusterConfigMethodName string
	}{
		{config: runtimeGardenerConfig, insertClusterConfigMethodName: "InsertGardenerConfig", description: "Should rollback transaction when failed to insert record to GardenerConfig table"},
		{config: runtimeGCPConfig, insertClusterConfigMethodName: "InsertGCPConfig", description: "Should rollback transaction when failed to insert record to GCPConfig table"},
	}

	for _, cfg := range clusterConfigurations {
		t.Run(cfg.description, func(t *testing.T) {
			// given
			sessionFactoryMock := &sessionMocks.Factory{}
			writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
			uuidGenerator := &uuidMocks.UUIDGenerator{}

			writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
			writeSessionWithinTransactionMock.On(cfg.insertClusterConfigMethodName, cfg.config.ClusterConfig).Return(dberrors.Internal("some error"))
			writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()

			sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

			runtimeService := NewService(sessionFactoryMock, uuidGenerator)

			// when
			_, err := runtimeService.SetProvisioningStarted(runtimeID, cfg.config)

			// then
			assert.Error(t, err)

			sessionFactoryMock.AssertExpectations(t)
			writeSessionWithinTransactionMock.AssertExpectations(t)
			uuidGenerator.AssertExpectations(t)
		})
	}

	t.Run("Should rollback transaction when failed to insert record to Type table", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		uuidGenerator := &uuidMocks.UUIDGenerator{}

		uuidGenerator.On("New").Return(operationID, nil)

		writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGCPConfig", gcpConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", kymaConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(dberrors.Internal("some error"))

		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewService(sessionFactoryMock, uuidGenerator)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})

	t.Run("Should return error when failed to commit transaction", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		uuidGenerator := &uuidMocks.UUIDGenerator{}

		uuidGenerator.On("New").Return(operationID, nil)

		writeSessionWithinTransactionMock.On("InsertCluster", mock.MatchedBy(clusterMatcher)).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGCPConfig", gcpConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", kymaConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)

		writeSessionWithinTransactionMock.On("Commit").Return(dberrors.Internal("some error"))
		writeSessionWithinTransactionMock.On("RollbackUnlessCommitted").Return()

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewService(sessionFactoryMock, uuidGenerator)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})
}

func TestSetDeprovisioning(t *testing.T) {
	runtimeID := "runtimeID"
	operationID := "operationID"

	operation := model.Operation{
		ID:             operationID,
		Type:           model.Deprovision,
		StartTimestamp: time.Now(),
		State:          model.InProgress,
		Message:        "Deprovisioning started.",
		ClusterID:      runtimeID,
	}

	operationMatcher := getOperationMatcher(operation)

	t.Run("Should set deprovisioning started", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		uuidGenerator := &uuidMocks.UUIDGenerator{}

		uuidGenerator.On("New").Return(operationID, nil)

		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)

		sessionFactoryMock.On("NewWriteSession").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewService(sessionFactoryMock, uuidGenerator)

		// when
		provisioningOperation, err := runtimeService.SetDeprovisioningStarted(runtimeID)

		// then
		assert.NoError(t, err)
		assert.Equal(t, provisioningOperation.ID, "operationID")
		assert.Equal(t, provisioningOperation.Type, model.Deprovision)
		assert.Equal(t, provisioningOperation.State, model.InProgress)
		assert.Equal(t, provisioningOperation.ClusterID, runtimeID)
		assert.NotEmpty(t, provisioningOperation.Message)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
		uuidGenerator.AssertExpectations(t)
	})
}

func TestSetUpgrade(t *testing.T) {

	runtimeID := "runtimeID"
	operationID := "operationID"

	operation := model.Operation{
		ID:             operationID,
		Type:           model.Upgrade,
		StartTimestamp: time.Now(),
		State:          model.InProgress,
		Message:        "Upgrade started.",
		ClusterID:      runtimeID,
	}

	operationMatcher := getOperationMatcher(operation)

	t.Run("Should set upgrade started", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}
		uuidGenerator := &uuidMocks.UUIDGenerator{}

		uuidGenerator.On("New").Return(operationID, nil)

		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return(nil)

		sessionFactoryMock.On("NewWriteSession").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewService(sessionFactoryMock, uuidGenerator)

		// when
		provisioningOperation, err := runtimeService.SetUpgradeStarted(runtimeID)

		// then
		assert.NoError(t, err)
		assert.Equal(t, provisioningOperation.ID, operationID)
		assert.Equal(t, provisioningOperation.Type, model.Upgrade)
		assert.Equal(t, provisioningOperation.State, model.InProgress)
		assert.Equal(t, provisioningOperation.ClusterID, runtimeID)
		assert.NotEmpty(t, provisioningOperation.Message)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
	})
}

func TestGetRuntimeStatus(t *testing.T) {

	runtimeID := "runtimeID"
	runtimeName := "runtimeName"
	operation := model.Operation{
		Type:           model.Provision,
		StartTimestamp: time.Now(),
		State:          model.InProgress,
		Message:        "Provisioning started",
		ClusterID:      runtimeID,
	}

	gcpConfig := model.GCPConfig{
		Name:        "name",
		ProjectName: "projectName",
	}

	kymaConfig := model.KymaConfig{
		Release: model.Release{
			Id:            "releaseID",
			Version:       "1.7",
			TillerYAML:    "tiller: yaml",
			InstallerYAML: "installer: yaml",
		},
		Components: []model.KymaComponentConfig{
			{ID: "id1", Component: model.KymaComponent("core")},
			{ID: "id2", Component: model.KymaComponent("monitoring")},
		},
	}

	cluster := model.Cluster{
		ID:             runtimeID,
		RuntimeName:    runtimeName,
		Kubeconfig:     nil,
		TerraformState: []byte("state"),
	}

	t.Run("Should get runtime status", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSessionMock := &sessionMocks.ReadSession{}
		uuidGenerator := &uuidMocks.UUIDGenerator{}

		readSessionMock.On("GetLastOperation", runtimeID).Return(operation, nil)
		readSessionMock.On("GetProviderConfig", runtimeID).Return(gcpConfig, nil)
		readSessionMock.On("GetKymaConfig", runtimeID).Return(kymaConfig, nil)
		readSessionMock.On("GetCluster", runtimeID).Return(cluster, nil)

		sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)

		expected := model.RuntimeStatus{
			LastOperationStatus: operation,
			RuntimeConfiguration: model.Cluster{
				ID:             runtimeID,
				RuntimeName:    runtimeName,
				ClusterConfig:  gcpConfig,
				KymaConfig:     kymaConfig,
				TerraformState: []byte("state"),
			},
		}

		// when
		runtimeService := NewService(sessionFactoryMock, uuidGenerator)
		runtimeStatus, err := runtimeService.GetRuntimeStatus(runtimeID)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expected, runtimeStatus)
	})

	for _, testCase := range []struct {
		description     string
		sessionMockFunc func(session *sessionMocks.ReadSession)
	}{
		{
			description: "should fail to get Runtime status when getting last operation failed",
			sessionMockFunc: func(readSessionMock *sessionMocks.ReadSession) {
				readSessionMock.On("GetLastOperation", runtimeID).Return(model.Operation{}, dberrors.Internal("some error"))
			},
		},
		{
			description: "should fail to get Runtime status when getting provider config failed",
			sessionMockFunc: func(readSessionMock *sessionMocks.ReadSession) {
				readSessionMock.On("GetLastOperation", runtimeID).Return(operation, nil)
				readSessionMock.On("GetProviderConfig", runtimeID).Return(model.GCPConfig{}, dberrors.Internal("some error"))
			},
		},
		{
			description: "should fail to get Runtime status when getting kyma config failed",
			sessionMockFunc: func(readSessionMock *sessionMocks.ReadSession) {
				readSessionMock.On("GetLastOperation", runtimeID).Return(operation, nil)
				readSessionMock.On("GetProviderConfig", runtimeID).Return(gcpConfig, nil)
				readSessionMock.On("GetKymaConfig", runtimeID).Return(model.KymaConfig{}, dberrors.Internal("some error"))
			},
		},
		{
			description: "should fail to get Runtime status when getting cluster config failed",
			sessionMockFunc: func(readSessionMock *sessionMocks.ReadSession) {
				readSessionMock.On("GetLastOperation", runtimeID).Return(operation, nil)
				readSessionMock.On("GetProviderConfig", runtimeID).Return(gcpConfig, nil)
				readSessionMock.On("GetKymaConfig", runtimeID).Return(kymaConfig, nil)
				readSessionMock.On("GetCluster", runtimeID).Return(model.Cluster{}, dberrors.Internal("some error"))
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			sessionFactoryMock := &sessionMocks.Factory{}
			readSessionMock := &sessionMocks.ReadSession{}
			uuidGenerator := &uuidMocks.UUIDGenerator{}

			testCase.sessionMockFunc(readSessionMock)

			sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)

			// when
			runtimeService := NewService(sessionFactoryMock, uuidGenerator)
			_, err := runtimeService.GetRuntimeStatus(runtimeID)

			// then
			assert.Error(t, err)
		})
	}
}

func TestGetClusterData(t *testing.T) {
	runtimeID := "runtimeID"
	runtimeName := "runtimeName"

	gcpConfig := model.GCPConfig{
		Name:        "name",
		ProjectName: "projectName",
	}

	kymaConfig := model.KymaConfig{
		Release: model.Release{
			Id:            "releaseID",
			Version:       "1.7",
			TillerYAML:    "tiller: yaml",
			InstallerYAML: "installer: yaml",
		},
		Components: []model.KymaComponentConfig{
			{ID: "id1", Component: model.KymaComponent("core")},
			{ID: "id2", Component: model.KymaComponent("monitoring")},
		},
	}

	cluster := model.Cluster{
		ID:             runtimeID,
		RuntimeName:    runtimeName,
		Kubeconfig:     nil,
		TerraformState: []byte("state"),
	}

	t.Run("Should get cluster data", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSessionMock := &sessionMocks.ReadSession{}
		uuidGenerator := &uuidMocks.UUIDGenerator{}

		readSessionMock.On("GetProviderConfig", runtimeID).Return(gcpConfig, nil)
		readSessionMock.On("GetKymaConfig", runtimeID).Return(kymaConfig, nil)
		readSessionMock.On("GetCluster", runtimeID).Return(cluster, nil)

		sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)

		expected := model.Cluster{
			ID:             runtimeID,
			RuntimeName:    runtimeName,
			Kubeconfig:     nil,
			TerraformState: []byte("state"),
			KymaConfig:     kymaConfig,
			ClusterConfig:  gcpConfig,
		}

		// when
		runtimeService := NewService(sessionFactoryMock, uuidGenerator)
		runtimeStatus, err := runtimeService.GetClusterData(runtimeID)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expected, runtimeStatus)
	})
}

func TestCleanupClusterData(t *testing.T) {

	t.Run("Should clean up cluster data", func(t *testing.T) {

		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionMock := &sessionMocks.WriteSession{}
		runtimeID := "runtimeID"

		sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock)
		writeSessionMock.On("DeleteCluster", runtimeID).Return(nil)

		// when
		runtimeService := NewService(sessionFactoryMock, nil)
		err := runtimeService.CleanupClusterData(runtimeID)

		// then
		assert.NoError(t, err)
	})
}

func getOperationMatcher(expected model.Operation) func(model.Operation) bool {
	return func(op model.Operation) bool {
		return op.Type == expected.Type &&
			op.Message == expected.Message && op.ClusterID == expected.ClusterID &&
			op.State == expected.State && op.ID == expected.ID
	}
}

func getClusterMatcher(expected model.Cluster) func(model.Cluster) bool {
	return func(cluster model.Cluster) bool {
		return cluster.ID == expected.ID &&
			string(cluster.TerraformState) == string(expected.TerraformState) && cluster.Kubeconfig == expected.Kubeconfig && cluster.RuntimeName == expected.RuntimeName
	}
}
