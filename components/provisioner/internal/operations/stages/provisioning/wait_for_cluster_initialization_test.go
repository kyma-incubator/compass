package provisioning

import (
	"testing"
)

func TestWaitForClusterInitialization_Run(t *testing.T) {

	//clusterName := "name"
	//runtimeID := "runtimeID"
	//tenant := "tenant"
	//domain := "cluster.kymaa.com"
	//
	//cluster := model.Cluster{
	//	ID:     runtimeID,
	//	Tenant: tenant,
	//	ClusterConfig: model.GardenerConfig{
	//		Name: clusterName,
	//	},
	//	Kubeconfig: util.StringPtr(kubeconfig),
	//}
	//
	//for _, testCase := range []struct {
	//	description   string
	//	mockFunc      func(gardenerClient *mocks.GardenerClient)
	//	expectedStage model.OperationStage
	//	expectedDelay time.Duration
	//}{
	//	{
	//		description: "should continue waiting if domain name is not set",
	//		mockFunc: func(gardenerClient *mocks.GardenerClient) {
	//			gardenerClient.On("Get", clusterName, mock.Anything).Return(&gardener_types.Shoot{}, nil)
	//		},
	//		expectedStage: model.WaitingForClusterDomain,
	//		expectedDelay: 5 * time.Second,
	//	},
	//	{
	//		description: "should go to the next stage if domain name is available",
	//		mockFunc: func(gardenerClient *mocks.GardenerClient) {
	//			gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootWithDomainSet(clusterName, domain), nil)
	//
	//
	//		},
	//		expectedStage: nextStageName,
	//		expectedDelay: 0,
	//	},
	//} {
	//	t.Run(testCase.description, func(t *testing.T) {
	//		// given
	//		gardenerClient := &mocks.GardenerClient{}
	//		dbSession := &dbSessionMocks.Factory{}
	//		secretClient := &mocks.
	//		testCase.mockFunc(gardenerClient)
	//
	//		waitForClusterInitializationStep := NewWaitForClusterInitializationStep(gardenerClient, dbSession, nextStageName, 10*time.Minute)
	//
	//		// when
	//		result, err := waitForClusterInitializationStep.Run(cluster, model.Operation{}, logrus.New())
	//
	//		// then
	//		require.NoError(t, err)
	//		assert.Equal(t, testCase.expectedStage, result.Stage)
	//		assert.Equal(t, testCase.expectedDelay, result.Delay)
	//		gardenerClient.AssertExpectations(t)
	//	})
	//}
	//
	//for _, testCase := range []struct {
	//	description        string
	//	mockFunc           func(gardenerClient *mocks.GardenerClient, directorClient *directormock.DirectorClient)
	//	cluster            model.Cluster
	//	unrecoverableError bool
	//}{
	//	{
	//		description: "should return unrecoverable error when failed to get GardenerConfig",
	//		mockFunc: func(gardenerClient *mocks.GardenerClient, directorClient *directormock.DirectorClient) {
	//
	//		},
	//		unrecoverableError: true,
	//	},
	//	{
	//		description: "should return error if failed to read Shoot",
	//		mockFunc: func(gardenerClient *mocks.GardenerClient, directorClient *directormock.DirectorClient) {
	//			gardenerClient.On("Get", clusterName, mock.Anything).Return(nil, errors.New("some error"))
	//		},
	//		unrecoverableError: false,
	//		cluster:            cluster,
	//	},
	//	{
	//		description: "should return error if failed to get Runtime from Director",
	//		mockFunc: func(gardenerClient *mocks.GardenerClient, directorClient *directormock.DirectorClient) {
	//			gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootWithDomainSet(clusterName, domain), nil)
	//
	//			directorClient.On("GetRuntime", runtimeID, tenant).Return(graphql.RuntimeExt{}, errors.New("some error"))
	//
	//		},
	//		unrecoverableError: false,
	//		cluster:            cluster,
	//	},
	//	{
	//		description: "should return error if failed to update Runtime in Director",
	//		mockFunc: func(gardenerClient *mocks.GardenerClient, directorClient *directormock.DirectorClient) {
	//			gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootWithDomainSet(clusterName, domain), nil)
	//
	//			runtime := fixRuntime(runtimeID, clusterName, map[string]interface{}{
	//				"label": "value",
	//			})
	//			directorClient.On("GetRuntime", runtimeID, tenant).Return(runtime, nil)
	//
	//			directorClient.On("UpdateRuntime", runtimeID, mock.Anything, tenant).Return(errors.New("some error"))
	//		},
	//		unrecoverableError: false,
	//		cluster:            cluster,
	//	},
	//} {
	//	t.Run(testCase.description, func(t *testing.T) {
	//		// given
	//		gardenerClient := &mocks.GardenerClient{}
	//		directorClient := &directormock.DirectorClient{}
	//
	//		testCase.mockFunc(gardenerClient, directorClient)
	//
	//		waitForClusterDomainStep := NewWaitForClusterDomainStep(gardenerClient, directorClient, nextStageName, 10*time.Minute)
	//
	//		// when
	//		_, err := waitForClusterDomainStep.Run(testCase.cluster, model.Operation{}, logrus.New())
	//
	//		// then
	//		require.Error(t, err)
	//		nonRecoverable := operations.NonRecoverableError{}
	//		require.Equal(t, testCase.unrecoverableError, errors.As(err, &nonRecoverable))
	//
	//		gardenerClient.AssertExpectations(t)
	//		directorClient.AssertExpectations(t)
	//	})
	//}
}
