package open_resource_discovery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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
	testHttpClient := NewTestClient(func(req *http.Request) *http.Response {
		var data []byte
		var err error
		statusCode := http.StatusOK
		if strings.Contains(req.URL.String(), open_resource_discovery.WellKnownEndpoint) {
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
	})

	client := open_resource_discovery.NewClient(testHttpClient)
	docs, err := client.FetchOpenResourceDiscoveryDocuments(context.TODO(), baseURL)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, fixORDDocument(), *docs[0])
}
