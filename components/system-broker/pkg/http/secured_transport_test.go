package http_test

import (
	"errors"
	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/http/httpfakes"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestSecuredTransport_RoundTripSuccessfullyObtainsTokenAndUsesItUntilExpire(t *testing.T) {
	const accessToken = "accessToken"

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
			authHeader := req.Header.Get("Authorization")
			require.Equal(t, "Bearer " + accessToken, authHeader)

			return nil, nil
	}

	tokenProvider := &httpfakes.FakeTokenProvider{}
	tokenProvider.GetAuthorizationTokenReturns(httputil.Token{
		AccessToken: accessToken,
		Expiration: time.Now().Add(time.Second * 10).Unix(),
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

	_, err = securedTransport.RoundTrip(request)
	require.NoError(t, err)
}

func TestSecuredTransport_RoundTripSuccessfullyObtainsNewTokenAfterExpiration(t *testing.T) {
	const accessToken1 = "accessToken1"
	const accessToken2 = "accessToken2"

	tokenProvider := &httpfakes.FakeTokenProvider{}
	tokenProvider.GetAuthorizationTokenReturnsOnCall(0, httputil.Token{
		AccessToken: accessToken1,
		Expiration: time.Now().Add(time.Millisecond * 100).Unix(),
	}, nil)
	tokenProvider.GetAuthorizationTokenReturnsOnCall(1, httputil.Token{
		AccessToken: accessToken2,
		Expiration: time.Now().Add(time.Second * 10).Unix(),
	}, nil)

	transport := &httpfakes.FakeHTTPRoundTripper{}

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
		authHeader := req.Header.Get("Authorization")
		require.Equal(t, "Bearer " + accessToken1, authHeader)

		return nil, nil
	}

	securedTransport := httputil.NewSecuredTransport(transport, tokenProvider)
	_, err = securedTransport.RoundTrip(request)
	require.NoError(t, err)

	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
		authHeader := req.Header.Get("Authorization")
		require.Equal(t, "Bearer " + accessToken2, authHeader)

		return nil, nil
	}

	_, err = securedTransport.RoundTrip(request)
	require.NoError(t, err)
}

func TestSecuredTransport_RoundTripFailsOnTokenProviderError(t *testing.T) {
	transport := &httpfakes.FakeHTTPRoundTripper{}

	tokenProvider := &httpfakes.FakeTokenProvider{}
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
	require.Error(t, err)
}
