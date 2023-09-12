package apiclient_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/apiclient"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	externalID    = "externalID"
	internalID    = "internalID"
	tenantHeader  = "Tenant"
	appID         = "appID"
	appTemplateID = "appTemplateID"
	fakeEndpoint  = "http://fake-endpoint"
)

func TestORDClient_Aggregate(t *testing.T) {
	// GIVEN
	mockClient, mockServerCloseFn, endpoint := fixOrdClient()
	defer mockServerCloseFn()

	ORDClientConfig := apiclient.OrdAggregatorClientConfig{
		ClientTimeout:             time.Second * 30,
		OrdAggregatorAggregateAPI: endpoint,
		SkipSSLValidation:         true,
	}

	client := apiclient.NewORDClient(ORDClientConfig)

	client.SetHTTPClient(mockClient)

	t.Run("Success calling aggregate api", func(t *testing.T) {
		//GIVEN
		ctx := tenant.SaveToContext(context.TODO(), internalID, externalID)
		// WHEN
		err := client.Aggregate(ctx, appID, appTemplateID)
		// THEN
		require.NoError(t, err)
	})
	t.Run("Error when tenant is missing from context", func(t *testing.T) {
		// WHEN
		err := client.Aggregate(context.TODO(), appID, appTemplateID)
		// THEN
		require.Error(t, err)
	})
	t.Run("Error when calling aggregate api", func(t *testing.T) {
		//GIVEN
		ctx := tenant.SaveToContext(context.TODO(), internalID, externalID)

		ORDClientConfig = apiclient.OrdAggregatorClientConfig{
			ClientTimeout:             time.Second * 30,
			OrdAggregatorAggregateAPI: fakeEndpoint,
			SkipSSLValidation:         true,
		}
		client = apiclient.NewORDClient(ORDClientConfig)

		client.SetHTTPClient(mockClient)
		// WHEN
		err := client.Aggregate(ctx, appID, appTemplateID)
		// THEN
		require.Error(t, err)
	})
}

func fixOrdClient() (*http.Client, func(), string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/aggregate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(tenantHeader, externalID)
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}
