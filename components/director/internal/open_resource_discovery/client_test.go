package ord_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
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

func TestClient_FetchOpenResourceDiscoveryDocuments(t *testing.T) {
	testCases := []struct {
		Name           string
		RoundTripFunc  func(req *http.Request) *http.Response
		ExpectedResult ord.Documents
		ExpectedErr    error
	}{
		{
			Name: "Success",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				var err error
				statusCode := http.StatusOK
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					data, err = json.Marshal(fixWellKnownConfig())
					require.NoError(t, err)
				} else if strings.Contains(req.URL.String(), ordDocURI) {
					data, err = json.Marshal(fixORDDocument())
					require.NoError(t, err)
				} else {
					statusCode = http.StatusNotFound
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
		},
		{
			Name: "Error fetching well-known config",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				statusCode := http.StatusNotFound
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					statusCode = http.StatusInternalServerError
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedErr: errors.New("error while fetching open resource discovery well-known configuration: status code 500"),
		},
		{
			Name: "Error when well-known config is not proper json",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				statusCode := http.StatusNotFound
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					statusCode = http.StatusOK
					data = []byte("test")
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedErr: errors.New("error unmarshaling json body"),
		},
		{
			Name: "Document with unsupported access strategy is skipped",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				var err error
				statusCode := http.StatusOK
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					config := fixWellKnownConfig()
					config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = "custom"
					config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].CustomType = "test"
					data, err = json.Marshal(config)
					require.NoError(t, err)
				} else if strings.Contains(req.URL.String(), ordDocURI) {
					require.FailNow(t, "document with unsupported access strategy should not be fetched")
				} else {
					statusCode = http.StatusNotFound
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedResult: ord.Documents{},
		},
		{
			Name: "Error fetching document",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				var err error
				statusCode := http.StatusOK
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					data, err = json.Marshal(fixWellKnownConfig())
					require.NoError(t, err)
				} else if strings.Contains(req.URL.String(), ordDocURI) {
					statusCode = http.StatusInternalServerError
				} else {
					statusCode = http.StatusNotFound
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedResult: ord.Documents{},
			ExpectedErr:    errors.Errorf("error while fetching open resource discovery document %q: status code %d", baseURL+ordDocURI, 500),
		},
		{
			Name: "Error when document is not proper json",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				var err error
				statusCode := http.StatusOK
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					data, err = json.Marshal(fixWellKnownConfig())
					require.NoError(t, err)
				} else if strings.Contains(req.URL.String(), ordDocURI) {
					data = []byte("test")
				} else {
					statusCode = http.StatusNotFound
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedResult: ord.Documents{},
			ExpectedErr:    errors.New("error unmarshaling document"),
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			testHTTPClient := NewTestClient(test.RoundTripFunc)

			client := ord.NewClient(testHTTPClient)
			docs, err := client.FetchOpenResourceDiscoveryDocuments(context.TODO(), baseURL)

			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Len(t, docs, len(test.ExpectedResult))
				require.Equal(t, test.ExpectedResult, docs)
			}
		})
	}
}
