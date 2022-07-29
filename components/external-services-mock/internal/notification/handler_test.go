package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var url = "https://target-url.com"
var reqBody = `{"ucl-formation-id":"96bbd806-0d56-4f39-bcf1-e15aee9e50cc","items":[{"region":"testRegion","application-namespace":"appNamespce","tenant-id":"localTenantID","ucl-system-tenant-id":"9d8bb1a5-4799-453c-a406-84439f151d45"}]}`

func TestHandler_Patch(t *testing.T) {
	tenantId := "tenantId"
	apiPath := fmt.Sprintf("/formation-callback/%s", tenantId)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		ExpectedResponseCode int
		ExpectedMappings     map[string][]Response
	}{
		{
			Name:                 "success",
			RequestBody:          reqBody,
			TenantID:             tenantId,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings: map[string][]Response{
				"tenantId": {
					{
						Operation:     "assign",
						ApplicationID: nil,
						RequestBody:   json.RawMessage(reqBody),
					},
				},
			},
		},
		{
			Name:                 "Error tenant id not found in path",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]Response{},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             tenantId,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings: map[string][]Response{
				"tenantId": {},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(http.MethodPatch, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			if testCase.TenantID != "" {
				req = mux.SetURLVars(req, map[string]string{"tenantId": testCase.TenantID})
			}

			h := NewHandler()
			r := httptest.NewRecorder()

			//WHEN
			h.Patch(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.mappings)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	tenantId := "tenantId"
	appID := "appID"
	apiPath := fmt.Sprintf("/formation-callback/%s/%s", tenantId, appID)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		AppID                string
		ExpectedResponseCode int
		ExpectedMappings     map[string][]Response
	}{
		{
			Name:                 "success",
			RequestBody:          reqBody,
			TenantID:             tenantId,
			AppID:                appID,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings: map[string][]Response{
				"tenantId": {
					{
						Operation:     "unassign",
						ApplicationID: &appID,
						RequestBody:   json.RawMessage(reqBody),
					},
				},
			},
		},
		{
			Name:                 "Error tenant id not found in path",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]Response{},
		},
		{
			Name:                 "Error appID not found in path",
			TenantID:             tenantId,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]Response{},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             tenantId,
			AppID:                appID,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings: map[string][]Response{
				"tenantId": {},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(http.MethodDelete, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			vars := map[string]string{}
			if testCase.TenantID != "" {
				vars["tenantId"] = testCase.TenantID
			}
			if testCase.AppID != "" {
				vars["applicationId"] = testCase.AppID
			}
			req = mux.SetURLVars(req, vars)

			h := NewHandler()
			r := httptest.NewRecorder()

			//WHEN
			h.Delete(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.mappings)
		})
	}
}

func TestGetResponses(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	h := NewHandler()
	h.mappings = map[string][]Response{
		"tenantId": {
			{
				Operation:     "assign",
				ApplicationID: nil,
				RequestBody:   json.RawMessage(reqBody),
			},
		},
	}
	r := httptest.NewRecorder()

	//WHEN
	h.GetResponses(r, req)
	resp := r.Result()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	mappings := map[string][]Response{}
	require.NoError(t, json.Unmarshal(body, &mappings))

	//THEN
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.Equal(t, h.mappings, mappings)
}

func TestCleanup(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	h := NewHandler()
	h.mappings = map[string][]Response{
		"tenantId": {
			{
				Operation:     "assign",
				ApplicationID: nil,
				RequestBody:   json.RawMessage(reqBody),
			},
		},
	}
	r := httptest.NewRecorder()

	//WHEN
	h.Cleanup(r, req)
	resp := r.Result()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	//THEN
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.Equal(t, map[string][]Response{}, h.mappings)
}
