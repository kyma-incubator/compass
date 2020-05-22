package provisioning

import (
	"errors"
	"testing"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	gardener_mocks "github.com/kyma-incubator/compass/components/provisioner/internal/operations/stages/deprovisioning/mocks"
	provisioning_mocks "github.com/kyma-incubator/compass/components/provisioner/internal/operations/stages/provisioning/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	dbMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWaitForClusterInitialization_Run(t *testing.T) {

	clusterName := "name"
	runtimeID := "runtimeID"
	tenant := "tenant"

	cluster := model.Cluster{
		ID:     runtimeID,
		Tenant: tenant,
		ClusterConfig: model.GardenerConfig{
			Name: clusterName,
		},
	}

	for _, testCase := range []struct {
		description   string
		mockFunc      func(ardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider)
		expectedStage model.OperationStage
		expectedDelay time.Duration
	}{
		{
			description: "should continue waiting if cluster not created",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider) {

				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootInProcessingState(clusterName), nil)
			},
			expectedStage: model.WaitingForClusterCreation,
			expectedDelay: 20 * time.Second,
		},
		{
			description: "should continue waiting if last operation not set",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider) {

				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootInUnknownState(clusterName), nil)
			},
			expectedStage: model.WaitingForClusterCreation,
			expectedDelay: 20 * time.Second,
		},
		{
			description: "should go to the next stage if cluster was created",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootInSucceededState(clusterName), nil)
				kubeconfigProvider.On("FetchRaw", clusterName).Return([]byte("kubeconfig"), nil)

				dbSession.On("UpdateCluster", cluster.ID, "kubeconfig", []byte(nil)).Return(nil)

			},
			expectedStage: nextStageName,
			expectedDelay: 0,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			gardenerClient := &gardener_mocks.GardenerClient{}
			dbSession := &dbMocks.ReadWriteSession{}
			kubeconfigProvider := &provisioning_mocks.KubeconfigProvider{}

			testCase.mockFunc(gardenerClient, dbSession, kubeconfigProvider)

			waitForClusterCreationStep := NewWaitForClusterCreationStep(gardenerClient, dbSession, kubeconfigProvider, nextStageName, 10*time.Minute)
			// when
			result, err := waitForClusterCreationStep.Run(cluster, model.Operation{}, logrus.New())

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedStage, result.Stage)
			assert.Equal(t, testCase.expectedDelay, result.Delay)
			gardenerClient.AssertExpectations(t)
		})
	}

	for _, testCase := range []struct {
		description        string
		mockFunc           func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider)
		cluster            model.Cluster
		unrecoverableError bool
	}{
		{
			description: "should return unrecoverable error when failed to get GardenerConfig",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider) {
			},
			unrecoverableError: true,
			cluster:            model.Cluster{},
		},
		{
			description: "should return error if failed to read Shoot",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(nil, errors.New("some error"))
			},
			unrecoverableError: false,
			cluster:            cluster,
		},
		{
			description: "should return error if failed to fetch kubeconfig",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootInSucceededState(clusterName), nil)
				kubeconfigProvider.On("FetchRaw", clusterName).Return(nil, errors.New("some error"))
			},
			unrecoverableError: false,
			cluster:            cluster,
		},
		{
			description: "should return unrecoverable error if Shoot is in failed state",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootInFailedState(clusterName), nil)
			},
			unrecoverableError: true,
			cluster:            cluster,
		},
		{
			description: "should return error if failed to update data in database",
			mockFunc: func(gardenerClient *gardener_mocks.GardenerClient, dbSession *dbMocks.ReadWriteSession, kubeconfigProvider *provisioning_mocks.KubeconfigProvider) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootInSucceededState(clusterName), nil)
				kubeconfigProvider.On("FetchRaw", clusterName).Return([]byte("kubeconfig"), nil)

				dbSession.On("UpdateCluster", cluster.ID, "kubeconfig", []byte(nil)).Return(dberrors.Internal("some error"))
			},
			unrecoverableError: false,
			cluster:            cluster,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			gardenerClient := &gardener_mocks.GardenerClient{}
			dbSession := &dbMocks.ReadWriteSession{}
			kubeconfigProvider := &provisioning_mocks.KubeconfigProvider{}

			testCase.mockFunc(gardenerClient, dbSession, kubeconfigProvider)

			waitForClusterCreationStep := NewWaitForClusterCreationStep(gardenerClient, dbSession, kubeconfigProvider, nextStageName, 10*time.Minute)

			// when
			_, err := waitForClusterCreationStep.Run(testCase.cluster, model.Operation{}, logrus.New())

			// then
			require.Error(t, err)
			nonRecoverable := operations.NonRecoverableError{}
			require.Equal(t, testCase.unrecoverableError, errors.As(err, &nonRecoverable))

			gardenerClient.AssertExpectations(t)
			dbSession.AssertExpectations(t)
			kubeconfigProvider.AssertExpectations(t)
		})
	}
}

func fixShootInSucceededState(name string) *gardener_types.Shoot {
	return fixShoot(name, &gardener_types.LastOperation{
		State: gardencorev1beta1.LastOperationStateSucceeded,
	})
}

func fixShootInFailedState(name string) *gardener_types.Shoot {
	return fixShoot(name, &gardener_types.LastOperation{
		State: gardencorev1beta1.LastOperationStateFailed,
	})
}

func fixShootInProcessingState(name string) *gardener_types.Shoot {
	return fixShoot(name, &gardener_types.LastOperation{
		State: gardencorev1beta1.LastOperationStateProcessing,
	})
}

func fixShootInUnknownState(name string) *gardener_types.Shoot {
	return fixShoot(name, nil)
}

func fixShoot(name string, lastOperation *gardener_types.LastOperation) *gardener_types.Shoot {
	return &gardener_types.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: gardener_types.ShootStatus{
			LastOperation: lastOperation,
		},
	}
}
