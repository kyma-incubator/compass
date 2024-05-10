package apiclient_test

import (
	"context"
	systemfielddiscoveryengine "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/apiclient"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	appID             = "appID"
	tenantID          = "tenantID"
	fakeEndpoint      = "http://fake-endpoint"
	discoveryEndpoint = "/discover"
)

func TestSFDClient_Discover(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	var unsupportedRegistry systemfielddiscoveryengine.SystemFieldDiscoveryRegistry = "unsupported"

	testCases := []struct {
		Name          string
		Endpoint      string
		SFDClientFunc func() (*http.Client, func(), string)
		Registry      systemfielddiscoveryengine.SystemFieldDiscoveryRegistry
		ExpectedErr   error
	}{
		{
			Name:          "Success calling discover api",
			Endpoint:      discoveryEndpoint,
			SFDClientFunc: fixSFDClient,
			Registry:      systemfielddiscoveryengine.SystemFieldDiscoverySaaSRegistry,
			ExpectedErr:   nil,
		},
		{
			Name:          "Error when status code is 404",
			Endpoint:      discoveryEndpoint,
			SFDClientFunc: fixErrorNotFoundSFDClient,
			Registry:      systemfielddiscoveryengine.SystemFieldDiscoverySaaSRegistry,
			ExpectedErr:   errors.New("received unexpected status code"),
		},
		{
			Name:          "Error when calling aggregate api",
			Endpoint:      fakeEndpoint,
			SFDClientFunc: fixSFDClient,
			Registry:      systemfielddiscoveryengine.SystemFieldDiscoverySaaSRegistry,
			ExpectedErr:   errors.New("while executing request to system field discovery"),
		},
		{
			Name:          "Error - unsupported registry",
			Endpoint:      discoveryEndpoint,
			SFDClientFunc: fixSFDClient,
			Registry:      unsupportedRegistry,
			ExpectedErr:   errors.New("unsupported registry"),
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			mockClient, mockServerCloseFn, URL := test.SFDClientFunc()
			defer mockServerCloseFn()

			SFDClientConfig := apiclient.SystemFieldDiscoveryEngineClientConfig{
				ClientTimeout: time.Second * 30,
				SystemFieldDiscoveryEngineSaaSRegistryAPI: URL + test.Endpoint,
				SkipSSLValidation:                         true,
			}
			client := apiclient.NewSystemFieldDiscoveryEngineClient(SFDClientConfig)
			client.SetHTTPClient(mockClient)

			// WHEN
			err := client.Discover(ctx, appID, tenantID, test.Registry)

			// THEN
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func fixSFDClient() (*http.Client, func(), string) {
	mux := http.NewServeMux()

	mux.HandleFunc(discoveryEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}

func fixErrorNotFoundSFDClient() (*http.Client, func(), string) {
	mux := http.NewServeMux()
	mux.HandleFunc(discoveryEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}
