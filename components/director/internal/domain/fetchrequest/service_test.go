package fetchrequest_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	logrustest "github.com/sirupsen/logrus/hooks/test"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type nopCloser struct {
	io.Reader
}

func (n *nopCloser) Close() error {
	return nil
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	response := f(req)
	if response.StatusCode == http.StatusBadRequest {
		return nil, errors.New("error")
	}
	return response, nil

}

func NewTestClient(fn RoundTripFunc) *http.Client {
	//TODO check for error in request
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
	logger, hook := logrustest.NewNullLogger()
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

	modelInputFilter := model.FetchRequest{
		ID:     "test",
		Mode:   model.FetchModeSingle,
		Filter: str.Ptr("filter"),
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Condition: model.FetchRequestStatusConditionInitial,
		},
	}

	modelInputMissingCredentials := model.FetchRequest{
		ID: "test",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: nil,
			},
		},
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Condition: model.FetchRequestStatusConditionInitial,
		},
	}

	modelInputBasicCredentials := model.FetchRequest{
		ID: "test",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: &model.BasicCredentialData{
					Username: "username",
					Password: "password",
				},
			},
		},
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Condition: model.FetchRequestStatusConditionSucceeded,
		},
	}

	modelInputOauth := model.FetchRequest{
		ID: "test",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: &model.OAuthCredentialData{
					ClientID:     "clId",
					ClientSecret: "clSecret",
					URL:          "mocked-url/.well-known/openid-configuration",
				},
			},
		},
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			//TODO check condition
			Condition: model.FetchRequestStatusConditionSucceeded,
		},
	}

	modelInputPackageWithMessage := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModePackage,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Message:   str.Ptr("Invalid data [reason=Unsupported fetch mode: PACKAGE]"),
			Condition: model.FetchRequestStatusConditionInitial},
	}

	modelInputFilterWithMessage := model.FetchRequest{
		ID:     "test",
		Mode:   model.FetchModeSingle,
		Filter: str.Ptr("filter"),
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Message:   str.Ptr("Invalid data [reason=Filter for Fetch Request was provided, currently it's unsupported]"),
			Condition: model.FetchRequestStatusConditionInitial},
	}

	modelInputMissingCredentialsWithMessage := model.FetchRequest{
		ID: "test",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: nil,
			},
		},
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Message:   str.Ptr("While fetching API Spec: Invalid data [reason=Credentials not provided]"),
			Condition: model.FetchRequestStatusConditionFailed},
	}

	modelInputOauthWithMessage := model.FetchRequest{
		ID: "test",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: &model.OAuthCredentialData{
					ClientID:     "clId",
					ClientSecret: "clSecret",
					URL:          "mocked-url/.well-known/openid-configuration",
				},
			},
		},
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Message:   str.Ptr("While fetching API Spec: oauth2: cannot fetch token: \nResponse: {\"token_endpoint\": \"fail\"}"),
			Condition: model.FetchRequestStatusConditionFailed},
	}

	modelInputOauthWithMessageFailedToGetTokenEndpoint := model.FetchRequest{
		ID: "test",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: &model.OAuthCredentialData{
					ClientID:     "clId",
					ClientSecret: "clSecret",
					URL:          "mocked-url/.well-known/openid-configuration",
				},
			},
		},
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Timestamp: timestamp,
			Message:   str.Ptr("While fetching API Spec: Get \"mocked-url/.well-known/openid-configuration/.well-known/openid-configuration\": error"),
			Condition: model.FetchRequestStatusConditionFailed},
	}

	testCases := []struct {
		Name                     string
		Client                   *http.Client
		InputAPI                 model.APIDefinition
		FetchRequestRepoFn       func() *automock.FetchRequestRepository
		InputFr                  model.FetchRequest
		ExpectedOutput           *string
		ExpectedMessage          *string
		ExpectedError            *string
		ExpectedBasicCredentials model.BasicCredentialData
		ExpectedTokenURL         fetchrequest.OpenIDMetadata
	}{

		{
			Name: "Success",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
				}
			}),
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputSucceeded).Return(nil).Once()
				return repo
			},
			InputFr:        modelInput,
			ExpectedOutput: &mockSpec,
		},
		{
			Name: "Nil when fetch request validation fails due to mode Package",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{}
			}),
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputPackageWithMessage).Return(nil).Once()
				return repo
			},
			InputFr:        modelInputPackage,
			ExpectedOutput: nil,
			ExpectedError: str.Ptr("Invalid data [reason=Unsupported fetch mode: PACKAGE]"),
		},
		{
			Name: "Nil when fetch request validation fails due to provided filter",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{}
			}),
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputFilterWithMessage).Return(nil).Once()
				return repo
			},
			InputFr:        modelInputFilter,
			ExpectedOutput: nil,
		},
		{
			Name: "Nil when auth without credentials is provided",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{}
			}),
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputMissingCredentialsWithMessage).Return(nil).Once()
				return repo
			},
			InputFr:        modelInputMissingCredentials,
			ExpectedOutput: nil,
		},
		{
			Name: "Error when fetch request fails",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
				}
			}),
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputFailed).Return(nil).Once()
				return repo
			},
			InputFr:         modelInput,
			ExpectedMessage: str.Ptr(fmt.Sprintf("While fetching API Spec status code: %d", http.StatusInternalServerError)),
			ExpectedOutput:  nil,
		},
		{
			Name: "Success when basic credentials are provided",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				username, password, _ := req.BasicAuth()
				body := bytes.NewBuffer(
					[]byte(
						fmt.Sprintf(`{"username": "%s", "password":"%s"}`, username, password),
					),
				)
				return &http.Response{
					Body: &nopCloser{
						Reader: body,
					},
					StatusCode: http.StatusOK,
				}
			}),
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputBasicCredentials).Return(nil).Once()
				return repo
			},
			InputFr:        modelInputBasicCredentials,
			ExpectedOutput: nil,
			ExpectedBasicCredentials: model.BasicCredentialData{
				Username: "username",
				Password: "password",
			},
		},
		{
			Name: "when fail to fetch token endpoint",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
				}
			}),
			FetchRequestRepoFn:
			func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputOauthWithMessageFailedToGetTokenEndpoint).Return(nil).Once()
				return repo
			},
			InputFr:        modelInputOauth,
			ExpectedOutput: nil,
			ExpectedTokenURL: fetchrequest.OpenIDMetadata{
				TokenEndpoint: "http://mocked-oauth.com/oauth/token",
			},
		},

		{
			Name: "when fail to get token",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				if strings.HasSuffix(req.URL.String(), "/.well-known/openid-configuration") {
					body := bytes.NewBuffer(
						[]byte(
							fmt.Sprintf(`{"token_endpoint": "%s"}`, "http://mocked-oauth.com/oauth/token"),
						),
					)
					return &http.Response{
						Body: &nopCloser{
							Reader: body,
						},
						StatusCode: http.StatusOK,
					}
				}

				body := bytes.NewBuffer(
					[]byte(
						fmt.Sprintf(`{"token_endpoint": "%s"}`, "fail"),
					),
				)
				return &http.Response{
					Body: &nopCloser{
						Reader: body,
					},
					StatusCode: http.StatusInternalServerError,
				}
			}),
			FetchRequestRepoFn:
			func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputOauthWithMessage).Return(nil).Once()
				return repo
			},
			InputFr:        modelInputOauth,
			ExpectedOutput: nil,
			ExpectedTokenURL: fetchrequest.OpenIDMetadata{
				TokenEndpoint: "http://mocked-oauth.com/oauth/token",
			},
		},

		{
			Name: "when succeed to fetch token",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				if strings.HasSuffix(req.URL.String(), "/.well-known/openid-configuration") {
					body := bytes.NewBuffer(
						[]byte(
							fmt.Sprintf(`{"token_endpoint": "%s"}`, "http://mocked-oauth.com/oauth/token"),
						),
					)
					return &http.Response{
						Body: &nopCloser{
							Reader: body,
						},
						StatusCode: http.StatusOK,
					}
				}

				if req.URL.String() == "http://mocked-oauth.com/oauth/token" {
					body := bytes.NewBuffer(
						[]byte(
							fmt.Sprint(`{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNzE2MjM5MDIyfQ.gFgwPdkGPEcQGjoy934vFv9pyjVO6e_18MAF7Fpf9kI"}`),
						),
					)
					return &http.Response{
						Body: &nopCloser{
							Reader: body,
						},
						StatusCode: http.StatusOK,
					}
				}
				return &http.Response{
					StatusCode: http.StatusOK,
				}
			}),
			FetchRequestRepoFn:
			func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputOauth).Return(nil).Once()
				return repo
			},
			InputFr:        modelInputOauth,
			ExpectedOutput: nil,
			ExpectedTokenURL: fetchrequest.OpenIDMetadata{
				TokenEndpoint: "http://mocked-oauth.com/oauth/token",
			},
			//ExpectedError:   str.Ptr(testErr.Error()),
		},
		{
			Name: "Nil when failed to update fetch request status",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
				}
			}),
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Update", ctx, &modelInputSucceeded).Return(testErr).Once()
				return repo
			},
			InputFr:         modelInput,
			ExpectedOutput:  nil,
			ExpectedMessage: str.Ptr(fmt.Sprintf("An error has occurred while updating fetch request status.")),
			ExpectedError:   str.Ptr(testErr.Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actualLog.Reset()

			frRepo := testCase.FetchRequestRepoFn()

			svc := fetchrequest.NewService(frRepo, testCase.Client)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			output := svc.HandleSpec(ctx, &testCase.InputFr)

			if testCase.ExpectedMessage != nil {
				assert.Equal(t, *testCase.ExpectedMessage, hook.LastEntry().Message)
			}
			if testCase.ExpectedError != nil {
				assert.Equal(t, *testCase.ExpectedError, hook.LastEntry().Data["error"].(error).Error())
				hook.Reset()}
			//}else{
			//	assert.Equal(t, 0,len(hook.AllEntries()))
			//	hook.Reset()
			//}
			if testCase.ExpectedOutput != nil {
				assert.Equal(t, testCase.ExpectedOutput, output)
			}

		})
	}

}
