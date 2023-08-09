package notification_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/notification"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

var assignMappingsWithoutConfig = fixFormationAssignmentMappings(notification.Assign, testTenantID, formationAssignmentReqBody, nil)
var assignMappingsWithConfig = fixFormationAssignmentMappings(notification.Assign, testTenantID, formationAssignmentReqConfigBody, nil)
var unassignMappings = fixFormationAssignmentMappings(notification.Unassign, testTenantID, formationAssignmentReqBody, &appID)

func TestHandler_Patch(t *testing.T) {
	apiPath := fmt.Sprintf("/formation-callback/%s", testTenantID)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		ExpectedResponseCode int
		ExpectedMappings     map[string][]notification.Response
	}{
		{
			Name:                 "success",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings:     assignMappingsWithoutConfig,
		},
		{
			Name:                 "Error tenant id not found in path",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{testTenantID: {}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(http.MethodPatch, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			if testCase.TenantID != "" {
				req = mux.SetURLVars(req, map[string]string{tenantIDParam: testCase.TenantID})
			}

			h := notification.NewHandler(notification.NotificationsConfiguration{})
			r := httptest.NewRecorder()

			//WHEN
			h.Patch(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.Mappings)
		})
	}
}

func TestHandler_PatchWithState(t *testing.T) {
	apiPath := fmt.Sprintf("/formation-callback/%s", testTenantID)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		ExpectedResponseCode int
		ExpectedMappings     map[string][]notification.Response
	}{
		{
			Name:                 "success",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings:     assignMappingsWithoutConfig,
		},
		{
			Name:                 "Error tenant id not found in path",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{testTenantID: {}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(http.MethodPatch, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			if testCase.TenantID != "" {
				req = mux.SetURLVars(req, map[string]string{tenantIDParam: testCase.TenantID})
			}

			h := notification.NewHandler(notification.NotificationsConfiguration{})
			r := httptest.NewRecorder()

			//WHEN
			h.PatchWithState(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.Mappings)
		})
	}
}

