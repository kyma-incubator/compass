package installation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util/k8s"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	kymaSystemNamespace      = "kyma-system"
	kymaIntegrationNamespace = "kyma-integration"

	rafterSourceURL = "github.com/kyma-project/kyma.git//resources/rafter"
)

const (
	kubeconfig = `apiVersion: v1
clusters:
- cluster:
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
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURBRENDQWVpZ0F3SUJBZ0lCQWpBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwdGFXNXAKYTNWaVpVTkJNQjRYRFRFNU1URXhOekE0TXpBek1sb1hEVEl3TVRFeE56QTRNekF6TWxvd01URVhNQlVHQTFVRQpDaE1PYzNsemRHVnRPbTFoYzNSbGNuTXhGakFVQmdOVkJBTVREVzFwYm1scmRXSmxMWFZ6WlhJd2dnRWlNQTBHCkNTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFDNmY2SjZneElvL2cyMHArNWhybklUaUd5SDh0VW0KWGl1OElaK09UKyt0amd1OXRneXFnbnNsL0dDT1Q3TFo4ejdOVCttTEdKL2RLRFdBV3dvbE5WTDhxMzJIQlpyNwpDaU5hK3BBcWtYR0MzNlQ2NEQyRjl4TEtpVVpuQUVNaFhWOW1oeWVCempscTh1NnBjT1NrY3lJWHRtdU9UQUVXCmErWlp5UlhOY3BoYjJ0NXFUcWZoSDhDNUVDNUIrSm4rS0tXQ2Y1Nm5KZGJQaWduRXh4SFlaMm9TUEc1aXpkbkcKZDRad2d0dTA3NGttaFNtNXQzbjgyNmovK29tL25VeWdBQ24yNmR1K21aZzRPcWdjbUMrdnBYdUEyRm52bk5LLwo5NWErNEI3cGtNTER1bHlmUTMxcjlFcStwdHBkNUR1WWpldVpjS1Bxd3ZVcFUzWVFTRUxVUzBrUkFnTUJBQUdqClB6QTlNQTRHQTFVZER3RUIvd1FFQXdJRm9EQWRCZ05WSFNVRUZqQVVCZ2dyQmdFRkJRY0RBUVlJS3dZQkJRVUgKQXdJd0RBWURWUjBUQVFIL0JBSXdBREFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBQ3JnbExWemhmemZ2aFNvUgowdWNpNndBZDF6LzA3bW52MDRUNmQyTkpjRG80Uzgwa0o4VUJtRzdmZE5qMlJEaWRFbHRKRU1kdDZGa1E1TklOCk84L1hJdENiU0ZWYzRWQ1NNSUdPcnNFOXJDajVwb24vN3JxV3dCbllqYStlbUVYOVpJelEvekJGU3JhcWhud3AKTkc1SmN6bUg5ODRWQUhGZEMvZWU0Z2szTnVoV25rMTZZLzNDTTFsRkxlVC9Cbmk2K1M1UFZoQ0x3VEdmdEpTZgorMERzbzVXVnFud2NPd3A3THl2K3h0VGtnVmdSRU5RdTByU2lWL1F2UkNPMy9DWXdwRTVIRFpjalM5N0I4MW0yCmVScVBENnVoRjVsV3h4NXAyeEd1V2JRSkY0WnJzaktLTW1CMnJrUnR5UDVYV2xWZU1mR1VjbFdjc1gxOW91clMKaWpKSTFnPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBdW4raWVvTVNLUDROdEtmdVlhNXlFNGhzaC9MVkpsNHJ2Q0dmamsvdnJZNEx2YllNCnFvSjdKZnhnamsreTJmTSt6VS9waXhpZjNTZzFnRnNLSlRWUy9LdDlod1dhK3dvald2cVFLcEZ4Z3Qrayt1QTkKaGZjU3lvbEdad0JESVYxZlpvY25nYzQ1YXZMdXFYRGtwSE1pRjdacmprd0JGbXZtV2NrVnpYS1lXOXJlYWs2bgo0Ui9BdVJBdVFmaVovaWlsZ24rZXB5WFd6NG9KeE1jUjJHZHFFanh1WXMzWnhuZUdjSUxidE8rSkpvVXB1YmQ1Ci9OdW8vL3FKdjUxTW9BQXA5dW5idnBtWU9EcW9ISmd2cjZWN2dOaFo3NXpTdi9lV3Z1QWU2WkRDdzdwY24wTjkKYS9SS3ZxYmFYZVE3bUkzcm1YQ2o2c0wxS1ZOMkVFaEMxRXRKRVFJREFRQUJBb0lCQVFDTEVFa3pXVERkYURNSQpGb0JtVGhHNkJ1d0dvMGZWQ0R0TVdUWUVoQTZRTjI4QjB4RzJ3dnpZNGt1TlVsaG10RDZNRVo1dm5iajJ5OWk1CkVTbUxmU3VZUkxlaFNzaTVrR0cwb1VtR3RGVVQ1WGU3cWlHMkZ2bm9GRnh1eVg5RkRiN3BVTFpnMEVsNE9oVkUKTzI0Q1FlZVdEdXc4ZXVnRXRBaGJ3dG1ERElRWFdPSjcxUEcwTnZKRHIwWGpkcW1aeExwQnEzcTJkZTU2YmNjawpPYzV6dmtJNldrb0o1TXN0WkZpU3pVRDYzN3lIbjh2NGd3cXh0bHFoNWhGLzEwV296VmZqVGdWSG0rc01ZaU9SCmNIZ0dMNUVSbDZtVlBsTTQzNUltYnFnU1R2NFFVVGpzQjRvbVBsTlV5Yksvb3pPSWx3RjNPTkJjVVV6eDQ1cGwKSHVJQlQwZ1JBb0dCQU9SR2lYaVBQejdsay9Bc29tNHkxdzFRK2hWb3Yvd3ovWFZaOVVkdmR6eVJ1d3gwZkQ0QgpZVzlacU1hK0JodnB4TXpsbWxYRHJBMklYTjU3UEM3ZUo3enhHMEVpZFJwN3NjN2VmQUN0eDN4N0d0V2pRWGF2ClJ4R2xDeUZxVG9LY3NEUjBhQ0M0Um15VmhZRTdEY0huLy9oNnNzKys3U2tvRVMzNjhpS1RiYzZQQW9HQkFORW0KTHRtUmZieHIrOE5HczhvdnN2Z3hxTUlxclNnb2NmcjZoUlZnYlU2Z3NFd2pMQUs2ZHdQV0xWQmVuSWJ6bzhodApocmJHU1piRnF0bzhwS1Q1d2NxZlpKSlREQnQxYmhjUGNjWlRmSnFmc0VISXc0QW5JMVdRMlVzdzVPcnZQZWhsCmh0ek95cXdBSGZvWjBUTDlseTRJUHRqbXArdk1DQ2NPTHkwanF6NWZBb0dCQUlNNGpRT3hqSkN5VmdWRkV5WTMKc1dsbE9DMGdadVFxV3JPZnY2Q04wY1FPbmJCK01ZRlBOOXhUZFBLeC96OENkVyszT0syK2FtUHBGRUdNSTc5cApVdnlJdUxzTGZMZDVqVysyY3gvTXhaU29DM2Z0ZmM4azJMeXEzQ2djUFA5VjVQQnlUZjBwRU1xUWRRc2hrRG44CkRDZWhHTExWTk8xb3E5OTdscjhMY3A2L0FvR0FYNE5KZC9CNmRGYjRCYWkvS0lGNkFPQmt5aTlGSG9iQjdyVUQKbTh5S2ZwTGhrQk9yNEo4WkJQYUZnU09ENWhsVDNZOHZLejhJa2tNNUVDc0xvWSt4a1lBVEpNT3FUc3ZrOThFRQoyMlo3Qy80TE55K2hJR0EvUWE5Qm5KWDZwTk9XK1ErTWRFQTN6QzdOZ2M3U2U2L1ZuNThDWEhtUmpCeUVTSm13CnI3T1BXNDhDZ1lBVUVoYzV2VnlERXJxVDBjN3lIaXBQbU1wMmljS1hscXNhdC94YWtobENqUjZPZ2I5aGQvNHIKZm1wUHJmd3hjRmJrV2tDRUhJN01EdDJrZXNEZUhRWkFxN2xEdjVFT2k4ZG1uM0ZPNEJWczhCOWYzdm52MytmZwpyV2E3ZGtyWnFudU12cHhpSWlqOWZEak9XbzdxK3hTSFcxWWdSNGV2Q1p2NGxJU0FZRlViemc9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`

	runtimeId = "abcd-efgh"
)

