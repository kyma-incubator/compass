package tenantfetcher_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_FetchTenantEventsPage(t *testing.T) {
	// GIVEN
	mockClient, mockServerCloseFn, endpoint := fixHTTPClient(t)
	defer mockServerCloseFn()
	apiCfg := tenantfetcher.APIConfig{
		EndpointTenantCreated: endpoint + "/created",
		EndpointTenantDeleted: endpoint + "/deleted",
		EndpointTenantUpdated: endpoint + "/updated",
	}
	client := tenantfetcher.NewClient(tenantfetcher.OAuth2Config{}, apiCfg)
	client.SetHTTPClient(mockClient)

	t.Run("Success fetching creation events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.CreatedEventsType, 1)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching update events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.UpdatedEventsType, 1)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching deletion events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.DeletedEventsType, 1)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Error when unkown events type", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(-1, 1)
		// THEN
		require.EqualError(t, err, "unknown events type")
		assert.Empty(t, res)
	})

	// GIVEN
	apiCfg = tenantfetcher.APIConfig{
		EndpointTenantCreated: "___ :// ___ ",
		EndpointTenantDeleted: "http://127.0.0.1:8111/badpath",
		EndpointTenantUpdated: endpoint + "/empty",
	}
	client = tenantfetcher.NewClient(tenantfetcher.OAuth2Config{}, apiCfg)
	client.SetHTTPClient(mockClient)

	t.Run("Success when no content", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.UpdatedEventsType, 1)
		// THEN
		require.NoError(t, err)
		require.Empty(t, res)
	})

	t.Run("Error when endpoint not parsable", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.CreatedEventsType, 1)
		// THEN
		require.EqualError(t, err, "parse ___ :// ___ : first path segment in URL cannot contain colon")
		assert.Empty(t, res)
	})

	t.Run("Error when bad path", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.DeletedEventsType, 1)
		// THEN
		require.EqualError(t, err, "while sending get request: Get http://127.0.0.1:8111/badpath?page=1&resultsPerPage=1000&ts=1: dial tcp 127.0.0.1:8111: connect: connection refused")
		assert.Empty(t, res)
	})
}

func fixHTTPClient(t *testing.T) (*http.Client, func(), string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/created", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixCreatedTenantsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/deleted", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixDeletedTenantsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/updated", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixUpdatedTenantsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}

func fixCreatedTenantsJSON() string {
	return `{
  "events": [
    {
      "id": 5,
      "type": "CREATION",
      "timestamp": "1579771215736",
      "eventData": "{\"id\":\"55\",\"displayName\":\"TEN5\",\"model\":\"default\"}"
    },
    {
      "id": 4,
      "type": "CREATION",
      "timestamp": "1579771215636",
      "eventData": "{\"id\":\"44\",\"displayName\":\"TEN4\",\"model\":\"default\"}"
    },
	{
      "id": 3,
      "type": "CREATION",
      "timestamp": "1579771215536",
      "eventData": "{\"id\":\"33\",\"displayName\":\"TEN3\",\"model\":\"default\"}"
    },
	{
      "id": 2,
      "type": "CREATION",
      "timestamp": "1579771215436",
      "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
    },
	{
      "id": 1,
      "type": "CREATION",
      "timestamp": "1579771215336",
      "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
    }
  ],
  "totalResults": 5,
  "totalPages": 1
}`
}

func fixUpdatedTenantsJSON() string {
	return `{
  "events": [
	{
      "id": 2,
      "type": "UPDATE",
      "timestamp": "1579771215436",
      "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
    },
	{
      "id": 1,
      "type": "UPDATE",
      "timestamp": "1579771215336",
      "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
    }
  ],
  "totalResults": 2,
  "totalPages": 1
}`
}

func fixDeletedTenantsJSON() string {
	return `{
  "events": [
	{
      "id": 2,
      "type": "DELETION",
      "timestamp": "1579771215436",
      "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
    },
	{
      "id": 1,
      "type": "DELETION",
      "timestamp": "1579771215336",
      "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
    }
  ],
  "totalResults": 2,
  "totalPages": 1
}`
}
