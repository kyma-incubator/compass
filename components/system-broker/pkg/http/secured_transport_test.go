package http_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/http/httpfakes"
	"github.com/stretchr/testify/require"
)

func TestSecuredTransport_RoundTripSuccessfullyObtainsAuthorizationFromCorrectAuthorizationProviderAndUsesIt(t *testing.T) {
	const accessToken = "accessToken"

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
		authHeader := req.Header.Get("Authorization")
		require.Equal(t, "Bearer "+accessToken, authHeader)

		return nil, nil
	}

	tokenAuthorizationProvider := &httpfakes.FakeAuthorizationProvider{}
	tokenAuthorizationProvider.MatchesReturns(true)
	tokenAuthorizationProvider.GetAuthorizationReturns("Bearer "+accessToken, nil)

	tokenAuthorizationProvider2 := &httpfakes.FakeAuthorizationProvider{}
	tokenAuthorizationProvider2.MatchesReturns(false)
	tokenAuthorizationProvider2.GetAuthorizationReturns("Bearer "+accessToken+"2", nil)

	tokenAuthorizationProvider3 := &httpfakes.FakeAuthorizationProvider{}
	tokenAuthorizationProvider3.MatchesReturns(true)
	tokenAuthorizationProvider3.GetAuthorizationReturns("Bearer "+accessToken+"3", nil)

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	securedTransport := httputil.NewSecuredTransport(transport, tokenAuthorizationProvider)
	_, err = securedTransport.RoundTrip(request)
	require.NoError(t, err)
}

func TestSecuredTransport_RoundTripCouldNotObtainAuthorizationWhenNoAuthorizationProviderMatches(t *testing.T) {
	const accessToken = "accessToken"

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
		authHeader := req.Header.Get("Authorization")
		require.Empty(t, authHeader)

		return nil, nil
	}

	tokenAuthorizationProvider := &httpfakes.FakeAuthorizationProvider{}
	tokenAuthorizationProvider.MatchesReturns(false)
	tokenAuthorizationProvider.GetAuthorizationReturns("Bearer "+accessToken, nil)

	tokenAuthorizationProvider2 := &httpfakes.FakeAuthorizationProvider{}
	tokenAuthorizationProvider2.MatchesReturns(false)
	tokenAuthorizationProvider2.GetAuthorizationReturns("Bearer "+accessToken+"2", nil)

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	securedTransport := httputil.NewSecuredTransport(transport, tokenAuthorizationProvider)
	_, err = securedTransport.RoundTrip(request)
	require.EqualError(t, err, "context did not match any authorization provider")
	require.Equal(t, request.URL, testUrl)
}

func TestSecuredTransport_RoundTripFailsOnAuthorizationProviderError(t *testing.T) {
	transport := &httpfakes.FakeHTTPRoundTripper{}

	tokenAuthorizationProvider := &httpfakes.FakeAuthorizationProvider{}
	tokenAuthorizationProvider.MatchesReturns(true)
	tokenAuthorizationProvider.GetAuthorizationReturns("", errors.New("error"))

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	securedTransport := httputil.NewSecuredTransport(transport, tokenAuthorizationProvider)
	_, err = securedTransport.RoundTrip(request)
	require.EqualError(t, err, "error while obtaining authorization: error")
	require.Equal(t, request.URL, testUrl)
}