func TestInstallationService_TriggerInstallation(t *testing.T) {

	kymaVersion := "1.7.0"
	kymaRelease := model.Release{Version: kymaVersion, TillerYAML: tillerYAML, InstallerYAML: installerYAML}

	globalConfig := fixGlobalConfig()
	componentsConfig := fixComponentsConfig()

	expectedInstallation := installation.Installation{
		TillerYaml:    tillerYAML,
		InstallerYaml: installerYAML,
		Configuration: fixInstallationConfig(),
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(kubeconfig))
	require.NoError(t, err)

	t.Run("should trigger installation", func(t *testing.T) {
		installationHandlerConstructor := newMockInstallerHandler(t, expectedInstallation, nil, nil)
		installationSvc := NewInstallationService(10*time.Minute, installationHandlerConstructor, installErrFailureThreshold)

		// when
		err := installationSvc.TriggerInstallation(k8sConfig, kymaRelease, globalConfig, componentsConfig)

		// then
		require.NoError(t, err)
	})

}

func TestInstallationService_InstallKyma(t *testing.T) {

	kymaVersion := "1.7.0"
	kymaRelease := model.Release{Version: kymaVersion, TillerYAML: tillerYAML, InstallerYAML: installerYAML}

	globalConfig := fixGlobalConfig()
	componentsConfig := fixComponentsConfig()

	defaultExpectedInstallation := installation.Installation{
		TillerYaml:    tillerYAML,
		InstallerYaml: installerYAML,
		Configuration: installation.Configuration{
			Configuration:          make([]installation.ConfigEntry, 0),
			ComponentConfiguration: make([]installation.ComponentConfiguration, 0),
		},
	}

	for _, testCase := range []struct {
		description          string
		installationMock     func(chan installation.InstallationState, chan error)
		globalConfig         model.Configuration
		componentsConfig     []model.KymaComponentConfig
		shouldFail           bool
		expectedInstallation installation.Installation
	}{
		{
			description: "should install Kyma successfully",
			installationMock: func(stateChan chan installation.InstallationState, errChannel chan error) {
				stateChan <- installation.InstallationState{State: "Installed"}
				close(errChannel)
				close(stateChan)
			},
			globalConfig:     fixGlobalConfig(),
			componentsConfig: fixComponentsConfig(),
			shouldFail:       false,
			expectedInstallation: installation.Installation{
				TillerYaml:    tillerYAML,
				InstallerYaml: installerYAML,
				Configuration: fixInstallationConfig(),
			},
		},
		{
			description: "should install Kyma with empty configs",
			installationMock: func(stateChan chan installation.InstallationState, errChannel chan error) {
				stateChan <- installation.InstallationState{State: "Installed"}
				close(errChannel)
				close(stateChan)
			},
			globalConfig:         model.Configuration{},
			componentsConfig:     []model.KymaComponentConfig{},
			shouldFail:           false,
			expectedInstallation: defaultExpectedInstallation,
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
			globalConfig:         model.Configuration{},
			componentsConfig:     []model.KymaComponentConfig{},
			shouldFail:           false,
			expectedInstallation: defaultExpectedInstallation,
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
			globalConfig:         model.Configuration{},
			componentsConfig:     []model.KymaComponentConfig{},
			shouldFail:           true,
			expectedInstallation: defaultExpectedInstallation,
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
			globalConfig:         model.Configuration{},
			componentsConfig:     []model.KymaComponentConfig{},
			shouldFail:           true,
			expectedInstallation: defaultExpectedInstallation,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			stateChannel := make(chan installation.InstallationState)
			errChannel := make(chan error)

			installationHandlerConstructor := newMockInstallerHandler(t, testCase.expectedInstallation, stateChannel, errChannel)
			installationSvc := NewInstallationService(10*time.Minute, installationHandlerConstructor, installErrFailureThreshold)

			go testCase.installationMock(stateChannel, errChannel)

			// when
			err := installationSvc.InstallKyma(runtimeId, kubeconfig, kymaRelease, testCase.globalConfig, testCase.componentsConfig)

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
		err := installationSvc.InstallKyma(runtimeId, kubeconfig, kymaRelease, globalConfig, componentsConfig)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to prepare installation", func(t *testing.T) {
		// given
		installationHandlerConstructor := newErrorInstallerHandler(t, errors.New("error"), nil)
		installationSvc := NewInstallationService(10*time.Minute, installationHandlerConstructor, installErrFailureThreshold)

		// when
		err := installationSvc.InstallKyma(runtimeId, kubeconfig, kymaRelease, globalConfig, componentsConfig)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to parse kubeconfig", func(t *testing.T) {
		// given
		installationSvc := NewInstallationService(10*time.Minute, nil, installErrFailureThreshold)

		// when
		err := installationSvc.InstallKyma(runtimeId, "", kymaRelease, globalConfig, componentsConfig)

		// then
		require.Error(t, err)
	})

}

