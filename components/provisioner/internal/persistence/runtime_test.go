package persistence

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	sessionMocks "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetProvisioning(t *testing.T) {

	runtimeID := "runtimeId"

	gcpConfig := model.GCPConfig{
		Name:              "name",
		ProjectName:       "projectName",
		KubernetesVersion: "1.15",
		NumberOfNodes:     3,
		BootDiskSize:      "1TB",
		MachineType:       "big_one",
		Region:            "region",
		Zone:              "zone",
	}

	gardenerConfig := model.GardenerConfig{
		Name:              "name",
		ProjectName:       "projectName",
		KubernetesVersion: "1.15",
		NodeCount:         3,
		VolumeSize:        "1TB",
		DiskType:          "SSD",
		MachineType:       "big_one",
		TargetProvider:    "GCP",
		TargetSecret:      "secret",
		Cidr:              "cidr",
		Region:            "region",
		Zone:              "zone",
		AutoScalerMin:     1,
		AutoScalerMax:     10,
		MaxSurge:          2,
		MaxUnavailable:    2,
	}

	kymaConfig := model.KymaConfig{
		Version: "1.6",
		Modules: []model.KymaConfigModule{
			{ID: "id1", Module: model.KymaModule("core")},
			{ID: "id2", Module: model.KymaModule("monitoring")},
		},
	}

	runtimeGCPConfig := model.RuntimeConfig{
		KymaConfig:    kymaConfig,
		ClusterConfig: gcpConfig,
	}

	runtimeGardenerConfig := model.RuntimeConfig{
		KymaConfig:    kymaConfig,
		ClusterConfig: gardenerConfig,
	}

	timestamp := time.Now()

	operation := model.Operation{
		Type:           model.Provision,
		StartTimestamp: timestamp,
		State:          model.InProgress,
		Message:        "Provisioning started",
		ClusterID:      runtimeID,
	}

	operationMatcher := getOperationMather(operation)

	runtimeConfigurations := []struct {
		config model.RuntimeConfig

		description                   string
		insertClusterConfigMethodName string
	}{
		{config: runtimeGardenerConfig, insertClusterConfigMethodName: "InsertGardenerConfig", description: "Should set provisioning on Gardener started"},
		{config: runtimeGCPConfig, insertClusterConfigMethodName: "InsertGCPConfig", description: "Should set provisioning on GCP started"},
	}

	for _, cfg := range runtimeConfigurations {
		t.Run(cfg.description, func(t *testing.T) {
			// given
			sessionFactoryMock := &sessionMocks.Factory{}
			writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

			terraformState := "{}"
			writeSessionWithinTransactionMock.On("InsertCluster", runtimeID, mock.AnythingOfType("Time"), terraformState).Return(nil)
			writeSessionWithinTransactionMock.On(cfg.insertClusterConfigMethodName, runtimeID, cfg.config.ClusterConfig).Return(nil)
			writeSessionWithinTransactionMock.On("InsertKymaConfig", runtimeID, kymaConfig.Version).Return("kymaID", nil)
			writeSessionWithinTransactionMock.On("InsertKymaConfigModule", "kymaID", model.KymaModule("core")).Return(nil)
			writeSessionWithinTransactionMock.On("InsertKymaConfigModule", "kymaID", model.KymaModule("monitoring")).Return(nil)
			writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return("operationID", nil)
			writeSessionWithinTransactionMock.On("Commit").Return(nil)

			sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

			runtimeService := NewRuntimeService(sessionFactoryMock)

			// when
			provisioningOperation, err := runtimeService.SetProvisioningStarted(runtimeID, cfg.config)

			// then
			assert.NoError(t, err)
			assert.Equal(t, provisioningOperation.ID, "operationID")
			assert.Equal(t, provisioningOperation.Type, model.Provision)
			assert.Equal(t, provisioningOperation.State, model.InProgress)
			assert.Equal(t, provisioningOperation.ClusterID, runtimeID)
			assert.NotEmpty(t, provisioningOperation.Message)

			sessionFactoryMock.AssertExpectations(t)
			writeSessionWithinTransactionMock.AssertExpectations(t)
		})
	}

	t.Run("Should return erro when failed to rollback transaction", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		terraformState := "{}"
		writeSessionWithinTransactionMock.On("InsertCluster", runtimeID, mock.AnythingOfType("Time"), terraformState).Return(dberrors.Internal("some error"))
		writeSessionWithinTransactionMock.On("Rollback").Return(dberrors.Internal("some error"))

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewRuntimeService(sessionFactoryMock)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
	})

	t.Run("Should rollback transaction when failed to insert record to Cluster table", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		terraformState := "{}"
		writeSessionWithinTransactionMock.On("InsertCluster", runtimeID, mock.AnythingOfType("Time"), terraformState).Return(dberrors.Internal("some error"))
		writeSessionWithinTransactionMock.On("Rollback").Return(nil)

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewRuntimeService(sessionFactoryMock)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
	})

	t.Run("Should rollback transaction when failed to insert record to KymaConfig table", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		terraformState := "{}"
		writeSessionWithinTransactionMock.On("InsertCluster", runtimeID, mock.AnythingOfType("Time"), terraformState).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGCPConfig", runtimeID, gcpConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", runtimeID, kymaConfig.Version).Return("", dberrors.Internal("some error"))
		writeSessionWithinTransactionMock.On("Rollback").Return(nil)

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewRuntimeService(sessionFactoryMock)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
	})

	t.Run("Should rollback transaction when failed to insert record to KymaConfigModule table", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		terraformState := "{}"
		writeSessionWithinTransactionMock.On("InsertCluster", runtimeID, mock.AnythingOfType("Time"), terraformState).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGCPConfig", runtimeID, gcpConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", runtimeID, kymaConfig.Version).Return("kymaID", nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfigModule", "kymaID", model.KymaModule("core")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfigModule", "kymaID", model.KymaModule("monitoring")).Return(dberrors.Internal("some error"))
		writeSessionWithinTransactionMock.On("Rollback").Return(nil)

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewRuntimeService(sessionFactoryMock)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
	})

	clusterConfigurations := []struct {
		config model.RuntimeConfig

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

			terraformState := "{}"
			writeSessionWithinTransactionMock.On("InsertCluster", runtimeID, mock.AnythingOfType("Time"), terraformState).Return(nil)
			writeSessionWithinTransactionMock.On(cfg.insertClusterConfigMethodName, runtimeID, cfg.config.ClusterConfig).Return(dberrors.Internal("some error"))
			writeSessionWithinTransactionMock.On("Rollback").Return(nil)

			sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

			runtimeService := NewRuntimeService(sessionFactoryMock)

			// when
			_, err := runtimeService.SetProvisioningStarted(runtimeID, cfg.config)

			// then
			assert.Error(t, err)

			sessionFactoryMock.AssertExpectations(t)
			writeSessionWithinTransactionMock.AssertExpectations(t)
		})
	}

	t.Run("Should rollback transaction when failed to insert record to Type table", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		terraformState := "{}"
		writeSessionWithinTransactionMock.On("InsertCluster", runtimeID, mock.AnythingOfType("Time"), terraformState).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGCPConfig", runtimeID, gcpConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", runtimeID, kymaConfig.Version).Return("kymaID", nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfigModule", "kymaID", model.KymaModule("core")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfigModule", "kymaID", model.KymaModule("monitoring")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return("", dberrors.Internal("some error"))

		writeSessionWithinTransactionMock.On("Rollback").Return(nil)

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewRuntimeService(sessionFactoryMock)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to commit transaction", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		terraformState := "{}"
		writeSessionWithinTransactionMock.On("InsertCluster", runtimeID, mock.AnythingOfType("Time"), terraformState).Return(nil)
		writeSessionWithinTransactionMock.On("InsertGCPConfig", runtimeID, gcpConfig).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfig", runtimeID, kymaConfig.Version).Return("kymaID", nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfigModule", "kymaID", model.KymaModule("core")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertKymaConfigModule", "kymaID", model.KymaModule("monitoring")).Return(nil)
		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return("operationID", nil)

		writeSessionWithinTransactionMock.On("Commit").Return(dberrors.Internal("some error"))

		sessionFactoryMock.On("NewSessionWithinTransaction").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewRuntimeService(sessionFactoryMock)

		// when
		_, err := runtimeService.SetProvisioningStarted(runtimeID, runtimeGCPConfig)

		// then
		assert.Error(t, err)

		sessionFactoryMock.AssertExpectations(t)
		writeSessionWithinTransactionMock.AssertExpectations(t)
	})
}

