package proxy_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"

	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy/automock"
	"github.com/stretchr/testify/require"
)

func TestAuditLog(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		//GIVEN
		graphqlResp := fixGraphqResponse()
		graphqlPayload, err := json.Marshal(&graphqlResp)
		require.NoError(t, err)

		claims := fixBearerHeader(t)
		req := httptest.NewRequest("POST", "http://localhost", bytes.NewBuffer(graphqlPayload))
		req.Header = http.Header{
			"Authorization": []string{claims},
		}
		resp := http.Response{
			StatusCode:    http.StatusCreated,
			Body:          ioutil.NopCloser(bytes.NewBuffer([]byte(graphqlPayload))),
			ContentLength: (int64)(len(graphqlPayload)),
		}

		roundTripper := &automock.RoundTrip{}
		roundTripper.On("RoundTrip", req).Return(&resp, nil).Once()

		auditlogSvc := &automock.AuditlogService{}
		auditlogSvc.On("Log", string(graphqlPayload), string(graphqlPayload), fixClaims()).Return(nil)

		transport := proxy.NewTransport(auditlogSvc, roundTripper)

		//WHEN
		output, err := transport.RoundTrip(req)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, output)
		roundTripper.AssertExpectations(t)
		auditlogSvc.AssertExpectations(t)
	})

	t.Run("Success HTTP GET", func(t *testing.T) {
		//GIVEN
		req := httptest.NewRequest("GET", "http://localhost", nil)

		resp := http.Response{
			StatusCode:    http.StatusCreated,
			Body:          ioutil.NopCloser(bytes.NewBuffer([]byte("response"))),
			ContentLength: (int64)(len("response")),
		}

		roundTripper := &automock.RoundTrip{}
		roundTripper.On("RoundTrip", req).Return(&resp, nil).Once()

		transport := proxy.NewTransport(nil, roundTripper)

		//WHEN
		_, err := transport.RoundTrip(req)

		//THEN
		require.NoError(t, err)
	})
}

func fixClaims() proxy.Claims {
	return proxy.Claims{
		Tenant:       "e36c520b-caa2-4677-b289-8a171184192b",
		Scopes:       "scopes",
		ConsumerID:   "134039be-840a-47f1-a962-d13410edf311",
		ConsumerType: "Application",
	}
}

func fixBearerHeader(t *testing.T) string {

	claims := fixClaims()

	marshalledClaims, err := json.Marshal(&claims)
	require.NoError(t, err)

	header := `{"alg": "HS256","typ": "JWT"}`

	tokenClaims := base64.RawURLEncoding.EncodeToString(marshalledClaims)
	tokenHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	return fmt.Sprintf("Bearer %s", fmt.Sprintf("%s.%s.", tokenHeader, tokenClaims))
}

func fixGraphqResponse() model.GraphqlResponse {
	return model.GraphqlResponse{
		Errors: nil,
		Data:   "payload",
	}
}
