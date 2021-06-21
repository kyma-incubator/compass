package systemfetcher_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/stretchr/testify/require"
)

func TestFetchSystemsForTenant(t *testing.T) {
	systemsJson, err := json.Marshal(fixSystems())
	require.NoError(t, err)

	mock, url := fixHttpClient(t)
	mock.bodyToReturn = systemsJson
	mock.expectedFilterCriteria = "filter1"
	mock.expectedTenantFilterCriteria = "ffff and tenant eq 'tenant1'"

	client := systemfetcher.NewClient(systemfetcher.APIConfig{
		FilterCriteria:              "filter1",
		FilterTenantCriteriaPattern: "ffff and tenant eq '%s'",
		Endpoint:                    url + "/fetch",
	}, systemfetcher.OAuth2Config{}, mockClientCreator(mock.httpClient))

	t.Run("Success", func(t *testing.T) {
		mock.callNumber = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.NoError(t, err)
		require.Len(t, systems, 2)
		require.Equal(t, systems[0].TemplateType, "")
	})

	t.Run("Success with template mappings", func(t *testing.T) {
		systemfetcher.Mappings = []systemfetcher.TempMapping{
			{
				Name:        "type1",
				SourceKey:   []string{"templateProp"},
				SourceValue: []string{"type1"},
			},
		}
		mock.bodyToReturn = []byte(`[{
			"displayName": "name1",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		}, {
			"displayName": "name2",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type2"
		}]`)
		mock.callNumber = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.NoError(t, err)
		require.Len(t, systems, 4)
		require.Equal(t, systems[0].TemplateType, "type1")
	})

	t.Run("Fail with unexpected status code", func(t *testing.T) {
		mock.callNumber = 1
		mock.statusCodeToReturn = http.StatusBadRequest
		_, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.Contains(t, err.Error(), "unexpected status code")
	})

	t.Run("Fail because response body is not JSON", func(t *testing.T) {
		mock.callNumber = 1
		mock.bodyToReturn = []byte("not a JSON")
		mock.statusCodeToReturn = http.StatusOK
		_, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.Contains(t, err.Error(), "failed to unmarshal systems response")
	})
}

type mockData struct {
	expectedFilterCriteria       string
	expectedTenantFilterCriteria string
	statusCodeToReturn           int
	bodyToReturn                 []byte
	httpClient                   *http.Client
	callNumber                   int
}

func fixHttpClient(t *testing.T) (*mockData, string) {
	mux := http.NewServeMux()
	requests := []string{}

	mock := mockData{
		callNumber: 1,
	}

	mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("$filter")
		if mock.callNumber == 1 {
			require.Equal(t, mock.expectedFilterCriteria, filter)
		} else {
			require.Equal(t, mock.expectedTenantFilterCriteria, filter)
		}

		requests = append(requests, filter)
		w.Header().Set("Content-Type", "application/json")
		if mock.statusCodeToReturn == 0 {
			mock.statusCodeToReturn = http.StatusOK
		}
		w.WriteHeader(mock.statusCodeToReturn)

		if mock.statusCodeToReturn == http.StatusOK {
			_, err := w.Write(mock.bodyToReturn)
			require.NoError(t, err)
		} else {
			_, err := w.Write([]byte{})
			require.NoError(t, err)
		}
		mock.callNumber++
	})

	ts := httptest.NewServer(mux)
	mock.httpClient = ts.Client()

	return &mock, ts.URL
}

func mockClientCreator(client *http.Client) func(ctx context.Context, oauth2Config systemfetcher.OAuth2Config, scopes []string, tenant string) *http.Client {
	return func(ctx context.Context, oauth2Config systemfetcher.OAuth2Config, scopes []string, tenant string) *http.Client {
		return client
	}
}

func fixSystems() []systemfetcher.System {
	return []systemfetcher.System{
		{
			SystemBase: systemfetcher.SystemBase{
				DisplayName:            "System1",
				ProductDescription:     "System1 description",
				BaseURL:                "http://example1.com",
				InfrastructureProvider: "test",
			},
		},
	}
}