func TestHandler_RespondWithIncomplete(t *testing.T) {
	apiPath := fmt.Sprintf("/formation-callback/configuration/%s", testTenantID)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		ExpectedResponseCode int
		ExpectedMappings     map[string][]notification.Response
	}{
		{
			Name:                 "success with no config",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusNoContent,
			ExpectedMappings:     assignMappingsWithoutConfig,
		},
		{
			Name:                 "success with config",
			RequestBody:          formationAssignmentReqConfigBody,
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings:     assignMappingsWithConfig,
		},
		{
			Name:                 "Error tenant id not found in path",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{testTenantID: {}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(http.MethodPatch, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			if testCase.TenantID != "" {
				req = mux.SetURLVars(req, map[string]string{tenantIDParam: testCase.TenantID})
			}

			h := notification.NewHandler(notification.NotificationsConfiguration{})
			r := httptest.NewRecorder()

			//WHEN
			h.RespondWithIncomplete(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.Mappings)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	apiPath := fmt.Sprintf("/formation-callback/%s/%s", testTenantID, appID)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		AppID                string
		ExpectedResponseCode int
		ExpectedMappings     map[string][]notification.Response
	}{
		{
			Name:                 "success",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			AppID:                appID,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings:     unassignMappings,
		},
		{
			Name:                 "Error tenant id not found in path",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{},
		},
		{
			Name:                 "Error appID not found in path",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{testTenantID: make([]notification.Response, 0, 1)},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{testTenantID: {}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(http.MethodDelete, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			vars := map[string]string{}
			if testCase.TenantID != "" {
				vars[tenantIDParam] = testCase.TenantID
			}
			if testCase.AppID != "" {
				vars["applicationId"] = testCase.AppID
			}
			req = mux.SetURLVars(req, vars)

			h := notification.NewHandler(notification.NotificationsConfiguration{})
			r := httptest.NewRecorder()

			//WHEN
			h.Delete(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.Mappings)
		})
	}
}

func TestHandler_DeleteWithState(t *testing.T) {
	apiPath := fmt.Sprintf("/formation-callback/%s/%s", testTenantID, appID)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		AppID                string
		ExpectedResponseCode int
		ExpectedMappings     map[string][]notification.Response
	}{
		{
			Name:                 "success",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			AppID:                appID,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings:     unassignMappings,
		},
		{
			Name:                 "Error tenant id not found in path",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{},
		},
		{
			Name:                 "Error appID not found in path",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{testTenantID: make([]notification.Response, 0, 1)},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             testTenantID,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{testTenantID: {}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(http.MethodDelete, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			vars := map[string]string{}
			if testCase.TenantID != "" {
				vars[tenantIDParam] = testCase.TenantID
			}
			if testCase.AppID != "" {
				vars["applicationId"] = testCase.AppID
			}
			req = mux.SetURLVars(req, vars)

			h := notification.NewHandler(notification.NotificationsConfiguration{})
			r := httptest.NewRecorder()

			//WHEN
			h.DeleteWithState(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.Mappings)
		})
	}
}

func TestGetResponses(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	h := notification.NewHandler(notification.NotificationsConfiguration{})
	h.Mappings = assignMappingsWithoutConfig
	r := httptest.NewRecorder()

	//WHEN
	h.GetResponses(r, req)
	resp := r.Result()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	mappings := map[string][]notification.Response{}
	require.NoError(t, json.Unmarshal(body, &mappings))

	//THEN
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.Equal(t, h.Mappings, mappings)
}

func TestCleanup(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	h := notification.NewHandler(notification.NotificationsConfiguration{})
	h.Mappings = assignMappingsWithoutConfig
	r := httptest.NewRecorder()

	//WHEN
	h.Cleanup(r, req)
	resp := r.Result()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	//THEN
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.Equal(t, map[string][]notification.Response{}, h.Mappings)
}

func TestHandler_FailOnceResponse(t *testing.T) {
	apiPath := fmt.Sprintf("/formation-callback/fail-once/%s", testTenantID)

	testCases := []struct {
		Name                 string
		RequestBody          string
		TenantID             string
		AppID                *string
		Method               string
		ShouldFail           bool
		ExpectedResponseCode int
		ExpectedMappings     map[string][]notification.Response
	}{
		{
			Name:                 "assign should fail once",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			ShouldFail:           true,
			Method:               http.MethodPatch,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     assignMappingsWithoutConfig,
		},
		{
			Name:                 "assign should succeed",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			ShouldFail:           false,
			Method:               http.MethodPatch,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings:     assignMappingsWithoutConfig,
		},
		{
			Name:                 "unassign should fail once",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			AppID:                &appID,
			Method:               http.MethodDelete,
			ShouldFail:           true,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     unassignMappings,
		},
		{
			Name:                 "unassign should succeed",
			RequestBody:          formationAssignmentReqBody,
			TenantID:             testTenantID,
			AppID:                &appID,
			Method:               http.MethodDelete,
			ShouldFail:           false,
			ExpectedResponseCode: http.StatusOK,
			ExpectedMappings:     unassignMappings,
		},
		{
			Name:                 "Error tenant id not found in path",
			ExpectedResponseCode: http.StatusBadRequest,
			Method:               http.MethodPatch,
			ExpectedMappings:     map[string][]notification.Response{},
		},
		{
			Name:                 "Error when body is not valid json",
			RequestBody:          "invalid json",
			TenantID:             testTenantID,
			Method:               http.MethodPatch,
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedMappings:     map[string][]notification.Response{testTenantID: {}},
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
				req = mux.SetURLVars(req, map[string]string{tenantIDParam: testCase.TenantID, "applicationId": *testCase.AppID})
			} else if testCase.TenantID != "" {
				req = mux.SetURLVars(req, map[string]string{tenantIDParam: testCase.TenantID})
			}

			h := notification.NewHandler(notification.NotificationsConfiguration{})
			r := httptest.NewRecorder()

			h.ShouldReturnError = testCase.ShouldFail

			//WHEN
			h.FailOnceResponse(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, h.Mappings)
			require.False(t, h.ShouldReturnError)
		})
	}
}

func TestResetShouldFail(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	h := notification.NewHandler(notification.NotificationsConfiguration{})
	h.ShouldReturnError = false
	r := httptest.NewRecorder()

	//WHEN
	h.ResetShouldFail(r, req)
	resp := r.Result()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	//THEN
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.True(t, h.ShouldReturnError)
}

func TestHandler_PostAndDeleteFormation(t *testing.T) {
	formationID := "testFormationID"
	apiPath := fmt.Sprintf("/v1/businessIntegration/%s", formationID)

	testCases := []struct {
		name                 string
		requestBody          string
		formationID          string
		expectedResponseCode int
		expectedMappings     map[string][]notification.Response
	}{
		{
			name:                 "Success",
			requestBody:          formationReqBody,
			formationID:          formationID,
			expectedResponseCode: http.StatusOK,
			expectedMappings:     fixFormationMappings(notification.CreateFormation, formationID, formationReqBody),
		},
		{
			name:                 "Error when required formationID path parameter is missing",
			expectedResponseCode: http.StatusBadRequest,
			expectedMappings:     map[string][]notification.Response{},
		},
		{
			name:                 "Error when formation request body is not valid json",
			requestBody:          "invalid json",
			formationID:          formationID,
			expectedResponseCode: http.StatusBadRequest,
			expectedMappings:     map[string][]notification.Response{formationID: {}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(testCase.requestBody)))
			require.NoError(t, err)

			if testCase.formationID != "" {
				req = mux.SetURLVars(req, map[string]string{formationIDParam: testCase.formationID})
			}

			handler := notification.NewHandler(notification.NotificationsConfiguration{})
			recorder := httptest.NewRecorder()

			//WHEN
			handler.PostFormation(recorder, req)
			resp := recorder.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.expectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.expectedMappings, handler.Mappings)
		})
	}

	t.Run("Success when the operation is delete formation", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, url+apiPath, bytes.NewBuffer([]byte(formationReqBody)))
		require.NoError(t, err)
		req = mux.SetURLVars(req, map[string]string{formationIDParam: formationID})

		handler := notification.NewHandler(notification.NotificationsConfiguration{})
		recorder := httptest.NewRecorder()

		//WHEN
		handler.DeleteFormation(recorder, req)
		resp := recorder.Result()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
		require.Equal(t, fixFormationMappings(notification.DeleteFormation, formationID, formationReqBody), handler.Mappings)
	})
}

