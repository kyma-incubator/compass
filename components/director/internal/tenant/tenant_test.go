package tenant_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadFromContext(t *testing.T) {
	value := "foo"

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult     string
		ExpectedErrMessage string
	}{
		{
			Name:               "Success",
			Context:            context.WithValue(context.TODO(), tenant.TenantContextKey, value),
			ExpectedResult:     value,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Error",
			Context:            context.TODO(),
			ExpectedResult:     "",
			ExpectedErrMessage: "Cannot read tenant from context",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result, err := tenant.LoadFromContext(testCase.Context)

			// then
			if testCase.ExpectedErrMessage != "" {
				require.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			assert.Equal(t, testCase.ExpectedResult, result)
		})
	}
}

func TestSaveToLoadFromContext(t *testing.T) {
	// given
	value := "foo"
	ctx := context.TODO()

	// when
	result := tenant.SaveToContext(ctx, value)

	// then
	assert.Equal(t, value, result.Value(tenant.TenantContextKey))
}

func TestRequireHeaderMiddleware(t *testing.T) {
	body := "Body"
	sampleTenant := "foo"

	successHandler := func(t *testing.T, tenantID string) func(http.ResponseWriter, *http.Request) {
		return func(writer http.ResponseWriter, request *http.Request) {
			b, err := ioutil.ReadAll(request.Body)
			require.NoError(t, err)
			assert.Equal(t, []byte(body), b)

			if tenantID != "" {
				log.Printf("%+v", request.Context())
				assert.Equal(t, tenantID, request.Context().Value(tenant.TenantContextKey))
			}

			defer func() {
				err := request.Body.Close()
				if err != nil {
					panic(err)
				}
			}()
		}
	}

	failHandler := func(t *testing.T) func(writer http.ResponseWriter, request *http.Request) {
		return func(writer http.ResponseWriter, request *http.Request) {
			t.Fail()
		}
	}

	testCases := []struct {
		Name    string
		HandlerFn func(t *testing.T) http.HandlerFunc
		InputRequestFn         func() *http.Request
		ExpectedStatusCode     int
		ExpectedMessage        string
		ExpectedCtxTenantValue string
	}{
		{
			Name: "GET without tenant",
			InputRequestFn: func() *http.Request {
				req, err := fixHTTPRequest("GET", body, map[string][]string{})
				require.NoError(t, err)
				return req
			},
			HandlerFn: func(t *testing.T) http.HandlerFunc {
				return successHandler(t, "")
			},
			ExpectedStatusCode:     http.StatusOK,
			ExpectedMessage:        "",
			ExpectedCtxTenantValue: "",
		},
		{
			Name: "POST request with tenant",
			InputRequestFn: func() *http.Request {
				req, err := fixHTTPRequest("POST", body, map[string][]string{
					"Tenant": {sampleTenant},
				})
				require.NoError(t, err)
				return req
			},
			HandlerFn: func(t *testing.T) http.HandlerFunc {
				return successHandler(t, sampleTenant)
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedMessage:    "",
		},
		{
			Name: "POST request with multiple values of tenant",
			InputRequestFn: func() *http.Request {
				req, err := fixHTTPRequest("POST", "Body", map[string][]string{
					"Tenant": {sampleTenant, "bar"},
				})
				require.NoError(t, err)
				return req
			},
			HandlerFn: func(t *testing.T) http.HandlerFunc {
				return successHandler(t, sampleTenant)
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedMessage:    "",
		},
		{
			Name: "POST without tenant",
			InputRequestFn: func() *http.Request {
				req, err := fixHTTPRequest("POST", "Body", map[string][]string{})
				require.NoError(t, err)
				return req
			},
			HandlerFn: func(t *testing.T) http.HandlerFunc {
				return failHandler(t)
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedMessage:    "Header `tenant` is required\n",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			middleware := tenant.RequireHeaderMiddleware(testCase.HandlerFn(t))

			// when
			middleware.ServeHTTP(recorder, testCase.InputRequestFn())

			// then
			resp := recorder.Result()
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					panic(err)
				}
			}()

			reqBody, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.ExpectedMessage, string(reqBody))
		})
	}
}

func fixHTTPRequest(reqType, body string, additionalHeaders map[string][]string) (*http.Request, error) {
	rq, err := http.NewRequest(reqType, "foo.bar", bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}

	rq.Header = additionalHeaders
	rq.Header.Add("Content-Type", "application/json")
	return rq, nil
}
