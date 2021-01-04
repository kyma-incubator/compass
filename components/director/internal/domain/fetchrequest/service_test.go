package fetchrequest_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"

	"github.com/kyma-incubator/compass/components/director/internal/model"
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
	return &http.Client{
		Transport: fn,
	}
}

func TestService_HandleAPISpec(t *testing.T) {
	mockSpec := "spec"
	timestamp := time.Now()

	modelInput := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
	}

	modelInputPackage := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModePackage,
	}

	modelInputFilter := model.FetchRequest{
		ID:     "test",
		Mode:   model.FetchModeSingle,
		Filter: str.Ptr("filter"),
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
	}

	modelInputOauth := model.FetchRequest{
		ID:  "test",
		URL: "sth",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: &model.OAuthCredentialData{
					ClientID:     "clId",
					ClientSecret: "clSecret",
					URL:          "mocked-url/oauth/token",
				},
			},
		},
		Mode: model.FetchModeSingle,
	}

	testCases := []struct {
		Name   string
		Client *http.Client
		//FetchRequestRepoFn       func() *automock.FetchRequestRepository
		InputFr        model.FetchRequest
		ExpectedResult *string
		ExpectedStatus *model.FetchRequestStatus
		//ExpectedBasicCredentials model.BasicCredentialData
	}{

		{
			Name: "Success",
			Client: NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
				}
			}),
			InputFr:        modelInput,
			ExpectedResult: &mockSpec,
			ExpectedStatus:
		},
		//{
		//	Name: "Nil when fetch request validation fails due to mode Package",
		//	Client: NewTestClient(func(req *http.Request) *http.Response {
		//		return &http.Response{}
		//	}),
		//	FetchRequestRepoFn: func() *automock.FetchRequestRepository {
		//		repo := &automock.FetchRequestRepository{}
		//		repo.On("Update", ctx, &modelInputPackageWithMessage).Return(nil).Once()
		//		return repo
		//	},
		//	InputFr:        modelInputPackage,
		//	ExpectedResult: nil,
		//	ExpectedError:  str.Ptr("Invalid data [reason=Unsupported fetch mode: PACKAGE]"),
		//},
		//{
		//	Name: "Nil when fetch request validation fails due to provided filter",
		//	Client: NewTestClient(func(req *http.Request) *http.Response {
		//		return &http.Response{}
		//	}),
		//	FetchRequestRepoFn: func() *automock.FetchRequestRepository {
		//		repo := &automock.FetchRequestRepository{}
		//		repo.On("Update", ctx, &modelInputFilterWithMessage).Return(nil).Once()
		//		return repo
		//	},
		//	InputFr:        modelInputFilter,
		//	ExpectedResult: nil,
		//	ExpectedError:  str.Ptr("Invalid data [reason=Filter for Fetch Request was provided, currently it's unsupported]"),
		//},
		//{
		//	Name: "Nil when auth without credentials is provided",
		//	Client: NewTestClient(func(req *http.Request) *http.Response {
		//		return &http.Response{}
		//	}),
		//	FetchRequestRepoFn: func() *automock.FetchRequestRepository {
		//		repo := &automock.FetchRequestRepository{}
		//		repo.On("Update", ctx, &modelInputMissingCredentialsWithMessage).Return(nil).Once()
		//		return repo
		//	},
		//	InputFr:        modelInputMissingCredentials,
		//	ExpectedResult: nil,
		//	ExpectedError:  str.Ptr("Invalid data [reason=Credentials not provided]"),
		//},
		//{
		//	Name: "Error when fetch request fails",
		//	Client: NewTestClient(func(req *http.Request) *http.Response {
		//		return &http.Response{
		//			StatusCode: http.StatusInternalServerError,
		//		}
		//	}),
		//	FetchRequestRepoFn: func() *automock.FetchRequestRepository {
		//		repo := &automock.FetchRequestRepository{}
		//		repo.On("Update", ctx, &modelInputFailed).Return(nil).Once()
		//		return repo
		//	},
		//	InputFr:         modelInput,
		//	ExpectedMessage: str.Ptr(fmt.Sprintf("While fetching API Spec status code: %d", http.StatusInternalServerError)),
		//	ExpectedResult:  nil,
		//	ExpectedError:   str.Ptr("While fetching API Spec status code: 500"),
		//},
		//{
		//	Name: "Success when basic credentials are provided",
		//	Client: NewTestClient(func(req *http.Request) *http.Response {
		//		username, password, _ := req.BasicAuth()
		//		body := bytes.NewBuffer(
		//			[]byte(
		//				fmt.Sprintf(`{"username": "%s", "password":"%s"}`, username, password),
		//			),
		//		)
		//		return &http.Response{
		//			Body: &nopCloser{
		//				Reader: body,
		//			},
		//			StatusCode: http.StatusOK,
		//		}
		//	}),
		//	FetchRequestRepoFn: func() *automock.FetchRequestRepository {
		//		repo := &automock.FetchRequestRepository{}
		//		repo.On("Update", ctx, &modelInputBasicCredentials).Return(nil).Once()
		//		return repo
		//	},
		//	InputFr:        modelInputBasicCredentials,
		//	ExpectedResult: nil,
		//	ExpectedBasicCredentials: model.BasicCredentialData{
		//		Username: "username",
		//		Password: "password",
		//	},
		//},
		//{
		//	Name: "when fail to fetch token endpoint",
		//	Client: NewTestClient(func(req *http.Request) *http.Response {
		//		return &http.Response{
		//			StatusCode: http.StatusBadRequest,
		//		}
		//	}),
		//	FetchRequestRepoFn:
		//	func() *automock.FetchRequestRepository {
		//		repo := &automock.FetchRequestRepository{}
		//		repo.On("Update", ctx, &modelInputOauthWithMessageFailedToGetTokenEndpoint).Return(nil).Once()
		//		return repo
		//	},
		//	InputFr:        modelInputOauth,
		//	ExpectedResult: nil,
		//	ExpectedError:  str.Ptr("Get \\\"mocked-url//\\\": error"),
		//},
		//
		//{
		//	Name: "when fail to get token",
		//	Client: NewTestClient(func(req *http.Request) *http.Response {
		//		return &http.Response{
		//			StatusCode: http.StatusInternalServerError,
		//		}
		//	}),
		//	FetchRequestRepoFn:
		//	func() *automock.FetchRequestRepository {
		//		repo := &automock.FetchRequestRepository{}
		//		repo.On("Update", ctx, mock.Anything).Return(nil).Once()
		//		return repo
		//	},
		//	InputFr:        modelInputOauth,
		//	ExpectedResult: nil,
		//
		//	ExpectedError: str.Ptr("oauth2: cannot fetch token"),
		//},
		//
		//
		//{
		//	Name: "Nil when failed to update fetch request status",
		//	Client: NewTestClient(func(req *http.Request) *http.Response {
		//		return &http.Response{
		//			StatusCode: http.StatusOK,
		//			Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
		//		}
		//	}),
		//	FetchRequestRepoFn: func() *automock.FetchRequestRepository {
		//		repo := &automock.FetchRequestRepository{}
		//		repo.On("Update", ctx, &modelInputSucceeded).Return(testErr).Once()
		//		return repo
		//	},
		//	InputFr:         modelInput,
		//	ExpectedResult:  nil,
		//	ExpectedMessage: str.Ptr(fmt.Sprintf("An error has occurred while updating fetch request status.")),
		//	ExpectedError:   str.Ptr(testErr.Error()),
		//},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			frRepo := &automock.FetchRequestRepository{}
			frRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

			svc := fetchrequest.NewService(frRepo, testCase.Client)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			result := svc.HandleSpec(ctx, &testCase.InputFr)

			assert.Equal(t, testCase.ExpectedStatus, testCase.InputFr.Status)
			assert.Equal(t, testCase.ExpectedResult, result)

		})
	}

}
