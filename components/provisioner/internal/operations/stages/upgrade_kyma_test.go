package stages

import (
	"errors"
	installationMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUpgradeKymaStep_Run(t *testing.T) {
	t.Run("should return error when kubeconfig is nil", func(t *testing.T) {
		//given
		upgradeStep := NewUpgradeKymaStep(nil, nextStageName, 0)

		cluster := model.Cluster{Kubeconfig: nil}

		//when
		_, err := upgradeStep.Run(cluster, model.Operation{}, logrus.New())

		//then
		require.Error(t, err)
	})

	t.Run("should return error when kubeconfig is unparsable", func(t *testing.T) {
		//given
		upgradeStep := NewUpgradeKymaStep(nil, nextStageName, 0)

		kubeconfig := "wrong konfig"
		cluster := model.Cluster{Kubeconfig: &kubeconfig}

		//when
		_, err := upgradeStep.Run(cluster, model.Operation{}, logrus.New())

		//then
		require.Error(t, err)
	})

	t.Run("should return error when failed to check installation CR state", func(t *testing.T) {
		//given
		installationClient := &installationMocks.Service{}
		installationClient.On("CheckInstallationState", mock.Anything).Return(installation.InstallationState{}, errors.New("some error"))

		upgradeStep := NewUpgradeKymaStep(installationClient, nextStageName, 0)

		cluster := model.Cluster{Kubeconfig: util.StringPtr(kubeconfig)}

		//when
		_, err := upgradeStep.Run(cluster, model.Operation{}, logrus.New())

		//then
		require.Error(t, err)
	})

	t.Run("should return error when installation CR is not present on the cluster", func(t *testing.T) {
		//given
		installationClient := &installationMocks.Service{}
		installationClient.On("CheckInstallationState", mock.Anything).Return(installation.InstallationState{State: "NoInstallation"}, nil)

		upgradeStep := NewUpgradeKymaStep(installationClient, nextStageName, 0)

		cluster := model.Cluster{Kubeconfig: util.StringPtr(kubeconfig)}

		//when
		_, err := upgradeStep.Run(cluster, model.Operation{}, logrus.New())

		//then
		require.Error(t, err)
	})

	t.Run("should return next step when upgrade already in progress", func(t *testing.T) {
		//given
		installationClient := &installationMocks.Service{}
		installationClient.On("CheckInstallationState", mock.Anything).Return(installation.InstallationState{}, installation.InstallationError{ShortMessage: "upgrade in progress"})

		upgradeStep := NewUpgradeKymaStep(installationClient, nextStageName, 0)

		cluster := model.Cluster{Kubeconfig: util.StringPtr(kubeconfig)}

		//when
		result, err := upgradeStep.Run(cluster, model.Operation{}, logrus.New())

		//then
		require.NoError(t, err)
		assert.Equal(t, nextStageName, result.Stage)
	})

	t.Run("should return error when fail to trigger update", func(t *testing.T) {
		//given
		installationClient := &installationMocks.Service{}
		installationClient.On("CheckInstallationState", mock.Anything).Return(installation.InstallationState{State: "Installed"}, nil)
		installationClient.On("TriggerUpgrade", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("bad error"))

		upgradeStep := NewUpgradeKymaStep(installationClient, nextStageName, 0)

		cluster := model.Cluster{Kubeconfig: util.StringPtr(kubeconfig)}

		//when
		_, err := upgradeStep.Run(cluster, model.Operation{}, logrus.New())

		//then
		require.Error(t, err)
	})

	t.Run("should return next step when upgrade successfully triggered", func(t *testing.T) {
		//given
		installationClient := &installationMocks.Service{}
		installationClient.On("CheckInstallationState", mock.Anything).Return(installation.InstallationState{State: "Installed"}, nil)
		installationClient.On("TriggerUpgrade", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		upgradeStep := NewUpgradeKymaStep(installationClient, nextStageName, 0)

		cluster := model.Cluster{Kubeconfig: util.StringPtr(kubeconfig)}

		//when
		result, err := upgradeStep.Run(cluster, model.Operation{}, logrus.New())

		//then
		require.NoError(t, err)
		assert.Equal(t, nextStageName, result.Stage)
	})

	t.Run("should return next step when installation state in progress", func(t *testing.T) {
		//given
		installationClient := &installationMocks.Service{}
		installationClient.On("CheckInstallationState", mock.Anything).Return(installation.InstallationState{State: "In Progress"}, nil)

		upgradeStep := NewUpgradeKymaStep(installationClient, nextStageName, 0)

		cluster := model.Cluster{Kubeconfig: util.StringPtr(kubeconfig)}

		//when
		result, err := upgradeStep.Run(cluster, model.Operation{}, logrus.New())

		//then
		require.NoError(t, err)
		assert.Equal(t, nextStageName, result.Stage)
	})
}
