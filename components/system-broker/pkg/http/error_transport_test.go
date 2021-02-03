package http_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/http/httpfakes"
	"github.com/stretchr/testify/require"

	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
)

func TestErrorHandlerTransport_RoundTripReturnsAnErrorOnBadRequest(t *testing.T) {
	const failedResponseBody = "failBody"

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripReturns(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(bytes.NewBufferString(failedResponseBody)),
	}, nil)

	errTransport := httputil.NewErrorHandlerTransport(transport)
	resp, err := errTransport.RoundTrip(&http.Request{})

	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("statusCode: %d Body: %s", http.StatusBadRequest, failedResponseBody))
}

func TestErrorHandlerTransport_RoundTripReturnsAValidResponseOnSuccessRequest(t *testing.T) {
	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripReturns(&http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(&bytes.Buffer{}),
	}, nil)

	errTransport := httputil.NewErrorHandlerTransport(transport)
	resp, err := errTransport.RoundTrip(&http.Request{})

	require.NotNil(t, resp)
	require.NoError(t, err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Empty(t, bodyString)
}