func TestSetDeprovisioning(t *testing.T) {
	runtimeID := "runtimeID"

	operation := model.Operation{
		Type:           model.Deprovision,
		StartTimestamp: time.Now(),
		State:          model.InProgress,
		Message:        "Deprovisioning started.",
		ClusterID:      runtimeID,
	}

	operationMatcher := getOperationMather(operation)

	t.Run("Should set deprovisioning started", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return("operationID", nil)

		sessionFactoryMock.On("NewWriteSession").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewRuntimeService(sessionFactoryMock)

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
	})
}

func TestSetUpgrade(t *testing.T) {

	runtimeID := "runtimeID"

	operation := model.Operation{
		Type:           model.Upgrade,
		StartTimestamp: time.Now(),
		State:          model.InProgress,
		Message:        "Upgrade started.",
		ClusterID:      runtimeID,
	}

	operationMatcher := getOperationMather(operation)

	t.Run("Should set upgrade started", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionWithinTransactionMock := &sessionMocks.WriteSessionWithinTransaction{}

		writeSessionWithinTransactionMock.On("InsertOperation", mock.MatchedBy(operationMatcher)).Return("operationID", nil)

		sessionFactoryMock.On("NewWriteSession").Return(writeSessionWithinTransactionMock, nil)

		runtimeService := NewRuntimeService(sessionFactoryMock)

		// when
		provisioningOperation, err := runtimeService.SetUpgradeStarted(runtimeID)

		// then
		assert.NoError(t, err)
		assert.Equal(t, provisioningOperation.ID, "operationID")
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
		Version: "1.6",
		Modules: []model.KymaConfigModule{
			{ID: "id1", Module: model.KymaModule("core")},
			{ID: "id2", Module: model.KymaModule("monitoring")},
		},
	}

	t.Run("Should get runtime status", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSessionMock := &sessionMocks.ReadSession{}

		readSessionMock.On("GetLastOperation", runtimeID).Return(operation, nil)
		readSessionMock.On("GetClusterConfig", runtimeID).Return(gcpConfig, nil)
		readSessionMock.On("GetKymaConfig", runtimeID).Return(kymaConfig, nil)

		sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)

		expected := model.RuntimeStatus{
			LastOperationStatus: operation,
			RuntimeConfiguration: model.RuntimeConfig{
				ClusterConfig: gcpConfig,
				KymaConfig:    kymaConfig,
			},
		}

		// when
		runtimeService := NewRuntimeService(sessionFactoryMock)
		runtimeStatus, err := runtimeService.GetStatus(runtimeID)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expected, runtimeStatus)
	})

	t.Run("Should fail to get runtime status when getting last operation failed", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSessionMock := &sessionMocks.ReadSession{}

		readSessionMock.On("GetLastOperation", runtimeID).Return(model.Operation{}, dberrors.Internal("some error"))
		sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)

		// when
		runtimeService := NewRuntimeService(sessionFactoryMock)
		_, err := runtimeService.GetStatus(runtimeID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should fail to get runtime status when getting cluster config failed", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSessionMock := &sessionMocks.ReadSession{}

		readSessionMock.On("GetLastOperation", runtimeID).Return(operation, nil)
		readSessionMock.On("GetClusterConfig", runtimeID).Return(model.GCPConfig{}, dberrors.Internal("some error"))
		sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)

		// when
		runtimeService := NewRuntimeService(sessionFactoryMock)
		_, err := runtimeService.GetStatus(runtimeID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should fail to get runtime status when getting cluster config failed", func(t *testing.T) {
		// given
		sessionFactoryMock := &sessionMocks.Factory{}
		readSessionMock := &sessionMocks.ReadSession{}

		readSessionMock.On("GetLastOperation", runtimeID).Return(operation, nil)
		readSessionMock.On("GetClusterConfig", runtimeID).Return(gcpConfig, nil)
		readSessionMock.On("GetKymaConfig", runtimeID).Return(model.KymaConfig{}, dberrors.Internal("some error"))

		sessionFactoryMock.On("NewReadSession").Return(readSessionMock, nil)

		// when
		runtimeService := NewRuntimeService(sessionFactoryMock)
		_, err := runtimeService.GetStatus(runtimeID)

		// then
		assert.Error(t, err)
	})
}

func getOperationMather(expected model.Operation) func(model.Operation) bool {
	return func(op model.Operation) bool {
		return op.Type == expected.Type &&
			op.Message == expected.Message && op.ClusterID == expected.ClusterID &&
			op.State == expected.State && op.ID == expected.ID
	}
}
