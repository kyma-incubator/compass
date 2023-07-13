package systemfetcher_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"

	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/stretchr/testify/require"
)

var fourSystemsResp = `[{
			"displayName": "name1",
			"productDescription": "description",
			"productId": "FSM",
			"ppmsProductVersionId": "123456",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		}, 
		{
			"displayName": "name2",
			"productDescription": "description",
			"productId": "FSM",
			"ppmsProductVersionId": "123456",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		},
		{
			"displayName": "name3",
			"productDescription": "description",
			"productId": "FSM",
			"ppmsProductVersionId": "123456",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		}, {
			"displayName": "name4",
			"productDescription": "description",
			"productId": "FSM",
			"ppmsProductVersionId": "123456",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"templateProp": "type1"
		}]`

func TestFetchSystemsForTenant(t *testing.T) {
	systemsJSON, err := json.Marshal(fixSystems())
	require.NoError(t, err)

	mock, url := fixHTTPClient(t)
	mock.bodiesToReturn = [][]byte{systemsJSON}
	mock.expectedFilterCriteria = ""

	sourceKey := "key"
	labelFilter := "templateProp"

	tenantID := "tenantId1"
	syncTimestampID := "timestampId1"

	systemfetcher.SystemSourceKey = sourceKey
	systemfetcher.ApplicationTemplateLabelFilter = labelFilter

	var mutex sync.Mutex
	client := systemfetcher.NewClient(systemfetcher.APIConfig{
		Endpoint:        url + "/fetch",
		FilterCriteria:  "%s",
		PageSize:        4,
		PagingSkipParam: "$skip",
		PagingSizeParam: "$top",
		SystemSourceKey: sourceKey,
		SystemRPSLimit:  15,
	}, mock.httpClient)

	t.Run("Success", func(t *testing.T) {
		mock.callNumber = 0
		mock.pageCount = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1", &mutex)
		require.NoError(t, err)
		require.Len(t, systems, 5)
		require.Equal(t, systems[0].TemplateID, "")
	})

	t.Run("Success with template mappings", func(t *testing.T) {
		mock.expectedFilterCriteria = "(key eq 'type1')"

		systemfetcher.ApplicationTemplates = []systemfetcher.TemplateMapping{
			{
				AppTemplate: &model.ApplicationTemplate{
					ID: "type1",
				},
				Labels: map[string]*model.Label{
					labelFilter: {
						Key:   labelFilter,
						Value: "type1",
					},
				},
			},
		}
		mock.bodiesToReturn = [][]byte{[]byte(`[{
			"displayName": "name1",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"key": "type1"
		}, {
			"displayName": "name2",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"key": "type2"
		}]`)}
		mock.callNumber = 0
		mock.pageCount = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1", &mutex)
		require.NoError(t, err)
		require.Len(t, systems, 2)
		require.Equal(t, systems[0].TemplateID, "type1")
		require.Equal(t, systems[1].TemplateID, "")
	})

	t.Run("Success with template mappings and SystemSynchronizationTimestamps exist", func(t *testing.T) {
		mock.expectedFilterCriteria = "(key eq 'type1' and lastChangeDateTime gt '2023-05-02 20:30:00 +0000 UTC')"

		systemfetcher.ApplicationTemplates = []systemfetcher.TemplateMapping{
			{
				AppTemplate: &model.ApplicationTemplate{
					ID: "type1",
				},
				Labels: map[string]*model.Label{
					labelFilter: {
						Key:   labelFilter,
						Value: "type1",
					},
				},
			},
		}

		systemfetcher.SystemSynchronizationTimestamps = map[string]map[string]systemfetcher.SystemSynchronizationTimestamp{
			tenantID: {
				"type1": {
					ID:                syncTimestampID,
					LastSyncTimestamp: time.Date(2023, 5, 2, 20, 30, 0, 0, time.UTC).UTC(),
				},
			},
		}

		mock.bodiesToReturn = [][]byte{[]byte(`[{
			"displayName": "name1",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"key": "type1"
		}, {
			"displayName": "name2",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"key": "type2"
		}]`)}
		mock.callNumber = 0
		mock.pageCount = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1", &mutex)
		require.NoError(t, err)
		require.Len(t, systems, 2)
		require.Equal(t, systems[0].TemplateID, "type1")
		require.Equal(t, systems[1].TemplateID, "")

		systemfetcher.SystemSynchronizationTimestamps = nil
	})

	t.Run("Success for more than one page", func(t *testing.T) {
		mock.expectedFilterCriteria = "(key eq 'type1')"

		systemfetcher.ApplicationTemplates = []systemfetcher.TemplateMapping{
			{
				AppTemplate: &model.ApplicationTemplate{
					ID: "type1",
				},
				Labels: map[string]*model.Label{
					labelFilter: {
						Key:   labelFilter,
						Value: "type1",
					},
				},
			},
		}

		mock.bodiesToReturn = [][]byte{
			[]byte(fourSystemsResp),
			[]byte(`[{
			"displayName": "name5",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"key": "type1"
		}]`)}
		mock.callNumber = 0
		mock.pageCount = 2
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1", &mutex)
		require.NoError(t, err)
		require.Len(t, systems, 5)
	})

	t.Run("Does not map to the last template mapping if haven't matched before", func(t *testing.T) {
		mock.expectedFilterCriteria = "(key eq 'type1') or (key eq 'type2') or (key eq 'type3')"
		systemfetcher.ApplicationTemplates = []systemfetcher.TemplateMapping{
			{
				AppTemplate: &model.ApplicationTemplate{
					ID: "type1",
				},
				Labels: map[string]*model.Label{
					labelFilter: {
						Key:   labelFilter,
						Value: "type1",
					},
				},
			},
			{
				AppTemplate: &model.ApplicationTemplate{
					ID: "type2",
				},
				Labels: map[string]*model.Label{
					labelFilter: {
						Key:   labelFilter,
						Value: "type2",
					},
				},
			},
			{
				AppTemplate: &model.ApplicationTemplate{
					ID: "type3",
				},
				Labels: map[string]*model.Label{
					labelFilter: {
						Key:   labelFilter,
						Value: "type3",
					},
				},
			},
		}

		mock.bodiesToReturn = [][]byte{[]byte(`[{
			"displayName": "name1",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"key": "type1"
		}, {
			"displayName": "name2",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"key": "type2"
		}, {
			"displayName": "name3",
			"productDescription": "description",
			"baseUrl": "url",
			"infrastructureProvider": "provider1",
			"key": "type4"
		}]`)}
		mock.callNumber = 0
		mock.pageCount = 1
		systems, err := client.FetchSystemsForTenant(context.Background(), "tenant1", &mutex)
		require.NoError(t, err)
		require.Len(t, systems, 3)
		require.Equal(t, systems[0].TemplateID, "type1")
		require.Equal(t, systems[1].TemplateID, "type2")
		require.Equal(t, systems[2].TemplateID, "")
	})

	t.Run("Fail with unexpected status code", func(t *testing.T) {
		mock.callNumber = 0
		mock.pageCount = 1
		mock.statusCodeToReturn = http.StatusBadRequest
		_, err := client.FetchSystemsForTenant(context.Background(), "tenant1", &mutex)
		require.Contains(t, err.Error(), "unexpected status code")
	})

	t.Run("Fail because response body is not JSON", func(t *testing.T) {
		mock.callNumber = 0
		mock.pageCount = 1
		mock.bodiesToReturn = [][]byte{[]byte("not a JSON")}
		mock.statusCodeToReturn = http.StatusOK
		_, err := client.FetchSystemsForTenant(context.Background(), "tenant1", &mutex)
		require.Contains(t, err.Error(), "failed to unmarshal systems response")
	})
}

