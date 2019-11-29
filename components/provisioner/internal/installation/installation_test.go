package installation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"k8s.io/client-go/rest"

	"github.com/kyma-incubator/hydroform/install/installation"
)

const (
	tillerYAML    = "tillerYAML"
	installerYAML = "installerYAML"

	installErrFailureThreshold = 5
)

const (
	kubeconfig = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUU3VENDQXRpZ0F3SUJBZ0lVY2x0eEd2ZTl4b01Kd3F5b1JEUGZULzV3aFZjd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0lERWVNQndHQTFVRUF3d1ZNVEEwTGpFNU9TNDJNeTR4TXpBdWVHbHdMbWx2TUI0WERURTVNVEV3TnpBMwpOVGswTUZvWERURTVNVEl3TnpBM05UazBNRm93SURFZU1Cd0dBMVVFQXd3Vk1UQTBMakU1T1M0Mk15NHhNekF1CmVHbHdMbWx2TUlJQ0hqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0Fnc0FNSUlDQmdLQ0FmMFp4R1RDeHZKbFBlV2wKcFNxekpxcXg0TkxYUGpRQjlFLzMwdi9GZTUxS1NkRWdlSTBRei9RUzR2eU9hYThzMEJWeGo5K2IwUkpUTjlKawo2bWwvSStxQTFBTDRWYzNrN1NxWnZnYnk5YVZ1bVZ4THpxK0tvdFJPNnhXSG5ubUpXSkxhRnlOTmgwd3hOWUtGCk1SWlZtbzNQTUp3WU9FaGdkUHdPaDlXQW12V3dWeWRaWmk4RWxUY2U2MEVkdG1RT0VPY0pVRmhTUThuZ1RnMWsKSHJjbXROalBlRm1lbWZDeVFZQm53NjI0WEVBVUV5OGZtSVgxMldwSFViSkhTYiswaGhsSHE3NmwyOVdUMWF5TwowTVI2SlJ2L1VuR0wxZGtDWU9udmxHaW1QTmhmU2I3SDNOS0xXaW5tYmJlSG04THlYWW1RSGZBYW9aTXRDWjJsClNUYXN6VVdVQjZTNVBsQTI5Wkd1YjdrZmNmbzloR05jdG8zSU1qVTBkdlA0NyszOENTV0pvb3FkakQrV2pqbUMKMm12czFEMU9vZzZkSWpOZ3hYZC9icU0rL0VKekFsdU4xNng3SGVVbE5BaCs1WlBkaVE2Q3U0M2c4OG55VTBwYwp3S0owU2dMQ0xSUUZsZ1Q0N24vLzBsR09wdXNtbTRGY2RPenI1RFJ4WllnZEpaZm10S3I0VmdWaUtaanZxdkNZCnBrVkhoUXh6L2N3SmpsbHJFWkxKeld0bFFmNTFjL0FqNnRrbzRWbmpaeVRyb04wZ3ZCMVdlV3cwb1lxdTZYZHYKSjFxdVl1K1RZdWd2N2h0ME5DUjNpQ0tzWXZQQjRKekZSUTl1QjZWVXVCOFZWNGlrYTdvcDlNVkNOZzE5bzNXbQpaNVNKYmNDSVZackxQVFJmeG9raXpYRWVaUUlEQVFBQm95WXdKREFpQmdOVkhSRUVHekFaZ2hjcUxqRXdOQzR4Ck9Ua3VOak11TVRNd0xuaHBjQzVwYnpBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQWY0QUY0SmwyL2JsTHdwUUlTa0YKUDlKWnF2d3lSdnBGeit3cnFTU3JhQWNlcEdYTDRKaDBWblBYUUUvbmdMWW5kR3Y3M0crWWZ5SnVZd2FjakRVeQpvNWsreWwrSGpsSWZFM3NqK1haYXN1Zk9GNGMrbytadDV3MlBHb0NNMldsMTVtNDFYYWlTWXJ3ZlRaOUhEdVJiClk1OENwT0hQZm13WVZjbkpnMmJ2ZHJBZjJwWk1ZRTQ5bHN1ZzVIb25UUTE0WHJpR0NNK3BlQXkxYWJTY2xnTTQKa1JhTmhYcVBrZ2RsQ0hJQWlpZzBqZTJ1OUZTK3ExWUVGeWFOWVpkM0FhWTZVTlZZN1A0TVJlczVmajFDZnlDeQo1dmNqOTgrcGhBSTNiclNUOUswN21iUFNKbkdkcHlrWVJOdGhUYnlaNzd5N1JmYWZQUDRZU0dIRHZwb2QzRXlKCk1VR3dKQXJGOUIrZTFENGV3WjdmQ0xyRkN6U3h5Z1NaaTlGRVVEOVlJbFBlclVpZzBOK2EyV2NnNTFVZXhFY20KS3hoMjJVUTZxNGNOYzN3SFNmVzFjNWl1NVhwNG10R09UcDBuZkNNbmtrSHJDN0RsSUpya0dTYlBCVUF2T0hTOQpyRGFsQ0ZiRERFb0pjVmtlb2dNaUY5U2hWSnJWM0RNdWppTVdJTzc4d3JHZTZHTnpwWW16OWMxTDRUNTBhVm5HCmtIVFp6S3JIUFRMRnZGNkdJTlJET3U2dVdISEVnZnV6ZGlxQnM3MjNPLzE0eHJ1QnRqUGtMdDd2Y0R5cEZ5RFUKc3dYa2VOS2pIa2w5TElMZHZmZVVxazBQNkljSGNLWFpheXF2Ykpqa21MTHhIQUxGaXpVTUhWSnN6NVkzVW1RVwpTZkpjeDdlb3BhMTU5ZllEMHE3VExpST0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://192.168.64.4:8443
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
  user:
    client-certificate: /Users/i351738/.minikube/client.crt
    client-key: /Users/i351738/.minikube/client.key
`

	runtimeId = "abcd-efgh"
)

func TestInstallationService_InstallKyma(t *testing.T) {

	kymaVersion := "1.7.0"
	kymaRelease := model.Release{Version: kymaVersion, TillerYAML: tillerYAML, InstallerYAML: installerYAML}

	for _, testCase := range []struct {
		description      string
		installationMock func(chan installation.InstallationState, chan error)
		shouldFail       bool
	}{
		{
			description: "should install Kyma successfully",
			installationMock: func(stateChan chan installation.InstallationState, errChannel chan error) {
				stateChan <- installation.InstallationState{State: "Installed"}
				close(errChannel)
				close(stateChan)
			},
			shouldFail: false,
		},
		{
			description: "should continue installation if error threshold not exceeded",
			installationMock: func(stateChan chan installation.InstallationState, errChannel chan error) {
				stateChan <- installation.InstallationState{State: "Installing"}
				errChannel <- installation.InstallationError{
					ErrorEntries: make([]installation.ErrorEntry, 2),
				}
				time.Sleep(1 * time.Second)
				close(stateChan)
				close(errChannel)
			},
			shouldFail: false,
		},
		{
			description: "should fail if error threshold exceeded",
			installationMock: func(stateChan chan installation.InstallationState, errChannel chan error) {
				stateChan <- installation.InstallationState{State: "Installing"}
				errChannel <- installation.InstallationError{
					ErrorEntries: make([]installation.ErrorEntry, 10),
				}
				time.Sleep(1 * time.Second)
				close(stateChan)
				close(errChannel)
			},
			shouldFail: true,
		},
		{
			description: "should fail if error different than installation error occurred",
			installationMock: func(stateChan chan installation.InstallationState, errChannel chan error) {
				stateChan <- installation.InstallationState{State: "Installing"}
				errChannel <- errors.New("error")
				time.Sleep(1 * time.Second)
				close(stateChan)
				close(errChannel)
			},
			shouldFail: true,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			expectedInstallation := installation.Installation{
				TillerYaml:    tillerYAML,
				InstallerYaml: installerYAML,
				Configuration: installation.Configuration{},
			}

			stateChannel := make(chan installation.InstallationState)
			errChannel := make(chan error)

			installationHandlerConstructor := newMockInstallerHandler(t, expectedInstallation, stateChannel, errChannel)
			installationSvc := NewInstallationService(10*time.Minute, installationHandlerConstructor, installErrFailureThreshold)

			go testCase.installationMock(stateChannel, errChannel)

			// when
			err := installationSvc.InstallKyma(runtimeId, kubeconfig, kymaRelease)

			// then
			if testCase.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("should return error when failed to trigger installation", func(t *testing.T) {
		// given
		installationHandlerConstructor := newErrorInstallerHandler(t, nil, errors.New("error"))
		installationSvc := NewInstallationService(10*time.Minute, installationHandlerConstructor, installErrFailureThreshold)

		// when
		err := installationSvc.InstallKyma(runtimeId, kubeconfig, kymaRelease)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to prepare installation", func(t *testing.T) {
		// given
		installationHandlerConstructor := newErrorInstallerHandler(t, errors.New("error"), nil)
		installationSvc := NewInstallationService(10*time.Minute, installationHandlerConstructor, installErrFailureThreshold)

		// when
		err := installationSvc.InstallKyma(runtimeId, kubeconfig, kymaRelease)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to parse kubeconfig", func(t *testing.T) {
		// given
		installationSvc := NewInstallationService(10*time.Minute, nil, installErrFailureThreshold)

		// when
		err := installationSvc.InstallKyma(runtimeId, "", kymaRelease)

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
