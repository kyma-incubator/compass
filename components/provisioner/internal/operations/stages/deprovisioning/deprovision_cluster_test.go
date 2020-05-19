package deprovisioning

import (
	"errors"
	"testing"
	"time"

	directorMocks "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	installationMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	gardener_mocks "github.com/kyma-incubator/compass/components/provisioner/internal/operations/stages/deprovisioning/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	dbMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	runtimeID = "runtimeID"
	tenant    = "tenant"
)

func TestDeprovisionCluster_Run(t *testing.T) {

	cluster := model.Cluster{
		ID: "runtimeID",
		ClusterConfig: model.GardenerConfig{
			Name: clusterName,
		},
		Tenant: "tenant",
	}

	for _, testCase := range []struct {
		description   string
		mockFunc      func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient)
		expectedStage model.OperationStage
		expectedDelay time.Duration
	}{
		{
			description: "should go to the next step when Shoot was deleted successfully and Runtime unregistered",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
				gardenerClient.On("Delete", clusterName, mock.Anything).Return(nil)
				dbSession := &dbMocks.WriteSessionWithinTransaction{}
				dbSession.On("MarkClusterAsDeleted", runtimeID).Return(nil)
				dbSessionFactory.On("NewSessionWithinTransaction").Return(dbSession, nil)
				directorClient.On("RuntimeExists", runtimeID, tenant).Return(true, nil)
				directorClient.On("DeleteRuntime", runtimeID, tenant).Return(nil)
				dbSession.On("Commit").Return(nil)
				dbSession.On("RollbackUnlessCommitted").Return()
			},
			expectedStage: nextStageName,
			expectedDelay: 0,
		},
		{
			description: "should go to the next step when Shoot and Runtime not exists",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
				gardenerClient.On("Delete", clusterName, mock.Anything).Return(nil)
				dbSession := &dbMocks.WriteSessionWithinTransaction{}
				dbSession.On("MarkClusterAsDeleted", runtimeID).Return(nil)
				dbSessionFactory.On("NewSessionWithinTransaction").Return(dbSession, nil)
				directorClient.On("RuntimeExists", runtimeID, tenant).Return(false, nil)
				dbSession.On("Commit").Return(nil)
				dbSession.On("RollbackUnlessCommitted").Return()
			},
			expectedStage: nextStageName,
			expectedDelay: 0,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			gardenerClient := &gardener_mocks.GardenerClient{}
			dbSessionFactory := &dbMocks.Factory{}
			directorClient := &directorMocks.DirectorClient{}

			testCase.mockFunc(gardenerClient, dbSessionFactory, directorClient)

			deprovisionClusterStep := NewDeprovisionClusterStep(gardenerClient, dbSessionFactory, directorClient, nextStageName, 10*time.Minute)

			// when
			result, err := deprovisionClusterStep.Run(cluster, model.Operation{}, logrus.New())

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedStage, result.Stage)
			assert.Equal(t, testCase.expectedDelay, result.Delay)
			gardenerClient.AssertExpectations(t)
			directorClient.AssertExpectations(t)
		})
	}

	for _, testCase := range []struct {
		description        string
		mockFunc           func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient)
		cluster            model.Cluster
		unrecoverableError bool
	}{
		{
			description: "should return unrecoverable error when failed to get GardenerConfig",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
			},
			cluster:            model.Cluster{},
			unrecoverableError: true,
		},
		{
			description: "should return error when failed to delete shoot",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
				gardenerClient.On("Delete", clusterName, mock.Anything).Return(errors.New("some error"))
			},
			cluster:            cluster,
			unrecoverableError: false,
		},
		{
			description: "should return error when failed to start database transaction",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
				gardenerClient.On("Delete", clusterName, mock.Anything).Return(nil)
				dbSessionFactory.On("NewSessionWithinTransaction").Return(nil, dberrors.Internal("some error"))
			},
			cluster:            cluster,
			unrecoverableError: false,
		},
		{
			description: "should return error when failed to mark cluster as deleted",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
				gardenerClient.On("Delete", clusterName, mock.Anything).Return(nil)
				dbSession := &dbMocks.WriteSessionWithinTransaction{}
				dbSession.On("MarkClusterAsDeleted", runtimeID).Return(dberrors.Internal("some error"))
				dbSessionFactory.On("NewSessionWithinTransaction").Return(dbSession, nil)
				dbSession.On("RollbackUnlessCommitted").Return()
			},
			cluster:            cluster,
			unrecoverableError: false,
		},
		{
			description: "should return error when failed to check if Runtime exists",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
				gardenerClient.On("Delete", clusterName, mock.Anything).Return(nil)
				dbSession := &dbMocks.WriteSessionWithinTransaction{}
				dbSession.On("MarkClusterAsDeleted", runtimeID).Return(nil)
				dbSessionFactory.On("NewSessionWithinTransaction").Return(dbSession, nil)
				dbSession.On("RollbackUnlessCommitted").Return()
				directorClient.On("RuntimeExists", runtimeID, tenant).Return(false, errors.New("some error"))
			},
			cluster:            cluster,
			unrecoverableError: false,
		},
		{
			description: "should return error when failed to delete Runtime",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
				gardenerClient.On("Delete", clusterName, mock.Anything).Return(nil)
				dbSession := &dbMocks.WriteSessionWithinTransaction{}
				dbSession.On("MarkClusterAsDeleted", runtimeID).Return(nil)
				dbSessionFactory.On("NewSessionWithinTransaction").Return(dbSession, nil)
				dbSession.On("RollbackUnlessCommitted").Return()
				directorClient.On("RuntimeExists", runtimeID, tenant).Return(true, nil)
				directorClient.On("DeleteRuntime", runtimeID, tenant).Return(errors.New("some error"))
			},
			cluster:            cluster,
			unrecoverableError: false,
		},
		{
			description: "should return error when failed to commit database transaction",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSessionFactory *dbMocks.Factory, directorClient *directorMocks.DirectorClient) {
				gardenerClient.On("Delete", clusterName, mock.Anything).Return(nil)
				dbSession := &dbMocks.WriteSessionWithinTransaction{}
				dbSession.On("MarkClusterAsDeleted", mock.AnythingOfType("string")).Return(nil)
				dbSessionFactory.On("NewSessionWithinTransaction").Return(dbSession, nil)
				directorClient.On("RuntimeExists", runtimeID, tenant).Return(true, nil)
				directorClient.On("DeleteRuntime", runtimeID, tenant).Return(nil)
				dbSession.On("Commit").Return(dberrors.Internal("some error"))
				dbSession.On("RollbackUnlessCommitted").Return()
			},
			cluster:            cluster,
			unrecoverableError: false,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			installationSvc := &installationMocks.Service{}
			gardenerClient := &gardener_mocks.GardenerClient{}
			dbSessionFactory := &dbMocks.Factory{}
			directorClient := &directorMocks.DirectorClient{}

			testCase.mockFunc(gardenerClient, dbSessionFactory, directorClient)

			deprovisionClusterStep := NewDeprovisionClusterStep(gardenerClient, dbSessionFactory, directorClient, nextStageName, 10*time.Minute)

			// when
			_, err := deprovisionClusterStep.Run(testCase.cluster, model.Operation{}, logrus.New())

			// then
			require.Error(t, err)
			nonRecoverable := operations.NonRecoverableError{}
			require.Equal(t, testCase.unrecoverableError, errors.As(err, &nonRecoverable))
			installationSvc.AssertExpectations(t)
			gardenerClient.AssertExpectations(t)
			dbSessionFactory.AssertExpectations(t)
			directorClient.AssertExpectations(t)
		})
	}
}
