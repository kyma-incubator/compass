package apiclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/apiclient"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	tenantID     = "tenantID"
	fakeEndpoint = "http://fake-endpoint"
	syncEndpoint = "/sync"
)

func TestSystemFetcherClient_Sync(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testCases := []struct {
		Name                    string
		Endpoint                string
		SystemFetcherClientFunc func() (*http.Client, func(), string)
		ExpectedErr             error
	}{
		{
			Name:                    "Success calling sync api",
			Endpoint:                syncEndpoint,
			SystemFetcherClientFunc: fixSystemFetcherClient,
			ExpectedErr:             nil,
		},
		{
			Name:                    "Expected error when status code is 406",
			Endpoint:                syncEndpoint,
			SystemFetcherClientFunc: fixErrorNotAcceptableSystemFetcherClient,
			ExpectedErr:             errors.New("on-demand system sync is disabled"),
		},
		{
			Name:                    "Error when status code is 404",
			Endpoint:                syncEndpoint,
			SystemFetcherClientFunc: fixErrorNotFoundSystemFetcherClient,
			ExpectedErr:             errors.New("received unexpected status code"),
		},
		{
			Name:                    "Error when calling sync api",
			Endpoint:                fakeEndpoint,
			SystemFetcherClientFunc: fixSystemFetcherClient,
			ExpectedErr:             errors.New("while executing request to system fetcher"),
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			mockClient, mockServerCloseFn, URL := test.SystemFetcherClientFunc()
			defer mockServerCloseFn()

			SystemFetcherClientConfig := apiclient.SystemFetcherSyncClientConfig{
				ClientTimeout:        time.Second * 30,
				SystemFetcherSyncAPI: URL + test.Endpoint,
				SkipSSLValidation:    true,
			}
			client := apiclient.NewSystemFetcherClient(SystemFetcherClientConfig)
			client.SetHTTPClient(mockClient)

			// WHEN
			err := client.Sync(ctx, tenantID)

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

func fixSystemFetcherClient() (*http.Client, func(), string) {
	mux := http.NewServeMux()

	mux.HandleFunc(syncEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}

func fixErrorNotFoundSystemFetcherClient() (*http.Client, func(), string) {
	mux := http.NewServeMux()
	mux.HandleFunc(syncEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}

func fixErrorNotAcceptableSystemFetcherClient() (*http.Client, func(), string) {
	mux := http.NewServeMux()
	mux.HandleFunc(syncEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotAcceptable)
	})
	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}
