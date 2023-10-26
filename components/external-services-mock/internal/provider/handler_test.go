package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	url                  = "https://target-url.com"
	depConfApiPath       = "/v1/dependencies/configure"
	dependencies         = []string{"dependency-name"}
	multipleDependencies = []string{"dependency-name", "second-dependency"}
)

func TestHandler_OnSubscription(t *testing.T) {
	apiPath := fmt.Sprintf("/tenants/v1/regional/%s/callback/%s", "region", "tenantID")

	testCases := []struct {
		Name                 string
		RequestMethod        string
		RequestBody          string
		ExpectedResponseCode int
		ExpectedBody         string
	}{
		{
			Name:                 "with PUT request",
			RequestMethod:        http.MethodPut,
			ExpectedBody:         "https://github.com/kyma-incubator/compass",
			ExpectedResponseCode: http.StatusOK,
		},
		{
			Name:                 "with DELETE request",
			RequestMethod:        http.MethodDelete,
			ExpectedBody:         "https://github.com/kyma-incubator/compass",
			ExpectedResponseCode: http.StatusOK,
		},
		{
			Name:                 "with invalid request method",
			RequestMethod:        http.MethodPost,
			ExpectedResponseCode: http.StatusMethodNotAllowed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			onSubReq, err := http.NewRequest(testCase.RequestMethod, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)
			h := NewHandler("")
			r := httptest.NewRecorder()

			//WHEN
			h.OnSubscription(r, onSubReq)
			resp := r.Result()

			//THEN
			if len(testCase.ExpectedBody) > 0 {
				assertExpectedResponse(t, resp, testCase.ExpectedBody, testCase.ExpectedResponseCode)
			} else {
				require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode)
			}
		})
	}
}

func TestHandler_DependenciesConfigure(t *testing.T) {
	testCases := []struct {
		Name                 string
		RequestMethod        string
		RequestBody          any
		ExpectedResponseCode int
		ExpectedBody         string
	}{
		{
			Name:                 "Error when request method is not the expected one",
			RequestMethod:        http.MethodPut,
			ExpectedResponseCode: http.StatusMethodNotAllowed,
		},
		{
			Name:                 "Error when dependencies list is empty",
			RequestMethod:        http.MethodPost,
			RequestBody:          []string{""},
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedBody:         "{\"error\":\"The dependency list could not be empty. X-Request-Id: \"}\n",
		},
		{
			Name:                 "Error when unmarshalling request body fails",
			RequestMethod:        http.MethodPost,
			RequestBody:          "{invalid}",
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedBody:         "{\"error\":\"An error occurred while unmarshalling request body: json: cannot unmarshal string into Go value of type []string. X-Request-Id: \"}\n",
		},
		{
			Name:                 "Successfully handled dependency configure request",
			RequestMethod:        http.MethodPost,
			RequestBody:          dependencies,
			ExpectedResponseCode: http.StatusOK,
		},
		{
			Name:                 "Successfully handled dependency configure request with multiple dependencies",
			RequestMethod:        http.MethodPost,
			RequestBody:          multipleDependencies,
			ExpectedResponseCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			reqBody, err := json.Marshal(testCase.RequestBody)
			require.NoError(t, err)
			depConfigureReq, err := http.NewRequest(testCase.RequestMethod, url+depConfApiPath, bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			h := NewHandler("")
			r := httptest.NewRecorder()

			//WHEN
			h.DependenciesConfigure(r, depConfigureReq)
			resp := r.Result()

			//THEN
			if len(testCase.ExpectedBody) > 0 {
				assertExpectedResponse(t, resp, testCase.ExpectedBody, testCase.ExpectedResponseCode)
			} else {
				require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode)
			}
		})
	}
}

func TestHandler_Dependencies(t *testing.T) {
	depApiPath := "/v1/dependencies"

	testCases := []struct {
		Name                 string
		RequestBody          any
		ExpectedResponseCode int
		ExpectedResponse     string
	}{
		{
			Name:                 "Error when dependencies list is empty",
			RequestBody:          "{invalid}",
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedResponse:     "{\"error\":\"The dependency list could not be empty. X-Request-Id: \"}\n",
		},
		{
			Name:                 "Successfully handled get dependency request with one dependency",
			RequestBody:          dependencies,
			ExpectedResponseCode: http.StatusOK,
			ExpectedResponse:     fmt.Sprintf("[{\"xsappname\":\"%s\"}]", dependencies[0]),
		},
		{
			Name:                 "Successfully handled get dependency request with multiple dependencies",
			RequestBody:          multipleDependencies,
			ExpectedResponseCode: http.StatusOK,
			ExpectedResponse:     fmt.Sprintf("[{\"xsappname\":\"%s\"},{\"xsappname\":\"%s\"}]", multipleDependencies[0], multipleDependencies[1]),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			reqBody, err := json.Marshal(testCase.RequestBody)
			require.NoError(t, err)
			depConfigureReq, err := http.NewRequest(http.MethodPost, url+depConfApiPath, bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			depReq, err := http.NewRequest(http.MethodGet, url+depApiPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)
			h := NewHandler("")

			//WHEN
			r := httptest.NewRecorder()
			h.DependenciesConfigure(r, depConfigureReq)
			resp := r.Result()

			//THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode)

			// WHEN
			r = httptest.NewRecorder()
			h.Dependencies(r, depReq)
			resp = r.Result()

			//THEN
			assertExpectedResponse(t, resp, testCase.ExpectedResponse, testCase.ExpectedResponseCode)
		})
	}
}

func TestHandler_DependenciesIndirect(t *testing.T) {
	depApiPath := "/v1/dependencies/indirect"
	dependency := "direct-dependency-name"

	t.Run("Successfully handled get dependency request", func(t *testing.T) {
		//GIVEN
		depReq, err := http.NewRequest(http.MethodGet, url+depApiPath, bytes.NewBuffer([]byte{}))
		require.NoError(t, err)
		h := NewHandler(dependency)

		//WHEN
		r := httptest.NewRecorder()
		h.DependenciesIndirect(r, depReq)
		resp := r.Result()

		//THEN
		expectedBody := fmt.Sprintf("[{\"xsappname\":\"%s\"}]", dependency)
		assertExpectedResponse(t, resp, expectedBody, http.StatusOK)
	})
}

func assertExpectedResponse(t *testing.T, response *http.Response, expectedBody string, expectedStatusCode int) {
	require.Equal(t, expectedStatusCode, response.StatusCode)
	body, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
	require.Equal(t, expectedBody, string(body))
}
