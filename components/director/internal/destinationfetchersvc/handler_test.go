package destinationfetchersvc_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const expectedTenantID = "f09ba084-0e82-49ab-ab2e-b7ecc988312d"

func TestHandler_SyncDestinations(t *testing.T) {
	target := "/v1/fetch"

	validHandlerConfig := destinationfetchersvc.HandlerConfig{
		SyncDestinationsEndpoint:      "/v1/fetch",
		DestinationsSensitiveEndpoint: "/v1/info",
	}

	reqWithUserContext := httptest.NewRequest(http.MethodPut, target, nil)
	reqWithUserContext = reqWithUserContext.WithContext(
		tenant.SaveToContext(reqWithUserContext.Context(), expectedTenantID, ""))

	testCases := []struct {
		Name                string
		Request             *http.Request
		DestinationManager  func() *automock.DestinationManager
		ExpectedErrorOutput string
		ExpectedStatusCode  int
	}{
		{
			Name:    "Successful fetch on-demand",
			Request: reqWithUserContext,
			DestinationManager: func() *automock.DestinationManager {
				svc := &automock.DestinationManager{}
				svc.On("SyncTenantDestinations", mock.Anything, expectedTenantID).Return(nil)
				return svc
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:    "Missing tenant header",
			Request: httptest.NewRequest(http.MethodPut, target, nil),
			DestinationManager: func() *automock.DestinationManager {
				return &automock.DestinationManager{}
			},
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:    "Tenant not found",
			Request: reqWithUserContext,
			DestinationManager: func() *automock.DestinationManager {
				svc := &automock.DestinationManager{}
				err := apperrors.NewNotFoundErrorWithMessage(resource.Label,
					expectedTenantID, fmt.Sprintf("tenant %s not found", expectedTenantID))
				svc.On("SyncTenantDestinations", mock.Anything, expectedTenantID).Return(err)
				return svc
			},
			ExpectedErrorOutput: fmt.Sprintf("tenant %s not found", expectedTenantID),
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			Name:    "Internal Server Error",
			Request: reqWithUserContext,
			DestinationManager: func() *automock.DestinationManager {
				svc := &automock.DestinationManager{}
				err := fmt.Errorf("random error")
				svc.On("SyncTenantDestinations", mock.Anything, expectedTenantID).Return(err)
				return svc
			},
			ExpectedErrorOutput: fmt.Sprintf("Failed to sync destinations for tenant %s", expectedTenantID),
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tf := testCase.DestinationManager()
			defer mock.AssertExpectationsForObjects(t, tf)

			handler := destinationfetchersvc.NewDestinationsHTTPHandler(tf, validHandlerConfig)
			req := testCase.Request
			w := httptest.NewRecorder()

			// WHEN
			handler.SyncTenantDestinations(w, req)

			// THEN
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Contains(t, string(body), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)
		})
	}
}

func TestHandler_FetchDestinationsSensitiveData(t *testing.T) {
	const destQueryParameter = "name"

	json := []byte("{}")

	target := "/v1/info"

	validHandlerConfig := destinationfetchersvc.HandlerConfig{
		SyncDestinationsEndpoint:      "/v1/fetch",
		DestinationsSensitiveEndpoint: "/v1/info",
		DestinationsQueryParameter:    destQueryParameter,
	}

	namesQueryRaw := "name=Rand&name=Mat"
	names := []string{"Rand", "Mat"}
	reqWithUserContext := httptest.NewRequest(http.MethodPut, target, nil)
	reqWithUserContext = reqWithUserContext.WithContext(
		tenant.SaveToContext(reqWithUserContext.Context(), expectedTenantID, ""))

	testCases := []struct {
		Name                  string
		Request               *http.Request
		DestQuery             string
		DestinationFetcherSvc func() *automock.DestinationManager
		ExpectedErrorOutput   string
		ExpectedStatusCode    int
	}{
		{
			Name:      "Successful fetch data fetch",
			Request:   reqWithUserContext,
			DestQuery: namesQueryRaw,
			DestinationFetcherSvc: func() *automock.DestinationManager {
				svc := &automock.DestinationManager{}
				svc.On("FetchDestinationsSensitiveData", mock.Anything, expectedTenantID, names).
					Return(
						func(ctx context.Context, tenantID string, destNames []string) []byte {
							return json
						},
						func(ctx context.Context, tenantID string, destNames []string) error {
							return nil
						},
					)
				return svc
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:    "Missing tenant header",
			Request: httptest.NewRequest(http.MethodPut, target, nil),
			DestinationFetcherSvc: func() *automock.DestinationManager {
				return &automock.DestinationManager{}
			},
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:    "Missing destination query parameter.",
			Request: reqWithUserContext,
			DestinationFetcherSvc: func() *automock.DestinationManager {
				return &automock.DestinationManager{}
			},
			ExpectedStatusCode: http.StatusBadRequest,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tf := testCase.DestinationFetcherSvc()
			defer mock.AssertExpectationsForObjects(t, tf)

			handler := destinationfetchersvc.NewDestinationsHTTPHandler(tf, validHandlerConfig)
			req := testCase.Request
			//req is a pointer and the changes on the previous test are kept
			req.URL.RawQuery = ""
			if len(testCase.DestQuery) > 0 {
				req.URL.RawQuery = testCase.DestQuery
			}

			w := httptest.NewRecorder()

			// WHEN
			handler.FetchDestinationsSensitiveData(w, req)

			// THEN
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Contains(t, string(body), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)
		})
	}
}
