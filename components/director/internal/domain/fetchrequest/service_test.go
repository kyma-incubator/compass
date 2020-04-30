package fetchrequest_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil

}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestService_FetchAPISpec(t *testing.T) {
	mockSpec := "spec"
	timestamp := time.Now()

	modelInput := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Condition: model.FetchRequestStatusConditionInitial},
	}

	modelInputPackage := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModePackage,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Condition: model.FetchRequestStatusConditionInitial},
	}

	testCases := []struct {
		Name                    string
		RoundTripFn             func() RoundTripFunc
		Input                   model.FetchRequest
		ExpectedOutput          *string
		ExpectedStatusCondition model.FetchRequestStatusCondition
		ExpectedStatusMessage   *string
		ExpectedLog             *string
	}{
		{
			Name: "Success",
			RoundTripFn: func() RoundTripFunc {
				return func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
					}
				}
			},
			Input:                   modelInput,
			ExpectedOutput:          &mockSpec,
			ExpectedStatusCondition: model.FetchRequestStatusConditionSucceeded,
			ExpectedStatusMessage:   str.Ptr(""),
			ExpectedLog:             nil,
		},
		{
			Name: "Nil when mode is Package",
			RoundTripFn: func() RoundTripFunc {
				return func(req *http.Request) *http.Response {
					return &http.Response{}
				}
			},
			Input:                   modelInputPackage,
			ExpectedOutput:          nil,
			ExpectedStatusCondition: model.FetchRequestStatusConditionInitial,
			ExpectedStatusMessage:   str.Ptr("Unsupported fetch mode: PACKAGE"),
			ExpectedLog:             str.Ptr("Unsupported fetch mode: PACKAGE"),
		},
		{
			Name: "Error when fetching",
			RoundTripFn: func() RoundTripFunc {
				return func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
					}
				}
			},
			Input:                   modelInput,
			ExpectedOutput:          nil,
			ExpectedStatusCondition: model.FetchRequestStatusConditionFailed,
			ExpectedStatusMessage:   str.Ptr(fmt.Sprintf("While fetching API Spec status code: %d", http.StatusInternalServerError)),
			ExpectedLog:             str.Ptr(fmt.Sprintf("While fetching API Spec status code: %d", http.StatusInternalServerError)),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			client := NewTestClient(testCase.RoundTripFn())
			var actualLog bytes.Buffer
			logger := log.New()
			logger.SetFormatter(&log.TextFormatter{
				DisableTimestamp: true,
			})
			logger.SetOutput(&actualLog)

			svc := fetchrequest.NewService(client, logger)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			spec, status := svc.FetchAPISpec(&testCase.Input)

			if testCase.ExpectedLog != nil {
				expectedLog := fmt.Sprintf("level=error msg=\"%s\"\n", *testCase.ExpectedLog)
				assert.Equal(t, expectedLog, actualLog.String())
			}
			assert.Equal(t, testCase.ExpectedStatusCondition, status.Condition)
			assert.Equal(t, testCase.ExpectedStatusMessage, status.Message)

			assert.Equal(t, testCase.ExpectedOutput, spec)

		})
	}

}
