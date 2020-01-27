package hydroform_test

import (
	"errors"
	"testing"

	directormock "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/mocks"
	installationMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	sessionMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	runtimeId      = "runtimeId"
	operationId    = "operationId"
	kubeconfigFile = "kubeconfig data"
	kymaVersion    = "1.9.0"
)

var (
	kymaRelease = model.Release{
		Id:            "releaseId",
		Version:       kymaVersion,
		TillerYAML:    "tiller yaml",
		InstallerYAML: "installer yaml",
	}

	globalConfig     = model.Configuration{}
	componentsConfig = []model.KymaComponentConfig{}
)

func Test_Provision(t *testing.T) {

	cluster := model.Cluster{
		ID: runtimeId,
		KymaConfig: model.KymaConfig{
			Release:             kymaRelease,
			GlobalConfiguration: globalConfig,
			Components:          componentsConfig,
		},
	}

	t.Run("should provision cluster", func(t *testing.T) {
		// given
		hydroformMock := &mocks.Service{}
		installationSvc := &installationMocks.Service{}
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionMock := &sessionMocks.WriteSession{}

		hydroformMock.On("ProvisionCluster", cluster).Return(
			hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")},
			nil,
		)
		sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
		writeSessionMock.On("UpdateCluster", runtimeId, kubeconfigFile, []byte("")).Return(nil)
		installationSvc.On("InstallKyma", runtimeId, kubeconfigFile, kymaRelease, globalConfig, componentsConfig).Return(nil)
		writeSessionMock.On("UpdateOperationState", operationId, "Operation succeeded.", model.Succeeded).Return(nil)

		hydroformProvisioner := hydroform.NewHydroformProvisioner(hydroformMock, installationSvc, sessionFactoryMock, nil)

		// when
		channel, err := hydroformProvisioner.Provision(cluster, operationId)

		// then
		require.NoError(t, err)
		waitUntilFinished(channel)
	})

	for _, testCase := range []struct {
		description string
		mockFunc    func(*mocks.Service, *sessionMocks.Factory, *installationMocks.Service, *sessionMocks.WriteSession)
	}{
		{
			description: "fail to install Kyma",
			mockFunc: func(hydroformMock *mocks.Service, sessionFactoryMock *sessionMocks.Factory, installationSvc *installationMocks.Service, writeSessionMock *sessionMocks.WriteSession) {
				hydroformMock.On("ProvisionCluster", cluster).Return(
					hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")},
					nil,
				)
				sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
				writeSessionMock.On("UpdateCluster", runtimeId, kubeconfigFile, []byte("")).Return(nil)
				installationSvc.On("InstallKyma", runtimeId, kubeconfigFile, kymaRelease, globalConfig, componentsConfig).Return(errors.New("error"))
				writeSessionMock.On("UpdateOperationState", operationId, mock.AnythingOfType("string"), model.Failed).Return(nil)

			},
		},
		{
			description: "fail to save kubeconfig",
			mockFunc: func(hydroformMock *mocks.Service, sessionFactoryMock *sessionMocks.Factory, installationSvc *installationMocks.Service, writeSessionMock *sessionMocks.WriteSession) {
				hydroformMock.On("ProvisionCluster", cluster).Return(
					hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: kubeconfigFile, State: []byte("")},
					nil,
				)
				sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
				writeSessionMock.On("UpdateCluster", runtimeId, kubeconfigFile, []byte("")).Return(dberrors.Internal("error"))
				writeSessionMock.On("UpdateOperationState", operationId, mock.AnythingOfType("string"), model.Failed).Return(nil)

			},
		},
		{
			description: "status different than Provisioned",
			mockFunc: func(hydroformMock *mocks.Service, sessionFactoryMock *sessionMocks.Factory, installationSvc *installationMocks.Service, writeSessionMock *sessionMocks.WriteSession) {
				hydroformMock.On("ProvisionCluster", cluster).Return(
					hydroform.ClusterInfo{ClusterStatus: types.Errored, KubeConfig: kubeconfigFile, State: []byte("")},
					nil,
				)
				sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
				writeSessionMock.On("UpdateOperationState", operationId, mock.AnythingOfType("string"), model.Failed).Return(nil)
			},
		},
		{
			description: "failed to provision",
			mockFunc: func(hydroformMock *mocks.Service, sessionFactoryMock *sessionMocks.Factory, installationSvc *installationMocks.Service, writeSessionMock *sessionMocks.WriteSession) {
				hydroformMock.On("ProvisionCluster", cluster).Return(hydroform.ClusterInfo{}, errors.New("error"))
				sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
				writeSessionMock.On("UpdateOperationState", operationId, mock.AnythingOfType("string"), model.Failed).Return(nil)

			},
		},
	} {
		t.Run("should set operation as failed when "+testCase.description, func(t *testing.T) {
			// given
			hydroformMock := &mocks.Service{}
			installationSvc := &installationMocks.Service{}
			sessionFactoryMock := &sessionMocks.Factory{}
			writeSessionMock := &sessionMocks.WriteSession{}

			testCase.mockFunc(hydroformMock, sessionFactoryMock, installationSvc, writeSessionMock)

			hydroformProvisioner := hydroform.NewHydroformProvisioner(hydroformMock, installationSvc, sessionFactoryMock, nil)

			// when
			channel, err := hydroformProvisioner.Provision(cluster, operationId)

			// then
			require.NoError(t, err)
			waitUntilFinished(channel)
		})
	}

}