func TestHandler_FailOnceFormation(t *testing.T) {
	formationID := "testFormationID"
	apiPath := fmt.Sprintf("/v1/businessIntegration/fail-once/%s", formationID)

	testCases := []struct {
		Name                 string
		RequestBody          string
		FormationID          string
		ExpectedResponseCode int
		Method               string
		ExpectedMappings     map[string][]notification.Response
		ShouldFail           bool
	}{
		{
			Name:                 "Success for create formation when it shouldn't fail",
			RequestBody:          formationReqBody,
			FormationID:          formationID,
			Method:               http.MethodPost,
			ExpectedResponseCode: http.StatusOK,
			ShouldFail:           false,
			ExpectedMappings:     fixFormationMappings(notification.CreateFormation, formationID, formationReqBody),
		},
		{
			Name:                 "Success for delete formation when it shouldn't fail",
			RequestBody:          formationReqBody,
			FormationID:          formationID,
			Method:               http.MethodDelete,
			ExpectedResponseCode: http.StatusOK,
			ShouldFail:           false,
			ExpectedMappings:     fixFormationMappings(notification.DeleteFormation, formationID, formationReqBody),
		},
		{
			Name:                 "Success for create formation when it should fail",
			RequestBody:          formationReqBody,
			FormationID:          formationID,
			Method:               http.MethodPost,
			ExpectedResponseCode: http.StatusBadRequest,
			ShouldFail:           true,
			ExpectedMappings:     fixFormationMappings(notification.CreateFormation, formationID, formationReqBody),
		},
		{
			Name:                 "Success for delete formation when it should fail",
			RequestBody:          formationReqBody,
			FormationID:          formationID,
			Method:               http.MethodDelete,
			ExpectedResponseCode: http.StatusBadRequest,
			ShouldFail:           true,
			ExpectedMappings:     fixFormationMappings(notification.DeleteFormation, formationID, formationReqBody),
		},
		{
			Name:                 "Error when required formationID path parameter is missing",
			Method:               http.MethodPost,
			ExpectedResponseCode: http.StatusBadRequest,
			ShouldFail:           false,
			ExpectedMappings:     map[string][]notification.Response{},
		},
		{
			Name:                 "Error when formation request body is not valid json",
			RequestBody:          "invalid json",
			Method:               http.MethodPost,
			FormationID:          formationID,
			ExpectedResponseCode: http.StatusBadRequest,
			ShouldFail:           false,
			ExpectedMappings:     map[string][]notification.Response{formationID: {}},
		},
		{
			Name:                 "Error when required formationID path parameter is missing when should fail",
			Method:               http.MethodPost,
			ExpectedResponseCode: http.StatusBadRequest,
			ShouldFail:           true,
			ExpectedMappings:     map[string][]notification.Response{},
		},
		{
			Name:                 "Error when formation request body is not valid json when should fail",
			RequestBody:          "invalid json",
			Method:               http.MethodPost,
			FormationID:          formationID,
			ExpectedResponseCode: http.StatusBadRequest,
			ShouldFail:           true,
			ExpectedMappings:     map[string][]notification.Response{formationID: {}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			req, err := http.NewRequest(testCase.Method, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)

			if testCase.FormationID != "" {
				req = mux.SetURLVars(req, map[string]string{formationIDParam: testCase.FormationID})
			}

			handler := notification.NewHandler(notification.NotificationsConfiguration{})
			handler.ShouldReturnError = testCase.ShouldFail
			recorder := httptest.NewRecorder()

			//WHEN
			handler.FailOnceFormation(recorder, req)
			resp := recorder.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedMappings, handler.Mappings)
		})
	}
}

