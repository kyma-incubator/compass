package deprovisioning

import (
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"

	installationMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCleanupCluster_Run(t *testing.T) {

	clusterWithKubeconfig := model.Cluster{
		ClusterConfig: model.GardenerConfig{
			Name: clusterName,
		},
		Kubeconfig: util.StringPtr(kubeconfig),
	}

	clusterWithoutKubeconfig := model.Cluster{
		ClusterConfig: model.GardenerConfig{
			Name: clusterName,
		},
	}

	invalidKubeconfig := "invalid"

	for _, testCase := range []struct {
		description   string
		mockFunc      func(installationSvc *installationMocks.Service)
		expectedStage model.OperationStage
		expectedDelay time.Duration
		cluster       model.Cluster
	}{
		{
			description: "should go to the next step when kubeconfig is empty",
			mockFunc: func(installationSvc *installationMocks.Service) {
			},
			expectedStage: nextStageName,
			expectedDelay: 0,
			cluster:       clusterWithoutKubeconfig,
		},
		{
			description: "should go to the next step when cleanup was performed successfully",
			mockFunc: func(installationSvc *installationMocks.Service) {
				installationSvc.On("PerformCleanup", mock.AnythingOfType("*rest.Config")).Return(nil)
			},
			expectedStage: nextStageName,
			expectedDelay: 0,
			cluster:       clusterWithKubeconfig,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			installationSvc := &installationMocks.Service{}

			testCase.mockFunc(installationSvc)

			cleanupClusterStep := NewCleanupClusterStep(installationSvc, nextStageName, 10*time.Minute)

			// when
			result, err := cleanupClusterStep.Run(testCase.cluster, model.Operation{}, logrus.New())

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedStage, result.Stage)
			assert.Equal(t, testCase.expectedDelay, result.Delay)
			installationSvc.AssertExpectations(t)
		})
	}

	for _, testCase := range []struct {
		description        string
		mockFunc           func(installationSvc *installationMocks.Service)
		cluster            model.Cluster
		unrecoverableError bool
	}{
		{
			description: "should return error is failed to parse kubeconfig",
			mockFunc: func(installationSvc *installationMocks.Service) {
			},
			cluster: model.Cluster{
				Kubeconfig: &invalidKubeconfig,
			},
			unrecoverableError: true,
		},
		{
			description: "should return error when failed to perform cleanup",
			mockFunc: func(installationSvc *installationMocks.Service) {
				installationSvc.On("PerformCleanup", mock.AnythingOfType("*rest.Config")).Return(errors.New("some error"))
			},
			cluster:            clusterWithKubeconfig,
			unrecoverableError: false,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			installationSvc := &installationMocks.Service{}

			testCase.mockFunc(installationSvc)

			cleanupClusterStep := NewCleanupClusterStep(installationSvc, nextStageName, 10*time.Minute)

			// when
			_, err := cleanupClusterStep.Run(testCase.cluster, model.Operation{}, logrus.New())

			// then
			require.Error(t, err)
			nonRecoverable := operations.NonRecoverableError{}
			require.Equal(t, testCase.unrecoverableError, errors.As(err, &nonRecoverable))
			installationSvc.AssertExpectations(t)
		})
	}
}
