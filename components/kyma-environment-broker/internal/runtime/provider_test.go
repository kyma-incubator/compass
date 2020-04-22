package runtime_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path"
	"testing"

	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRuntimeComponentProviderGetSuccess(t *testing.T) {
	type given struct {
		kymaVersion                      string
		managedRuntimeComponentsYAMLPath string
	}
	tests := []struct {
		name               string
		given              given
		expectedRequestURL string
	}{
		{
			name: "Provide release Kyma version",
			given: given{
				kymaVersion:                      "1.9.0",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
			},
			expectedRequestURL: "https://github.com/kyma-project/kyma/releases/download/1.9.0/kyma-installer-cluster.yaml",
		},
		{
			name: "Provide on-demand Kyma version",
			given: given{
				kymaVersion:                      "master-ece6e5d9",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
			},
			expectedRequestURL: "https://storage.googleapis.com/kyma-development-artifacts/master-ece6e5d9/kyma-installer-cluster.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			installerYAML := readKymaInstallerClusterYAMLFromFile(t)
			fakeHTTPClient := newTestClient(t, installerYAML, http.StatusOK)

			listProvider := runtime.NewComponentsListProvider(tc.given.managedRuntimeComponentsYAMLPath).WithHTTPClient(fakeHTTPClient)

			expManagedComponents := readManagedComponentsFromFile(t, tc.given.managedRuntimeComponentsYAMLPath)

			// when
			allComponents, err := listProvider.AllComponents(tc.given.kymaVersion)

			// then
			require.NoError(t, err)
			assert.NotNil(t, allComponents)

			assert.Equal(t, tc.expectedRequestURL, fakeHTTPClient.RequestURL)
			assertManagedComponentsAtTheEndOfList(t, allComponents, expManagedComponents)
		})
	}
}

func TestRuntimeComponentProviderGetFailures(t *testing.T) {
	type given struct {
		kymaVersion                      string
		managedRuntimeComponentsYAMLPath string
		httpErrMessage                   string
	}
	tests := []struct {
		name             string
		given            given
		returnStatusCode int
		tempError        bool
		expErrMessage    string
	}{
		{
			name: "Provide release version not found",
			given: given{
				kymaVersion:                      "111.000.111",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
				httpErrMessage:                   "Not Found",
			},
			returnStatusCode: http.StatusNotFound,
			tempError:        false,
			expErrMessage:    "while getting open source kyma components: while checking response status code for Kyma components list: got unexpected status code, want 200, got 404, url: https://github.com/kyma-project/kyma/releases/download/111.000.111/kyma-installer-cluster.yaml, body: Not Found",
		},
		{
			name: "Provide on-demand version not found",
			given: given{
				kymaVersion:                      "master-123123",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
				httpErrMessage:                   "Not Found",
			},
			returnStatusCode: http.StatusNotFound,
			tempError:        false,
			expErrMessage:    "while getting open source kyma components: while checking response status code for Kyma components list: got unexpected status code, want 200, got 404, url: https://storage.googleapis.com/kyma-development-artifacts/master-123123/kyma-installer-cluster.yaml, body: Not Found",
		},
		{
			name: "Provide on-demand version not found, temporary server error",
			given: given{
				kymaVersion:                      "master-123123",
				managedRuntimeComponentsYAMLPath: path.Join("testdata", "managed-runtime-components.yaml"),
				httpErrMessage:                   "Internal Server Error",
			},
			returnStatusCode: http.StatusInternalServerError,
			tempError:        true,
			expErrMessage:    "while getting open source kyma components: while checking response status code for Kyma components list: got unexpected status code, want 200, got 500, url: https://storage.googleapis.com/kyma-development-artifacts/master-123123/kyma-installer-cluster.yaml, body: Internal Server Error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			fakeHTTPClient := newTestClient(t, tc.given.httpErrMessage, tc.returnStatusCode)

			listProvider := runtime.NewComponentsListProvider(tc.given.managedRuntimeComponentsYAMLPath).
				WithHTTPClient(fakeHTTPClient)

			// when
			components, err := listProvider.AllComponents(tc.given.kymaVersion)

			// then
			assert.Nil(t, components)
			assert.EqualError(t, err, tc.expErrMessage)
			assert.Equal(t, tc.tempError, kebError.IsTemporaryError(err))
		})
	}
}

type HTTPFakeClient struct {
	t                *testing.T
	installerContent string
	code             int

	RequestURL string
}

func (f *HTTPFakeClient) Do(req *http.Request) (*http.Response, error) {
	f.RequestURL = req.URL.String()

	return &http.Response{
		StatusCode: f.code,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(f.installerContent))),
		Request:    req,
	}, nil
}

func newTestClient(t *testing.T, installerContent string, code int) *HTTPFakeClient {
	return &HTTPFakeClient{
		t:                t,
		code:             code,
		installerContent: installerContent,
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

func readKymaInstallerClusterYAMLFromFile(t *testing.T) string {
	t.Helper()

	filename := path.Join("testdata", "kyma-installer-cluster.yaml")
	yamlFile, err := ioutil.ReadFile(filename)
	require.NoError(t, err)

	return string(yamlFile)
}

func readManagedComponentsFromFile(t *testing.T, path string) []v1alpha1.KymaComponent {
	t.Helper()

	yamlFile, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	var managedList struct {
		Components []v1alpha1.KymaComponent `json:"components"`
	}
	err = yaml.Unmarshal(yamlFile, &managedList)
	require.NoError(t, err)

	return managedList.Components
}
