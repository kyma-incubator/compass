package installation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/artifacts"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"
	"k8s.io/client-go/rest"

	"github.com/kyma-incubator/hydroform/install/installation"
)

const (
	tillerYAML    = "tillerYAML"
	installerYAML = "installerYAML"
)

func TestInstallationService_InstallKyma(t *testing.T) {

	kymaVersion := "1.7.0"

	t.Run("should install Kyma", func(t *testing.T) {
		// given
		releaseArtifacts := artifacts.ReleaseArtifacts{Version: kymaVersion, TillerYAML: tillerYAML, InstallerYAML: installerYAML}

		artifactsProvider := &mocks.ArtifactsProvider{}
		artifactsProvider.On("GetArtifacts", kymaVersion).Return(releaseArtifacts, nil)

		expectedInstallation := installation.Installation{
			TillerYaml:    tillerYAML,
			InstallerYaml: installerYAML,
			Configuration: installation.Configuration{},
		}

		stateChannel := make(chan installation.InstallationState)
		errChannel := make(chan error)

		installationHandlerConstructor := newMockInstallerHandler(t, expectedInstallation, stateChannel, errChannel)
		installationSvc := NewInstallationService(10*time.Minute, artifactsProvider, installationHandlerConstructor)

		go func() {
			stateChannel <- installation.InstallationState{State: "Installed"}
			close(errChannel)
			close(stateChannel)
		}()

		// when
		err := installationSvc.InstallKyma(nil, kymaVersion)

		// then
		require.NoError(t, err)
	})

	// TODO - error tests
	t.Run("should install Kyma", func(t *testing.T) {
		// given
		releaseArtifacts := artifacts.ReleaseArtifacts{Version: kymaVersion, TillerYAML: tillerYAML, InstallerYAML: installerYAML}

		artifactsProvider := &mocks.ArtifactsProvider{}
		artifactsProvider.On("GetArtifacts", kymaVersion).Return(releaseArtifacts, nil)

		expectedInstallation := installation.Installation{
			TillerYaml:    tillerYAML,
			InstallerYaml: installerYAML,
			Configuration: installation.Configuration{},
		}

		stateChannel := make(chan installation.InstallationState)
		errChannel := make(chan error)

		installationHandlerConstructor := newMockInstallerHandler(t, expectedInstallation, stateChannel, errChannel)
		installationSvc := NewInstallationService(10*time.Minute, artifactsProvider, installationHandlerConstructor)

		go func() {
			stateChannel <- installation.InstallationState{State: "Installed"}
			close(stateChannel)
			close(errChannel)
		}()

		// when
		err := installationSvc.InstallKyma(nil, kymaVersion)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error when failed to trigger installation", func(t *testing.T) {
		// given
		releaseArtifacts := artifacts.ReleaseArtifacts{Version: kymaVersion, TillerYAML: tillerYAML, InstallerYAML: installerYAML}

		artifactsProvider := &mocks.ArtifactsProvider{}
		artifactsProvider.On("GetArtifacts", kymaVersion).Return(releaseArtifacts, nil)

		installationHandlerConstructor := newErrorInstallerHandler(t, nil, errors.New("error"))
		installationSvc := NewInstallationService(10*time.Minute, artifactsProvider, installationHandlerConstructor)

		// when
		err := installationSvc.InstallKyma(nil, kymaVersion)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to prepare installation", func(t *testing.T) {
		// given
		releaseArtifacts := artifacts.ReleaseArtifacts{Version: kymaVersion, TillerYAML: tillerYAML, InstallerYAML: installerYAML}

		artifactsProvider := &mocks.ArtifactsProvider{}
		artifactsProvider.On("GetArtifacts", kymaVersion).Return(releaseArtifacts, nil)

		installationHandlerConstructor := newErrorInstallerHandler(t, errors.New("error"), nil)
		installationSvc := NewInstallationService(10*time.Minute, artifactsProvider, installationHandlerConstructor)

		// when
		err := installationSvc.InstallKyma(nil, kymaVersion)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to get artifacts", func(t *testing.T) {
		// given
		artifactsProvider := &mocks.ArtifactsProvider{}
		artifactsProvider.On("GetArtifacts", kymaVersion).Return(artifacts.ReleaseArtifacts{}, errors.New("error"))

		installationSvc := NewInstallationService(10*time.Minute, artifactsProvider, nil)

		// when
		err := installationSvc.InstallKyma(nil, kymaVersion)

		// then
		require.Error(t, err)
	})

}

type installerMock struct {
	t                    *testing.T
	expectedInstallation installation.Installation
	stateChannel         chan installation.InstallationState
	errorChannel         chan error
}

func (i installerMock) PrepareInstallation(installation installation.Installation) error {
	assert.Equal(i.t, i.expectedInstallation, installation)
	return nil
}

func (i installerMock) StartInstallation(context context.Context) (<-chan installation.InstallationState, <-chan error, error) {
	assert.NotEmpty(i.t, context)
	return i.stateChannel, i.errorChannel, nil
}

func newMockInstallerHandler(t *testing.T, expectedInstallation installation.Installation, stateChan chan installation.InstallationState, errChan chan error) InstallationHandler {
	return func(config *rest.Config, option ...installation.InstallationOption) (installer installation.Installer, e error) {
		return installerMock{
			t:                    t,
			expectedInstallation: expectedInstallation,
			stateChannel:         stateChan,
			errorChannel:         errChan,
		}, nil
	}
}

type errorInstallerMock struct {
	t                        *testing.T
	prepareInstallationError error
	startInstallationError   error
}

func newErrorInstallerHandler(t *testing.T, prepareErr, startErr error) InstallationHandler {
	return func(config *rest.Config, option ...installation.InstallationOption) (installer installation.Installer, e error) {
		return errorInstallerMock{
			t:                        t,
			prepareInstallationError: prepareErr,
			startInstallationError:   startErr,
		}, nil
	}
}

func (i errorInstallerMock) PrepareInstallation(installation installation.Installation) error {
	return i.prepareInstallationError
}

func (i errorInstallerMock) StartInstallation(context context.Context) (<-chan installation.InstallationState, <-chan error, error) {
	return nil, nil, i.startInstallationError
}
