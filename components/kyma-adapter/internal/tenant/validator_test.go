package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authmiddleware "github.com/kyma-incubator/compass/components/director/pkg/auth-middleware"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/config"
	"github.com/stretchr/testify/require"
)

const (
	testCertSubject = "OU=testCertSubject"
	tenant          = "testCertSubject"
)

func TestNewTenantValidationMiddleware(t *testing.T) {
	ctx := context.TODO()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"certSubject":"` + testCertSubject + `"}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	serverURL := server.URL

	testCases := []struct {
		Name                string
		Config              config.TenantInfo
		ExpectedCertSubject string
		ExpectedError       string
	}{
		{
			Name:                "Success",
			Config:              getTestConfig(serverURL),
			ExpectedCertSubject: testCertSubject,
		},
		{
			Name:          "Error while building request",
			Config:        getTestConfigWithBrokenURL(),
			ExpectedError: "failed to get auth tenant",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantValidationMiddleware, err := NewMiddleware(ctx, testCase.Config)

			if testCase.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedCertSubject, tenantValidationMiddleware.certSubject)
			}
		})
	}
}

func TestNewHTTPHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"certSubject":"` + testCertSubject + `"}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	serverURL := server.URL
	tenantValidationMiddleware, err := NewMiddleware(context.Background(), getTestConfig(serverURL))
	require.NoError(t, err)

	testCases := []struct {
		Name               string
		Context            context.Context
		ExpectedStatusCode int
		ExpectedResponse   string
	}{
		{
			Name:               "Success",
			Context:            authmiddleware.SaveToContext(context.Background(), tenant, tenant),
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse:   "OK",
		},
		{
			Name:               "Bad request when tenant is not present",
			Context:            context.Background(),
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   "cannot read tenant from context",
		},
		{
			Name:               "Unauthorized",
			Context:            authmiddleware.SaveToContext(context.Background(), "random", "random"),
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedResponse:   "Tenant random is not authorized",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(testCase.Context, "GET", "/", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			tenantValidationMiddleware.Handler()(testHandlerWithClientUser(t)).ServeHTTP(rr, req)

			require.Equal(t, testCase.ExpectedStatusCode, rr.Code)
			require.Contains(t, rr.Body.String(), testCase.ExpectedResponse)
		})
	}
}

func testHandlerWithClientUser(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}
