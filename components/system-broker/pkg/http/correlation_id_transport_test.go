package http_test

import (
	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/http/httpfakes"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/uid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func TestCorrelationIDTransport_RoundTripSetsCorrelationIDIfSuchIsPresent(t *testing.T) {
	const correlationID = "test"

	transport := &httpfakes.FakeHTTPRoundTripper{}
	transport.RoundTripStub = func(req *http.Request) (*http.Response, error) {
			entry := log.C(req.Context())

			correlationIDFromLogger, exists := entry.Data[log.FieldCorrelationID]
			require.True(t, exists)
			require.Equal(t, correlationID, correlationIDFromLogger)

			return nil, nil
	}

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodPost,
		URL:    testUrl,
		Header: map[string][]string{},
	}
	request.Header.Set("X-Correlation-ID", correlationID)

	correlationTransport := httputil.NewCorrelationIDTransport(transport, uid.NewService())
	_, err = correlationTransport.RoundTrip(request)
	require.NoError(t, err)
}
