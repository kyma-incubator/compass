package tenantfetcher_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher/automock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_FetchTenantEventsPage(t *testing.T) {
	// GIVEN
	mockClient, mockServerCloseFn, endpoint := fixHTTPClient(t)
	defer mockServerCloseFn()

	metricsPusherMock := fixMetricsPusherMock()
	defer metricsPusherMock.AssertExpectations(t)

	queryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": "1",
	}

	subaccountQueryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": "1",
		"region":    "test-region",
	}

	apiCfg := tenantfetcher.APIConfig{
		EndpointTenantCreated:     endpoint + "/ga-created",
		EndpointTenantDeleted:     endpoint + "/ga-deleted",
		EndpointTenantUpdated:     endpoint + "/ga-updated",
		EndpointSubaccountCreated: endpoint + "/sub-created",
		EndpointSubaccountDeleted: endpoint + "/sub-deleted",
		EndpointSubaccountUpdated: endpoint + "/sub-updated",
		EndpointSubaccountMoved:   endpoint + "/sub-moved",
	}
	client, err := tenantfetcher.NewClient(tenantfetcher.OAuth2Config{}, oauth.Standard, apiCfg, time.Second)
	require.NoError(t, err)

	client.SetMetricsPusher(metricsPusherMock)
	client.SetHTTPClient(mockClient)

	t.Run("Success fetching account creation events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.CreatedAccountType, queryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching account update events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.UpdatedAccountType, queryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching account deletion events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.DeletedAccountType, queryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching subaccount creation events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.CreatedSubaccountType, subaccountQueryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching subaccount update events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.UpdatedSubaccountType, subaccountQueryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching subaccount deletion events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.DeletedSubaccountType, subaccountQueryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching moved subaccount events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.MovedSubaccountType, subaccountQueryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Error when unknown events type", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(-1, queryParams)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unknown events type").Error())
		assert.Empty(t, res)
	})

	// GIVEN
	apiCfg = tenantfetcher.APIConfig{
		EndpointTenantCreated: "___ :// ___ ",
		EndpointTenantDeleted: "http://127.0.0.1:8111/badpath",
		EndpointTenantUpdated: endpoint + "/empty",
	}
	client, err = tenantfetcher.NewClient(tenantfetcher.OAuth2Config{}, oauth.Standard, apiCfg, time.Second)
	require.NoError(t, err)

	client.SetMetricsPusher(metricsPusherMock)
	client.SetHTTPClient(mockClient)

	t.Run("Success when no content", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.UpdatedAccountType, queryParams)
		// THEN
		require.NoError(t, err)
		require.Empty(t, res)
	})

	t.Run("Error when endpoint not parsable", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.CreatedAccountType, queryParams)
		// THEN
		require.EqualError(t, err, "parse \"___ :// ___ \": first path segment in URL cannot contain colon")
		assert.Empty(t, res)
	})

	t.Run("Error when bad path", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.DeletedAccountType, queryParams)
		// THEN
		require.EqualError(t, err, "while sending get request: Get \"http://127.0.0.1:8111/badpath?pageNum=1&pageSize=1&timestamp=1\": dial tcp 127.0.0.1:8111: connect: connection refused")
		assert.Empty(t, res)
	})

	// GIVEN
	apiCfg = tenantfetcher.APIConfig{
		EndpointTenantCreated: endpoint + "/created",
		EndpointTenantDeleted: endpoint + "/deleted",
		EndpointTenantUpdated: endpoint + "/badRequest",
	}
	client, err = tenantfetcher.NewClient(tenantfetcher.OAuth2Config{}, oauth.Standard, apiCfg, time.Second)
	require.NoError(t, err)

	client.SetMetricsPusher(metricsPusherMock)
	client.SetHTTPClient(mockClient)

	t.Run("Error when status code not equal to 200 OK and 204 No Content is returned", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.UpdatedAccountType, queryParams)
		// THEN
		require.EqualError(t, err, fmt.Sprintf("request to \"%s/badRequest?pageNum=1&pageSize=1&timestamp=1\" returned status code 400 and body \"\"", endpoint))
		assert.Empty(t, res)
	})

	// GIVEN
	apiCfg = tenantfetcher.APIConfig{
		EndpointSubaccountMoved: "",
	}
	client, err = tenantfetcher.NewClient(tenantfetcher.OAuth2Config{}, oauth.Standard, apiCfg, time.Second)
	require.NoError(t, err)

	client.SetMetricsPusher(metricsPusherMock)
	client.SetHTTPClient(mockClient)

	t.Run("Skip fetching moved subaccount events when endpoint is not provided", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(tenantfetcher.MovedSubaccountType, queryParams)
		// THEN
		require.NoError(t, err)
		require.Nil(t, res)
	})
}

func fixHTTPClient(t *testing.T) (*http.Client, func(), string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/ga-created", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixCreatedTenantsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/ga-deleted", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixDeletedAccountsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/ga-updated", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixUpdatedAccountsJSON())
		require.NoError(t, err)
	})

	mux.HandleFunc("/sub-created", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixCreatedSubaccountsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/sub-deleted", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixDeletedSubaccountsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/sub-updated", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixUpdatedSubaccountsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/sub-moved", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixMovedSubaccountsJSON())
		require.NoError(t, err)
	})

	mux.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/badRequest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}

