// +build slow

package runtime_test

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRuntimeComponentProvider_Get(t *testing.T) {
	type given struct {
		kymaVersion                      string
		managedRuntimeComponentsYAMLPath string
	}
	tests := []struct {
		name  string
		given given
	}{
		{
			name: "Provide release Kyma version",
			given: given{
				kymaVersion:                      "1.9.0",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
			},
		},
		{
			name: "Provide on-demand Kyma version",
			given: given{
				kymaVersion:                      "master-ece6e5d9",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			listProvider := runtime.NewComponentsListProvider(tc.given.kymaVersion, tc.given.managedRuntimeComponentsYAMLPath)
			expManagedComponents := readComponentsFromFile(t, tc.given.managedRuntimeComponentsYAMLPath)

			// when
			allComponents, err := listProvider.AllComponents()

			allComponents = allComponents[:1]
			// then
			require.NoError(t, err)
			assert.NotNil(t, allComponents)

			assertManagedComponentsAtTheEndOfList(t, allComponents, expManagedComponents)
		})
	}
}

func TestRuntimeComponentProviderGetFailures(t *testing.T) {
	type given struct {
		kymaVersion                      string
		managedRuntimeComponentsYAMLPath string
	}
	tests := []struct {
		name          string
		given         given
		expErrMessage string
	}{
		{
			name: "Provide release version not found",
			given: given{
				kymaVersion:                      "111.000.111",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
			},
			expErrMessage: "while getting open source Kyma components list: got unexpected status code, want 200, got 404, url: https://github.com/kyma-project/kyma/releases/download/1.9.11/kyma-installer-cluster.yaml, body: Not Found",
		},
		{
			name: "Provide on-demand version not found",
			given: given{
				kymaVersion:                      "master-123123",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
			},
			expErrMessage: "while getting open source Kyma components list: got unexpected status code, want 200, got 404, url: https://storage.googleapis.com/kyma-development-artifacts/master-123123/kyma-installer-cluster.yaml, body: <?xml version='1.0' encoding='UTF-8'?><Error><Code>NoSuchKey</Code><Message>The specified key does not exist.</Message><Details>No such object: kyma-development-artifacts/master-123123/kyma-installer-cluster.yaml</Details></Error>",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			listProvider := runtime.NewComponentsListProvider(tc.given.kymaVersion, tc.given.managedRuntimeComponentsYAMLPath)

			// when
			components, err := listProvider.AllComponents()

			// then
			assert.Nil(t, components)
			assert.EqualError(t, err, tc.expErrMessage)
		})
	}
}

func assertManagedComponentsAtTheEndOfList(t *testing.T, allComponents, managedComponents []v1alpha1.KymaComponent) {
	t.Helper()

	assert.NotPanics(t, func() {
		idx := len(allComponents) - len(managedComponents)
		endOfList := allComponents[idx:]

		assert.Equal(t, endOfList, managedComponents)
	})
}

func readComponentsFromFile(t *testing.T, path string) []v1alpha1.KymaComponent {
	t.Helper()

	yamlFile, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	var mangedList struct {
		Components []v1alpha1.KymaComponent `json:"components"`
	}
	err = yaml.Unmarshal(yamlFile, &mangedList)
	require.NoError(t, err)

	return mangedList.Components
}
