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

func TestSecuredTransport_RoundTripSuccessfullyObtainsTokenFromCorrectTokenProviderAndUsesIt(t *testing.T) {
	const accessToken = "accessToken"

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
		authHeader := req.Header.Get("Authorization")
		require.Equal(t, "Bearer "+accessToken, authHeader)

		return nil, nil
	}

	tokenProvider := &httpfakes.FakeTokenProvider{}
	tokenProvider.MatchesReturns(true)
	tokenProvider.GetAuthorizationTokenReturns(httputil.Token{
		AccessToken: accessToken,
	}, nil)

	tokenProvider2 := &httpfakes.FakeTokenProvider{}
	tokenProvider2.MatchesReturns(false)
	tokenProvider2.GetAuthorizationTokenReturns(httputil.Token{
		AccessToken: accessToken + "2",
	}, nil)

	tokenProvider3 := &httpfakes.FakeTokenProvider{}
	tokenProvider3.MatchesReturns(true)
	tokenProvider3.GetAuthorizationTokenReturns(httputil.Token{
		AccessToken: accessToken + "3",
	}, nil)

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	securedTransport := httputil.NewSecuredTransport(transport, tokenProvider)
	_, err = securedTransport.RoundTrip(request)
	require.NoError(t, err)
}

func TestSecuredTransport_RoundTripCouldNotObtainTokeWhenNoTokenProviderMatches(t *testing.T) {
	const accessToken = "accessToken"

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
		authHeader := req.Header.Get("Authorization")
		require.Equal(t, "", authHeader)

		return nil, nil
	}

	tokenProvider := &httpfakes.FakeTokenProvider{}
	tokenProvider.MatchesReturns(false)
	tokenProvider.GetAuthorizationTokenReturns(httputil.Token{
		AccessToken: accessToken,
	}, nil)

	tokenProvider2 := &httpfakes.FakeTokenProvider{}
	tokenProvider2.MatchesReturns(false)
	tokenProvider2.GetAuthorizationTokenReturns(httputil.Token{
		AccessToken: accessToken + "2",
	}, nil)

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	securedTransport := httputil.NewSecuredTransport(transport, tokenProvider)
	_, err = securedTransport.RoundTrip(request)
	require.EqualError(t, err, "context did not match any token provider")
}

func TestSecuredTransport_RoundTripFailsOnTokenProviderError(t *testing.T) {
	transport := &httpfakes.FakeHTTPRoundTripper{}

	tokenProvider := &httpfakes.FakeTokenProvider{}
	tokenProvider.MatchesReturns(true)
	tokenProvider.GetAuthorizationTokenReturns(httputil.Token{}, errors.New("error"))

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	securedTransport := httputil.NewSecuredTransport(transport, tokenProvider)
	_, err = securedTransport.RoundTrip(request)
	require.EqualError(t, err, "error while obtaining token: error")
}
