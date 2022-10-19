package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

var url = "https://target-url.com"
var reqBody = `{"ucl-formation-id":"96bbd806-0d56-4f39-bcf1-e15aee9e50cc","items":[{"region":"testRegion","application-namespace":"appNamespce","tenant-id":"localTenantID","ucl-system-tenant-id":"9d8bb1a5-4799-453c-a406-84439f151d45"}]}`
var reqConfigBody = `{"ucl-formation-id":"96bbd806-0d56-4f39-bcf1-e15aee9e50cc","configuration": "{\"key\":\"value\"}","items":[{"region":"testRegion","application-namespace":"appNamespce","tenant-id":"localTenantID","ucl-system-tenant-id":"9d8bb1a5-4799-453c-a406-84439f151d45"}]}`

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

func TestHandler_RespondWithIncomplete(t *testing.T) {
	tenantId := "tenantId"
	apiPath := fmt.Sprintf("/formation-callback/configuration/%s", tenantId)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		ExpectedResponseCode int
		ExpectedMappings     map[string][]Response
	}{
		{
			Name:                 "success with no config",
			RequestBody:          reqBody,
			TenantID:             tenantId,
			ExpectedResponseCode: http.StatusNoContent,
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
			Name:                 "success with config",
			RequestBody:          reqConfigBody,
			TenantID:             tenantId,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings: map[string][]Response{
				"tenantId": {
					{
						Operation:     "assign",
						ApplicationID: nil,
						RequestBody:   json.RawMessage(reqConfigBody),
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
			h.RespondWithIncomplete(r, req)
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

func TestHandler_FailOnceResponse(t *testing.T) {
	tenantId := "tenantId"
	appID := "appID"
	apiPath := fmt.Sprintf("/formation-callback/fail-once/%s", tenantId)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		AppID                *string
		Method               string
		ShouldFail           bool
		ExpectedResponseCode int
		ExpectedMappings     map[string][]Response
	}{
		{
			Name:                 "assign should fail once",
			RequestBody:          reqBody,
			TenantID:             tenantId,
			ShouldFail:           true,
			Method:               http.MethodPatch,
			ExpectedResponseCode: http.StatusBadRequest,
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
			Name:                 "assign should succeed",
			RequestBody:          reqBody,
			TenantID:             tenantId,
			ShouldFail:           false,
			Method:               http.MethodPatch,
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
			Name:                 "unassign should fail once",
			RequestBody:          reqBody,
			TenantID:             tenantId,
			AppID:                &appID,
			Method:               http.MethodDelete,
			ShouldFail:           true,
			ExpectedResponseCode: http.StatusBadRequest,
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
			Name:                 "unassign should succeed",
			RequestBody:          reqBody,
			TenantID:             tenantId,
			AppID:                &appID,
			Method:               http.MethodDelete,
			ShouldFail:           false,
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
			Method:               http.MethodPatch,
			ExpectedMappings:     map[string][]Response{},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             tenantId,
			Method:               http.MethodPatch,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings: map[string][]Response{
				"tenantId": {},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fullApiPath := apiPath
			if testCase.AppID != nil {
				fullApiPath += fmt.Sprintf("/%s", *testCase.AppID)
			}

			req, err := http.NewRequest(testCase.Method, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			if testCase.AppID != nil {
				req = mux.SetURLVars(req, map[string]string{"tenantId": testCase.TenantID, "applicationId": *testCase.AppID})
			} else if testCase.TenantID != "" {
				req = mux.SetURLVars(req, map[string]string{"tenantId": testCase.TenantID})
			}

			h := NewHandler()
			r := httptest.NewRecorder()

			h.shouldReturnError = testCase.ShouldFail

			//WHEN
			h.FailOnceResponse(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.mappings)
			require.False(t, h.shouldReturnError)
		})
	}
}

func TestResetShouldFail(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	h := NewHandler()
	h.shouldReturnError = false
	r := httptest.NewRecorder()

	//WHEN
	h.ResetShouldFail(r, req)
	resp := r.Result()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	//THEN
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.True(t, h.shouldReturnError)
}