func Test_Deprovision(t *testing.T) {
	cluster := model.Cluster{
		ID: runtimeId,
		KymaConfig: model.KymaConfig{
			Release:             kymaRelease,
			GlobalConfiguration: globalConfig,
			Components:          componentsConfig,
		},
	}

	t.Run("should deprovision cluster", func(t *testing.T) {
		// given
		hydroformMock := &mocks.Service{}
		sessionFactoryMock := &sessionMocks.Factory{}
		writeSessionMock := &sessionMocks.WriteSession{}
		directorServiceMock := &directormock.DirectorClient{}

		hydroformMock.On("DeprovisionCluster", cluster).Return(nil)
		sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
		writeSessionMock.On("MarkClusterAsDeleted", runtimeId).Return(nil)
		directorServiceMock.On("DeleteRuntime", runtimeId).Return(nil)
		writeSessionMock.On("UpdateOperationState", operationId, "Operation succeeded.", model.Succeeded).Return(nil)

		hydroformProvisioner := hydroform.NewHydroformProvisioner(hydroformMock, nil, sessionFactoryMock, directorServiceMock)

		// when
		op, channel, err := hydroformProvisioner.Deprovision(cluster, operationId)

		// then
		require.NoError(t, err)
		waitUntilFinished(channel)
		assert.Equal(t, operationId, op.ID)
	})

	for _, testCase := range []struct {
		description string
		mockFunc    func(*mocks.Service, *sessionMocks.Factory, *sessionMocks.WriteSession, *directormock.DirectorClient)
	}{
		{
			description: "failed to delete Runtime",
			mockFunc: func(hydroformMock *mocks.Service, sessionFactoryMock *sessionMocks.Factory, writeSessionMock *sessionMocks.WriteSession, directorServiceMock *directormock.DirectorClient) {
				hydroformMock.On("DeprovisionCluster", cluster).Return(nil)
				sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
				writeSessionMock.On("MarkClusterAsDeleted", runtimeId).Return(nil)
				directorServiceMock.On("DeleteRuntime", runtimeId).Return(errors.New("error"))
				writeSessionMock.On("UpdateOperationState", operationId, mock.AnythingOfType("string"), model.Failed).Return(nil)
			},
		},
		{
			description: "failed to mark cluster as deleted",
			mockFunc: func(hydroformMock *mocks.Service, sessionFactoryMock *sessionMocks.Factory, writeSessionMock *sessionMocks.WriteSession, directorServiceMock *directormock.DirectorClient) {
				hydroformMock.On("DeprovisionCluster", cluster).Return(nil)
				sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
				writeSessionMock.On("MarkClusterAsDeleted", runtimeId).Return(dberrors.Internal("error"))
				writeSessionMock.On("UpdateOperationState", operationId, mock.AnythingOfType("string"), model.Failed).Return(nil)
			},
		},
		{
			description: "failed to deprovision cluster",
			mockFunc: func(hydroformMock *mocks.Service, sessionFactoryMock *sessionMocks.Factory, writeSessionMock *sessionMocks.WriteSession, directorServiceMock *directormock.DirectorClient) {
				hydroformMock.On("DeprovisionCluster", cluster).Return(errors.New("error"))
				sessionFactoryMock.On("NewWriteSession").Return(writeSessionMock, nil)
				writeSessionMock.On("UpdateOperationState", operationId, mock.AnythingOfType("string"), model.Failed).Return(nil)
			},
		},
	} {
		t.Run("should set operation as failed when "+testCase.description, func(t *testing.T) {
			// given
			hydroformMock := &mocks.Service{}
			sessionFactoryMock := &sessionMocks.Factory{}
			writeSessionMock := &sessionMocks.WriteSession{}
			directorServiceMock := &directormock.DirectorClient{}

			testCase.mockFunc(hydroformMock, sessionFactoryMock, writeSessionMock, directorServiceMock)

			hydroformProvisioner := hydroform.NewHydroformProvisioner(hydroformMock, nil, sessionFactoryMock, directorServiceMock)

			// when
			op, channel, err := hydroformProvisioner.Deprovision(cluster, operationId)

			// then
			require.NoError(t, err)
			waitUntilFinished(channel)
			assert.Equal(t, operationId, op.ID)
		})
	}
}

func waitUntilFinished(finished <-chan struct{}) {
	for {
		_, ok := <-finished
		if !ok {
			break
		}
	}
}
