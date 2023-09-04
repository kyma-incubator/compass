package ord_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TODO Rewrite the whole test ScheduleAggregationForORDData test
func TestHandler_AggregateORDData(t *testing.T) {
	apiPath := "/aggregate"
	metricsConfig := ord.MetricsConfig{}
	operationsManager := &operationsmanager.OperationsManager{}
	testErr := errors.New("test error")

	testCases := []struct {
		Name                string
		RequestBody         ord.AggregationResources
		ORDService          func() *automock.ORDService
		ExpectedErrorOutput string
		ExpectedStatusCode  int
	}{
		{
			Name: "Successful ORD data aggregation",
			RequestBody: ord.AggregationResources{
				ApplicationID:         appID,
				ApplicationTemplateID: appTemplateID,
			},
			ORDService: func() *automock.ORDService {
				svc := &automock.ORDService{}
				svc.On("ProcessApplications", mock.Anything, metricsConfig, appID).Return(nil)
				svc.On("ProcessApplicationTemplates", mock.Anything, metricsConfig, appTemplateID).Return(nil)
				return svc
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "Successful ORD data aggregation - empty appIDs and appTemplateIDs",
			RequestBody: ord.AggregationResources{
				ApplicationID:         "",
				ApplicationTemplateID: "",
			},
			ORDService: func() *automock.ORDService {
				svc := &automock.ORDService{}
				svc.On("ProcessApplications", mock.Anything, metricsConfig, "").Return(nil)
				svc.On("ProcessApplicationTemplates", mock.Anything, metricsConfig, "").Return(nil)
				return svc
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "Aggregation failed for one or more applications",
			RequestBody: ord.AggregationResources{
				ApplicationID:         "",
				ApplicationTemplateID: "",
			},
			ORDService: func() *automock.ORDService {
				svc := &automock.ORDService{}
				svc.On("ProcessApplications", mock.Anything, metricsConfig, []string{}).Return(testErr)
				return svc
			},
			ExpectedErrorOutput: "ORD data aggregation failed for one or more applications",
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			Name: "Aggregation failed for one or more application templates",
			RequestBody: ord.AggregationResources{
				ApplicationID:         "",
				ApplicationTemplateID: "",
			},
			ORDService: func() *automock.ORDService {
				svc := &automock.ORDService{}
				svc.On("ProcessApplications", mock.Anything, metricsConfig, []string{}).Return(nil)
				svc.On("ProcessApplicationTemplates", mock.Anything, metricsConfig, []string{}).Return(testErr)

				return svc
			},
			ExpectedErrorOutput: "ORD data aggregation failed for one or more application templates",
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ORDService()
			defer mock.AssertExpectationsForObjects(t, svc)

			handler := ord.NewORDAggregatorHTTPHandler(operationsManager, nil, metricsConfig)
			requestBody, err := json.Marshal(testCase.RequestBody)
			assert.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewBuffer(requestBody))
			writer := httptest.NewRecorder()

			// WHEN
			handler.ScheduleAggregationForORDData(writer, request)

			// THEN
			resp := writer.Result()
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
