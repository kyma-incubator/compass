package apiclient_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/apiclient"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	appID             = "appID"
	appTemplateID     = "appTemplateID"
	fakeEndpoint      = "http://fake-endpoint"
	aggregateEndpoint = "/aggregate"
)

func TestORDClient_Aggregate(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testCases := []struct {
		Name          string
		Endpoint      string
		OrdClientFunc func() (*http.Client, func(), string)
		ExpectedErr   error
	}{
		{
			Name:          "Success calling aggregate api",
			Endpoint:      aggregateEndpoint,
			OrdClientFunc: fixOrdClient,
			ExpectedErr:   nil,
		},
		{
			Name:          "Error when status code is 404",
			Endpoint:      aggregateEndpoint,
			OrdClientFunc: fixErrorNotFoundOrdClient,
			ExpectedErr:   errors.New("received unexpected status code"),
		},
		{
			Name:          "Error when calling aggregate api",
			Endpoint:      fakeEndpoint,
			OrdClientFunc: fixOrdClient,
			ExpectedErr:   errors.New("while executing request to ord aggregator"),
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			mockClient, mockServerCloseFn, URL := test.OrdClientFunc()
			defer mockServerCloseFn()

			ORDClientConfig := apiclient.OrdAggregatorClientConfig{
				ClientTimeout:             time.Second * 30,
				OrdAggregatorAggregateAPI: URL + test.Endpoint,
				SkipSSLValidation:         true,
			}
			client := apiclient.NewORDClient(ORDClientConfig)
			client.SetHTTPClient(mockClient)

			// WHEN
			err := client.Aggregate(ctx, appID, appTemplateID)

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

func fixOrdClient() (*http.Client, func(), string) {
	mux := http.NewServeMux()

	mux.HandleFunc(aggregateEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}

func fixErrorNotFoundOrdClient() (*http.Client, func(), string) {
	mux := http.NewServeMux()
	mux.HandleFunc(aggregateEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}
