package oauth20_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestService_CreateClient(t *testing.T) {

	//uidSvc := &automock.UIDService{}
	//uidSvc.On("Generate").Return(id).Once()
	//defer uidSvc.AssertExpectations(t)

	t.Log("Not implemented")
	t.FailNow()

}

func TestService_DeleteClient(t *testing.T) {
	// when
	id := "foo"
	remoteURL := "foo.bar/clients"
	cfg := oauth20.Config{ClientEndpoint: remoteURL}

	givenURL, err := url.Parse(fmt.Sprintf("%s/%s", remoteURL, id))
	require.NoError(t, err)
	req := fixHTTPRequest("DELETE", givenURL)
	successRes := &http.Response{StatusCode: http.StatusNoContent}
	wrongStatusCodeRes := &http.Response{StatusCode: http.StatusInternalServerError}
	testErr := errors.New("Test err")

	testCases := []struct {
		Name          string
		ExpectedError error
		HTTPClientFn  func() *automock.HTTPClient
		Config oauth20.Config
		Request *http.Request
		Response *http.Response
	}{
		{
			Name:          "Success",
			ExpectedError: nil,
			HTTPClientFn: func() *automock.HTTPClient {
				httpCli := &automock.HTTPClient{}
				httpCli.On("Do", req).Return(successRes, nil).Once()
				return httpCli
			},
		},
		{
			Name:          "Error - Response Status Code",
			ExpectedError: errors.New("invalid HTTP status code: received: 500, expected 204"),
			HTTPClientFn: func() *automock.HTTPClient {
				httpCli := &automock.HTTPClient{}
				httpCli.On("Do", req).Return(wrongStatusCodeRes, nil).Once()
				return httpCli
			},
		},
		{
			Name:          "Error - HTTP call error",
			ExpectedError: errors.New("while doing request to foo.bar/clients: Test err"),
			HTTPClientFn: func() *automock.HTTPClient {
				httpCli := &automock.HTTPClient{}
				httpCli.On("Do", req).Return(nil, testErr).Once()
				return httpCli
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			httpCli := testCase.HTTPClientFn()
			defer httpCli.AssertExpectations(t)

			svc := oauth20.NewService(nil, nil, cfg)
			svc.SetHTTPClient(httpCli)

			// when
			err := svc.DeleteClient(ctx, id)

			// then
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedError.Error(), err.Error())
			}

		})
	}

}

func fixHTTPRequest(method string, url *url.URL) *http.Request {
	return &http.Request{Method: method,
		URL:        url,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{"Accept": []string{"application/json"},
			"Content-Type": []string{"application/json"}},
		Body: io.ReadCloser(nil),
	}
}