func fixCreatedTenantsJSON() string {
	return `{
 "events": [
   {
     "id": 5,
     "type": "GLOBALACCOUNT_CREATION",
     "timestamp": "1579771215736",
     "eventData": "{\"id\":\"55\",\"displayName\":\"TEN5\",\"model\":\"default\"}"
   },
   {
     "id": 4,
     "type": "GLOBALACCOUNT_CREATION",
     "timestamp": "1579771215636",
     "eventData": "{\"id\":\"44\",\"displayName\":\"TEN4\",\"model\":\"default\"}"
   },
	{
     "id": 3,
     "type": "GLOBALACCOUNT_CREATION",
     "timestamp": "1579771215536",
     "eventData": "{\"id\":\"33\",\"displayName\":\"TEN3\",\"model\":\"default\"}"
   },
	{
     "id": 2,
     "type": "GLOBALACCOUNT_CREATION",
     "timestamp": "1579771215436",
     "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
   },
	{
     "id": 1,
     "type": "GLOBALACCOUNT_CREATION",
     "timestamp": "1579771215336",
     "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
   }
 ],
 "totalResults": 5,
 "totalPages": 1
}`
}

func fixUpdatedAccountsJSON() string {
	return `{
 "events": [
	{
     "id": 2,
     "type": "GLOBALACCOUNT_UPDATE",
     "timestamp": "1579771215436",
     "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
   },
	{
     "id": 1,
     "type": "GLOBALACCOUNT_UPDATE",
     "timestamp": "1579771215336",
     "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
   }
 ],
 "totalResults": 2,
 "totalPages": 1
}`
}

func fixDeletedAccountsJSON() string {
	return `{
 "events": [
	{
     "id": 2,
     "type": "GLOBALACCOUNT_DELETION",
     "timestamp": "1579771215436",
     "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
   },
	{
     "id": 1,
     "type": "GLOBALACCOUNT_DELETION",
     "timestamp": "1579771215336",
     "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
   }
 ],
 "totalResults": 2,
 "totalPages": 1
}`
}

func fixCreatedSubaccountsJSON() string {
	return `{
 "events": [
   {
     "id": 5,
     "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
     "timestamp": "1579771215736",
     "eventData": "{\"id\":\"55\",\"displayName\":\"TEN5\",\"model\":\"default\"}"
   },
   {
     "id": 4,
     "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
     "timestamp": "1579771215636",
     "eventData": "{\"id\":\"44\",\"displayName\":\"TEN4\",\"model\":\"default\"}"
   },
	{
     "id": 3,
     "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
     "timestamp": "1579771215536",
     "eventData": "{\"id\":\"33\",\"displayName\":\"TEN3\",\"model\":\"default\"}"
   },
	{
     "id": 2,
     "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
     "timestamp": "1579771215436",
     "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
   },
	{
     "id": 1,
     "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
     "timestamp": "1579771215336",
     "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
   }
 ],
 "totalResults": 5,
 "totalPages": 1
}`
}

func fixUpdatedSubaccountsJSON() string {
	return `{
 "events": [
	{
     "id": 2,
     "type": "SUBACCOUNT_UPDATE",
	 "region": "test-region",
     "timestamp": "1579771215436",
     "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
   },
	{
     "id": 1,
     "type": "SUBACCOUNT_UPDATE",
	 "region": "test-region",
     "timestamp": "1579771215336",
     "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
   }
 ],
 "totalResults": 2,
 "totalPages": 1
}`
}

func fixDeletedSubaccountsJSON() string {
	return `{
 "events": [
	{
     "id": 2,
     "type": "SUBACCOUNT_DELETION",
	 "region": "test-region",
     "timestamp": "1579771215436",
     "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
   },
	{
     "id": 1,
     "type": "SUBACCOUNT_DELETION",
	 "region": "test-region",
     "timestamp": "1579771215336",
     "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
   }
 ],
 "totalResults": 2,
 "totalPages": 1
}`
}

func fixMovedSubaccountsJSON() string {
	return `{
 "events": [
	{
     "id": 2,
     "type": "SUBACCOUNT_MOVED",
	 "region": "test-region",
     "timestamp": "1579771215436",
     "eventData": "{\"id\":\"22\",\"source\":\"TEN1\",\"target\":\"TEN2\"}"
   },
	{
     "id": 1,
     "type": "SUBACCOUNT_MOVED",
	 "region": "test-region",
     "timestamp": "1579771215336",
     "eventData": "{\"id\":\"11\",\"source\":\"TEN3\",\"target\":\"TEN4\"}"
   }
 ],
 "totalResults": 2,
 "totalPages": 1
}`
}

func fixMetricsPusherMock() *automock.MetricsPusher {
	metricsPusherMock := &automock.MetricsPusher{}
	metricsPusherMock.On("RecordEventingRequest", http.MethodGet, http.StatusOK, "200 OK")
	metricsPusherMock.On("RecordEventingRequest", http.MethodGet, http.StatusNoContent, "204 No Content").Once()
	metricsPusherMock.On("RecordEventingRequest", http.MethodGet, 0, "connect: connection refused").Once()
	metricsPusherMock.On("RecordEventingRequest", http.MethodGet, http.StatusBadRequest, "400 Bad Request").Once()

	return metricsPusherMock
}
