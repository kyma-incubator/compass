package systemfetcher_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"

	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/stretchr/testify/require"
)

var fourSystemsResp = `[{
			"displayName": "name1",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		}, 
		{
			"displayName": "name2",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		},
		{
			"displayName": "name3",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		}, {
			"displayName": "name4",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		}]`

func TestFetchSystemsForTenant(t *testing.T) {
	systemsJSON, err := json.Marshal(fixSystems())
	require.NoError(t, err)

	mock, url := fixHTTPClient(t)
	mock.bodiesToReturn = [][]byte{systemsJSON}
	mock.expectedFilterCriteria = "filter1"
	mock.expectedTenantFilterCriteria = "ffff and tenant eq 'tenant1'"

	client := systemfetcher.NewClient(systemfetcher.APIConfig{
		Endpoint:                    url + "/fetch",
		FilterCriteria:              "filter1",
		FilterTenantCriteriaPattern: "ffff and tenant eq '%s'",
		PageSize:                    4,
		PagingSkipParam:             "$skip",
		PagingSizeParam:             "$top",
	}, mock.httpClient)

	t.Run("Success", func(t *testing.T) {
		mock.callNumber = 0
		mock.pageCount = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.NoError(t, err)
		require.Len(t, systems, 2)
		require.Equal(t, systems[0].TemplateID, "")
	})

	t.Run("Success with template mappings", func(t *testing.T) {
		systemfetcher.Mappings = []systemfetcher.TemplateMapping{
			{
				ID:          "type1",
				SourceKey:   []string{"templateProp"},
				SourceValue: []string{"type1"},
			},
		}
		mock.bodiesToReturn = [][]byte{[]byte(`[{
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
		}]`)}
		mock.callNumber = 0
		mock.pageCount = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.NoError(t, err)
		require.Len(t, systems, 4)
		require.Equal(t, systems[0].TemplateID, "type1")
		require.Equal(t, systems[1].TemplateID, "")
	})

	t.Run("Success for more than one page", func(t *testing.T) {
		systemfetcher.Mappings = []systemfetcher.TemplateMapping{
			{
				ID:          "type1",
				SourceKey:   []string{"templateProp"},
				SourceValue: []string{"type1"},
			},
		}
		mock.bodiesToReturn = [][]byte{
			[]byte(fourSystemsResp),
			[]byte(`[{
			"displayName": "name5",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		}]`)}
		mock.callNumber = 0
		mock.pageCount = 2
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.NoError(t, err)
		require.Len(t, systems, 10)
	})

	t.Run("Does not map to the last template mapping if haven't matched before", func(t *testing.T) {
		systemfetcher.Mappings = []systemfetcher.TemplateMapping{
			{
				Name:        "type1",
				ID:          "type1",
				SourceKey:   []string{"templateProp"},
				SourceValue: []string{"type1"},
			},
			{
				Name:        "type2",
				ID:          "type2",
				SourceKey:   []string{"templateProp"},
				SourceValue: []string{"type2"},
			},
			{
				Name:        "type3",
				ID:          "type3",
				SourceKey:   []string{"templateProp"},
				SourceValue: []string{"type3"},
			},
		}
		mock.bodiesToReturn = [][]byte{[]byte(`[{
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
		}, {
			"displayName": "name3",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type4"
		}]`)}
		mock.callNumber = 0
		mock.pageCount = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.NoError(t, err)
		require.Len(t, systems, 6)
		require.Equal(t, systems[0].TemplateID, "type1")
		require.Equal(t, systems[1].TemplateID, "type2")
		require.Equal(t, systems[2].TemplateID, "")
	})

	t.Run("Fail with unexpected status code", func(t *testing.T) {
		mock.callNumber = 0
		mock.pageCount = 1
		mock.statusCodeToReturn = http.StatusBadRequest
		_, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.Contains(t, err.Error(), "unexpected status code")
	})

	t.Run("Fail because response body is not JSON", func(t *testing.T) {
		mock.callNumber = 0
		mock.pageCount = 1
		mock.bodiesToReturn = [][]byte{[]byte("not a JSON")}
		mock.statusCodeToReturn = http.StatusOK
		_, err := client.FetchSystemsForTenant(context.Background(), "tenant1")
		require.Contains(t, err.Error(), "failed to unmarshal systems response")
	})
}

type mockData struct {
	expectedFilterCriteria       string
	expectedTenantFilterCriteria string
	statusCodeToReturn           int
	bodiesToReturn               [][]byte
	httpClient                   systemfetcher.ApiClient
	callNumber                   int
	pageCount                    int
}

func fixHTTPClient(t *testing.T) (*mockData, string) {
	mux := http.NewServeMux()
	requests := []string{}

	mock := mockData{
		callNumber: 1,
	}
	mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("$filter")
		if mock.callNumber < mock.pageCount {
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
			index := mock.callNumber % mock.pageCount //this way each of the body to return mocks will be returned once for both filter criteria
			_, err := w.Write(mock.bodiesToReturn[index])
			require.NoError(t, err)
		} else {
			_, err := w.Write([]byte{})
			require.NoError(t, err)
		}
		mock.callNumber++
	})

	ts := httptest.NewServer(mux)
	mock.httpClient = systemfetcher.NewOauthClient(oauth.Config{}, ts.Client())

	return &mock, ts.URL
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
