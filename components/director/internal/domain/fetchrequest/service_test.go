package fetchrequest_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"
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

func TestService_HandleAPISpec(t *testing.T) {
	mockSpec := "spec"
	timestamp := time.Now()
	testErr := errors.New("test")
	ctx := context.TODO()
	var actualLog bytes.Buffer
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})
	logger.SetOutput(&actualLog)
	ctx = log.ContextWithLogger(ctx, logrus.NewEntry(logger))

	modelInput := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Condition: model.FetchRequestStatusConditionInitial},
	}

	modelInputSucceeded := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Condition: model.FetchRequestStatusConditionSucceeded},
	}

	modelInputFailed := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Message:   str.Ptr("While fetching API Spec status code: 500"),
			Condition: model.FetchRequestStatusConditionFailed},
	}

	modelInputPackage := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModePackage,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Condition: model.FetchRequestStatusConditionInitial},
	}
	modelInputPackageWithMessage := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModePackage,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Message:   str.Ptr("Invalid data [reason=Unsupported fetch mode: PACKAGE]"),
			Condition: model.FetchRequestStatusConditionInitial},
	}

	testCases := []struct {
		Name               string
		RoundTripFn        func() RoundTripFunc
		InputAPI           model.APIDefinition
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		InputFr            model.FetchRequest
		ExpectedOutput     *string
		ExpectedLog        *string
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
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputSucceeded).Return(nil).Once()
				return repo
			},
			InputFr:        modelInput,
			ExpectedOutput: &mockSpec,
		},
		{
			Name: "Nil when mode is Package",
			RoundTripFn: func() RoundTripFunc {
				return func(req *http.Request) *http.Response {
					return &http.Response{}
				}
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputPackageWithMessage).Return(nil).Once()
				return repo
			},
			InputFr:        modelInputPackage,
			ExpectedOutput: nil,
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
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputFailed).Return(nil).Once()
				return repo
			},
			InputFr:        modelInput,
			ExpectedLog:    str.Ptr(fmt.Sprintf("While fetching API Spec status code: %d", http.StatusInternalServerError)),
			ExpectedOutput: nil,
		}, {
			Name: "Nil when failed to update status",
			RoundTripFn: func() RoundTripFunc {
				return func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
					}
				}
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputSucceeded).Return(testErr).Once()
				return repo
			},
			InputFr:        modelInput,
			ExpectedLog:    str.Ptr(fmt.Sprintf("While updating fetch request status: %s", testErr.Error())),
			ExpectedOutput: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			client := NewTestClient(testCase.RoundTripFn())
			actualLog.Reset()

			frRepo := testCase.FetchRequestRepoFn()

			svc := fetchrequest.NewService(frRepo, client)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			output := svc.HandleAPISpec(ctx, &testCase.InputFr)

			if testCase.ExpectedLog != nil {
				expectedLog := fmt.Sprintf("level=error msg=\"%s\"\n", *testCase.ExpectedLog)
				assert.Equal(t, expectedLog, actualLog.String())
			}
			if testCase.ExpectedOutput != nil {
				assert.Equal(t, testCase.ExpectedOutput, output)
			}

		})
	}

}