type mockData struct {
	expectedFilterCriteria string
	statusCodeToReturn     int
	bodiesToReturn         [][]byte
	httpClient             systemfetcher.APIClient
	callNumber             int
	pageCount              int
}

func fixHTTPClient(t *testing.T) (*mockData, string) {
	mux := http.NewServeMux()
	requests := []string{}

	mock := mockData{
		callNumber: 1,
	}
	mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("$filter")
		require.Equal(t, mock.expectedFilterCriteria, filter)

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
			SystemPayload: map[string]interface{}{
				"displayName":            "System1",
				"productDescription":     "System1 description",
				"baseUrl":                "http://example1.com",
				"infrastructureProvider": "test",
				"additionalUrls":         map[string]string{"mainUrl": "http://mainurl.com"},
			},
			StatusCondition: model.ApplicationStatusConditionInitial,
		},
	}
}

func fixSystemsWithTbt() []systemfetcher.System {
	return []systemfetcher.System{
		{
			SystemPayload: map[string]interface{}{
				"displayName":             "System2",
				"productDescription":      "System2 description",
				"baseUrl":                 "http://example2.com",
				"infrastructureProvider":  "test",
				"additionalUrls":          map[string]string{"mainUrl": "http://mainurl.com"},
				"businessTypeId":          "Test business type id",
				"businessTypeDescription": "Test business description",
			},
			StatusCondition: model.ApplicationStatusConditionInitial,
		},
	}
}
