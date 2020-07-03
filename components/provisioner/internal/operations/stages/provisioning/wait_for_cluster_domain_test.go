package provisioning

import (
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/control-plane/components/provisioner/internal/apperrors"

	"github.com/kyma-project/control-plane/components/provisioner/internal/operations"
	"github.com/stretchr/testify/mock"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	directormock "github.com/kyma-project/control-plane/components/provisioner/internal/director/mocks"
	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	gardenerMocks "github.com/kyma-project/control-plane/components/provisioner/internal/operations/stages/provisioning/mocks"
	"github.com/kyma-project/control-plane/components/provisioner/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWaitForClusterDomain_Run(t *testing.T) {

	clusterName := "name"
	runtimeID := "runtimeID"
	tenant := "tenant"
	domain := "cluster.kymaa.com"

	cluster := model.Cluster{
		ID:     runtimeID,
		Tenant: tenant,
		ClusterConfig: model.GardenerConfig{
			Name: clusterName,
		},
		Kubeconfig: util.StringPtr(kubeconfig),
	}

	for _, testCase := range []struct {
		description   string
		mockFunc      func(gardenerClient *gardenerMocks.GardenerClient, directorClient *directormock.DirectorClient)
		expectedStage model.OperationStage
		expectedDelay time.Duration
	}{
		{
			description: "should continue waiting if domain name is not set",
			mockFunc: func(gardenerClient *gardenerMocks.GardenerClient, directorClient *directormock.DirectorClient) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(&gardener_types.Shoot{}, nil)
			},
			expectedStage: model.WaitingForClusterDomain,
			expectedDelay: 5 * time.Second,
		},
		{
			description: "should go to the next stage if domain name is available",
			mockFunc: func(gardenerClient *gardenerMocks.GardenerClient, directorClient *directormock.DirectorClient) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootWithDomainSet(clusterName, domain), nil)

				runtime := fixRuntime(runtimeID, clusterName, map[string]interface{}{
					"label": "value",
				})
				directorClient.On("GetRuntime", runtimeID, tenant).Return(runtime, nil)
				directorClient.On("UpdateRuntime", runtimeID, mock.Anything, tenant).Return(nil)
			},
			expectedStage: nextStageName,
			expectedDelay: 0,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			gardenerClient := &gardenerMocks.GardenerClient{}
			directorClient := &directormock.DirectorClient{}

			testCase.mockFunc(gardenerClient, directorClient)

			waitForClusterDomainStep := NewWaitForClusterDomainStep(gardenerClient, directorClient, nextStageName, 10*time.Minute)

			// when
			result, err := waitForClusterDomainStep.Run(cluster, model.Operation{}, logrus.New())

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
		mockFunc           func(gardenerClient *gardenerMocks.GardenerClient, directorClient *directormock.DirectorClient)
		cluster            model.Cluster
		unrecoverableError bool
	}{
		{
			description: "should return error if failed to read Shoot",
			mockFunc: func(gardenerClient *gardenerMocks.GardenerClient, directorClient *directormock.DirectorClient) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(nil, apperrors.Internal("some error"))
			},
			unrecoverableError: false,
			cluster:            cluster,
		},
		{
			description: "should return error if failed to get Runtime from Director",
			mockFunc: func(gardenerClient *gardenerMocks.GardenerClient, directorClient *directormock.DirectorClient) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootWithDomainSet(clusterName, domain), nil)
				directorClient.On("GetRuntime", runtimeID, tenant).Return(graphql.RuntimeExt{}, apperrors.Internal("some error"))

			},
			unrecoverableError: false,
			cluster:            cluster,
		},
		{
			description: "should return error if failed to update Runtime in Director",
			mockFunc: func(gardenerClient *gardenerMocks.GardenerClient, directorClient *directormock.DirectorClient) {
				gardenerClient.On("Get", clusterName, mock.Anything).Return(fixShootWithDomainSet(clusterName, domain), nil)

				runtime := fixRuntime(runtimeID, clusterName, map[string]interface{}{
					"label": "value",
				})
				directorClient.On("GetRuntime", runtimeID, tenant).Return(runtime, nil)
				directorClient.On("UpdateRuntime", runtimeID, mock.Anything, tenant).Return(apperrors.Internal("some error"))
			},
			unrecoverableError: false,
			cluster:            cluster,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			gardenerClient := &gardenerMocks.GardenerClient{}
			directorClient := &directormock.DirectorClient{}

			testCase.mockFunc(gardenerClient, directorClient)

			waitForClusterDomainStep := NewWaitForClusterDomainStep(gardenerClient, directorClient, nextStageName, 10*time.Minute)

			// when
			_, err := waitForClusterDomainStep.Run(testCase.cluster, model.Operation{}, logrus.New())

			// then
			require.Error(t, err)
			nonRecoverable := operations.NonRecoverableError{}
			require.Equal(t, testCase.unrecoverableError, errors.As(err, &nonRecoverable))

			gardenerClient.AssertExpectations(t)
			directorClient.AssertExpectations(t)
		})
	}
}

func fixShootWithDomainSet(name, domain string) *gardener_types.Shoot {
	return &gardener_types.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: gardener_types.ShootSpec{
			DNS: &gardener_types.DNS{
				Domain: &domain,
			},
		},
	}
}

func fixRuntime(runtimeId, name string, labels map[string]interface{}) graphql.RuntimeExt {
	return graphql.RuntimeExt{
		Runtime: graphql.Runtime{
			ID:   runtimeId,
			Name: name,
		},
		Labels: labels,
	}
}
