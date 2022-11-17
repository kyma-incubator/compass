package http_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/http/httpfakes"
	"github.com/stretchr/testify/require"
)

func TestErrorHandlerTransport_RoundTripReturnsAnErrorOnBadRequest(t *testing.T) {
	const failedResponseBody = "failBody"

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripReturns(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewBufferString(failedResponseBody)),
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
		Body:       io.NopCloser(&bytes.Buffer{}),
	}, nil)

	errTransport := httputil.NewErrorHandlerTransport(transport)
	resp, err := errTransport.RoundTrip(&http.Request{})

	require.NotNil(t, resp)
	require.NoError(t, err)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyString := string(bodyBytes)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Empty(t, bodyString)
}
