package provider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var url = "https://target-url.com"

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
	apiPath := "/v1/dependencies/configure"
	dependency := "dependency-name"

	testCases := []struct {
		Name                 string
		RequestMethod        string
		RequestBody          string
		ExpectedResponseCode int
		ExpectedBody         string
	}{
		{
			Name:                 "Error when request method is not the expected one",
			RequestMethod:        http.MethodPut,
			ExpectedResponseCode: http.StatusMethodNotAllowed,
		},
		{
			Name:                 "Error when the request body is empty",
			RequestMethod:        http.MethodPost,
			RequestBody:          "",
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedBody:         "{\"error\":\"The request body is empty\"}\n",
		},
		{
			Name:                 "Successfully handled dependency configure request",
			RequestMethod:        http.MethodPost,
			RequestBody:          dependency,
			ExpectedResponseCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			depConfigureReq, err := http.NewRequest(testCase.RequestMethod, url+apiPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
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
	depConfApiPath := "/v1/dependencies/configure"
	depApiPath := "/v1/dependencies"
	dependency := "dependency-name"

	t.Run("Successfully handled get dependency request", func(t *testing.T) {
		//GIVEN
		depConfigureReq, err := http.NewRequest(http.MethodPost, url+depConfApiPath, bytes.NewBuffer([]byte(dependency)))
		require.NoError(t, err)
		depReq, err := http.NewRequest(http.MethodGet, url+depApiPath, bytes.NewBuffer([]byte{}))
		require.NoError(t, err)
		h := NewHandler("")

		//WHEN
		r := httptest.NewRecorder()
		h.DependenciesConfigure(r, depConfigureReq)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// WHEN
		r = httptest.NewRecorder()
		h.Dependencies(r, depReq)
		resp = r.Result()

		//THEN
		expectedBody := fmt.Sprintf("[{\"xsappname\":\"%s\"}]", dependency)
		assertExpectedResponse(t, resp, expectedBody, http.StatusOK)
	})
}

func TestHandler_DependenciesIndirect(t *testing.T) {
	depApiPath := "/v1/dependencies/indirect"
	indirectDependency := "indirect-dependency-name"

	t.Run("Successfully handled get dependency request", func(t *testing.T) {
		//GIVEN
		depReq, err := http.NewRequest(http.MethodGet, url+depApiPath, bytes.NewBuffer([]byte{}))
		require.NoError(t, err)
		h := NewHandler(indirectDependency)

		//WHEN
		r := httptest.NewRecorder()
		h.DependenciesIndirect(r, depReq)
		resp := r.Result()

		//THEN
		expectedBody := fmt.Sprintf("[{\"xsappname\":\"%s\"}]", indirectDependency)
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