func TestKymaBasicCredentials(t *testing.T) {
	t.Run("When method is PATCH", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPatch, url, nil)
		require.NoError(t, err)

		h := notification.NewHandler(notification.NotificationsConfiguration{})
		r := httptest.NewRecorder()

		//WHEN
		h.KymaBasicCredentials(r, req)
		resp := r.Result()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
		expectedBody := []byte("{\"state\":\"READY\",\"configuration\":{\"credentials\":{\"outboundCommunication\":{\"basicAuthentication\":{\"username\":\"user\",\"password\":\"pass\"},\"oauth2ClientCredentials\":{\"tokenServiceUrl\":\"\",\"clientId\":\"\",\"clientSecret\":\"\"}}}}}\n")
		require.Equal(t, expectedBody, body)
	})
	t.Run("When method is DELETE", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)

		h := notification.NewHandler(notification.NotificationsConfiguration{})
		r := httptest.NewRecorder()

		//WHEN
		h.KymaBasicCredentials(r, req)
		resp := r.Result()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	})
}

func TestOauthBasicCredentials(t *testing.T) {
	t.Run("When method is PATCH", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPatch, url, nil)
		require.NoError(t, err)

		h := notification.NewHandler(notification.NotificationsConfiguration{})
		r := httptest.NewRecorder()

		//WHEN
		h.KymaOauthCredentials(r, req)
		resp := r.Result()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
		expectedBody := []byte("{\"state\":\"READY\",\"configuration\":{\"credentials\":{\"outboundCommunication\":{\"basicAuthentication\":{\"username\":\"\",\"password\":\"\"},\"oauth2ClientCredentials\":{\"tokenServiceUrl\":\"url\",\"clientId\":\"id\",\"clientSecret\":\"secret\"}}}}}\n")
		require.Equal(t, expectedBody, body)
	})
	t.Run("When method is DELETE", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)

		h := notification.NewHandler(notification.NotificationsConfiguration{})
		r := httptest.NewRecorder()

		//WHEN
		h.KymaOauthCredentials(r, req)
		resp := r.Result()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	})
}