func Test_getInstallationCRModificationFunc(t *testing.T) {

	newInstallationCR := func() *v1alpha1.Installation {
		return &v1alpha1.Installation{
			ObjectMeta: v1.ObjectMeta{
				Name:      "kyma-installation",
				Namespace: "default",
			},
			Spec: v1alpha1.InstallationSpec{},
		}
	}

	t.Run("should create modification func", func(t *testing.T) {
		// given
		componentsConfig := fixComponentsConfig()
		installationCR := newInstallationCR()

		// when
		modificationFunc := GetInstallationCRModificationFunc(componentsConfig)

		modificationFunc(installationCR)

		// then
		require.Equal(t, 4, len(installationCR.Spec.Components))
		assertComponent(t, "cluster-essentials", kymaSystemNamespace, nil, installationCR.Spec.Components[0])
		assertComponent(t, "core", kymaSystemNamespace, nil, installationCR.Spec.Components[1])
		assertComponent(t, "rafter", kymaSystemNamespace, &v1alpha1.ComponentSource{URL: rafterSourceURL}, installationCR.Spec.Components[2])
		assertComponent(t, "application-connector", kymaIntegrationNamespace, nil, installationCR.Spec.Components[3])
	})

	t.Run("should have no components if configuration is empty", func(t *testing.T) {
		// given
		installationCR := newInstallationCR()

		// when
		modificationFunc := GetInstallationCRModificationFunc(nil)

		modificationFunc(installationCR)

		// then
		require.Equal(t, 0, len(installationCR.Spec.Components))
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

func (i installerMock) PrepareUpgrade(installation installation.Installation) error {
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
	t                      *testing.T
	prepareError           error
	startInstallationError error
}

func newErrorInstallerHandler(t *testing.T, prepareErr, startErr error) InstallationHandler {
	return func(config *rest.Config, option ...installation.InstallationOption) (installer installation.Installer, e error) {
		return errorInstallerMock{
			t:                      t,
			prepareError:           prepareErr,
			startInstallationError: startErr,
		}, nil
	}
}

func (i errorInstallerMock) PrepareInstallation(installation installation.Installation) error {
	return i.prepareError
}

func (i errorInstallerMock) StartInstallation(context context.Context) (<-chan installation.InstallationState, <-chan error, error) {
	return nil, nil, i.startInstallationError
}

func (i errorInstallerMock) PrepareUpgrade(installation installation.Installation) error {
	return i.prepareError
}

func assertComponent(t *testing.T, expectedName, expectedNamespace string, expectedSource *v1alpha1.ComponentSource, component v1alpha1.KymaComponent) {
	assert.Equal(t, expectedName, component.Name)
	assert.Equal(t, expectedNamespace, component.Namespace)
	assert.Equal(t, expectedSource, component.Source)
}

func fixInstallationConfig() installation.Configuration {
	return installation.Configuration{
		Configuration: []installation.ConfigEntry{
			fixInstallationConfigEntry("global.config.key", "globalValue", false),
			fixInstallationConfigEntry("global.config.key2", "globalValue2", false),
			fixInstallationConfigEntry("global.secret.key", "globalSecretValue", true),
		},
		ComponentConfiguration: []installation.ComponentConfiguration{
			{
				Component:     "cluster-essentials",
				Configuration: make([]installation.ConfigEntry, 0),
			},
			{
				Component: "core",
				Configuration: []installation.ConfigEntry{
					fixInstallationConfigEntry("test.config.key", "value", false),
					fixInstallationConfigEntry("test.config.key2", "value2", false),
				},
			},
			{
				Component:     "rafter",
				Configuration: make([]installation.ConfigEntry, 0),
			},
			{
				Component: "application-connector",
				Configuration: []installation.ConfigEntry{
					fixInstallationConfigEntry("test.config.key", "value", false),
					fixInstallationConfigEntry("test.secret.key", "secretValue", true),
				},
			},
		},
	}
}

func fixComponentsConfig() []model.KymaComponentConfig {
	return []model.KymaComponentConfig{
		{
			ID:            "id",
			KymaConfigID:  "id",
			Component:     "cluster-essentials",
			Namespace:     kymaSystemNamespace,
			Configuration: model.Configuration{ConfigEntries: make([]model.ConfigEntry, 0, 0)},
		},
		{
			ID:           "id",
			KymaConfigID: "id",
			Component:    "core",
			Namespace:    kymaSystemNamespace,
			Configuration: model.Configuration{
				ConfigEntries: []model.ConfigEntry{
					model.NewConfigEntry("test.config.key", "value", false),
					model.NewConfigEntry("test.config.key2", "value2", false),
				},
			},
		},
		{
			ID:            "id",
			KymaConfigID:  "id",
			Component:     "rafter",
			Namespace:     kymaSystemNamespace,
			SourceURL:     rafterSourceURL,
			Configuration: model.Configuration{ConfigEntries: make([]model.ConfigEntry, 0, 0)},
		},
		{
			ID:           "id",
			KymaConfigID: "id",
			Component:    "application-connector",
			Namespace:    kymaIntegrationNamespace,
			Configuration: model.Configuration{
				ConfigEntries: []model.ConfigEntry{
					model.NewConfigEntry("test.config.key", "value", false),
					model.NewConfigEntry("test.secret.key", "secretValue", true),
				},
			},
		},
	}
}

func fixGlobalConfig() model.Configuration {
	return model.Configuration{ConfigEntries: []model.ConfigEntry{
		model.NewConfigEntry("global.config.key", "globalValue", false),
		model.NewConfigEntry("global.config.key2", "globalValue2", false),
		model.NewConfigEntry("global.secret.key", "globalSecretValue", true),
	}}
}

func fixInstallationConfigEntry(key, val string, secret bool) installation.ConfigEntry {
	return installation.ConfigEntry{
		Key:    key,
		Value:  val,
		Secret: secret,
	}
}
