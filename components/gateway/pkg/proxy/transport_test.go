package proxy_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const ConsumerId = "134039be-840a-47f1-a962-d13410edf311"

func TestTransport(t *testing.T) {
	t.Run("Succeeds on HTTP GET request", func(t *testing.T) {
		//GIVEN
		req := httptest.NewRequest("GET", "http://localhost", nil)

		resp := http.Response{
			StatusCode:    http.StatusCreated,
			Body:          ioutil.NopCloser(bytes.NewBuffer([]byte("response"))),
			ContentLength: (int64)(len("response")),
		}

		roundTripper := &automock.RoundTrip{}
		roundTripper.On("RoundTrip", req).Return(&resp, nil).Once()

		transport := proxy.NewTransport(nil, nil, roundTripper)

		//WHEN
		_, err := transport.RoundTrip(req)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Succeeds on HTTP POST request", func(t *testing.T) {
		//GIVEN
		gqlResp := fixGraphQLResponse()
		gqlPayload, err := json.Marshal(&gqlResp)
		require.NoError(t, err)

		claims := fixBearerHeader(t)
		req := httptest.NewRequest("POST", "http://localhost", bytes.NewBuffer(gqlPayload))
		req.Header = http.Header{
			"Authorization": []string{claims},
		}
		resp := http.Response{
			StatusCode:    http.StatusCreated,
			Body:          ioutil.NopCloser(bytes.NewBuffer(gqlPayload)),
			ContentLength: (int64)(len(gqlPayload)),
		}

		roundTripper := &automock.RoundTrip{}
		roundTripper.On("RoundTrip", req).Return(&resp, nil).Once()

		preAuditlogSvc := &automock.PreAuditlogService{}
		preAuditlogSvc.On("PreLog", mock.Anything, mock.MatchedBy(func(msg proxy.AuditlogMessage) bool { return msg.Claims == fixClaims() })).Return(nil).Once()

		postAuditlogSvc := &automock.AuditlogService{}
		postAuditlogSvc.On("Log", mock.Anything, mock.MatchedBy(func(msg proxy.AuditlogMessage) bool { return msg.Claims == fixClaims() })).Return(nil).Once()

		transport := proxy.NewTransport(postAuditlogSvc, preAuditlogSvc, roundTripper)

		//WHEN
		output, err := transport.RoundTrip(req)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, output)
		roundTripper.AssertExpectations(t)
		preAuditlogSvc.AssertExpectations(t)
		postAuditlogSvc.AssertExpectations(t)
	})

	t.Run("Succeeds when post-audit log is not successful", func(t *testing.T) {
		//GIVEN
		gqlReq := fixGraphQLMutation()
		gqlReqPayload, err := json.Marshal(&gqlReq)
		require.NoError(t, err)

		gqlResp := fixGraphQLResponse()
		gqlRespPayload, err := json.Marshal(&gqlResp)
		require.NoError(t, err)

		claims := fixBearerHeader(t)
		req := httptest.NewRequest("POST", "http://localhost", bytes.NewBuffer(gqlReqPayload))
		req.Header = http.Header{
			"Authorization": []string{claims},
		}
		resp := http.Response{
			StatusCode:    http.StatusCreated,
			Body:          ioutil.NopCloser(bytes.NewBuffer(gqlRespPayload)),
			ContentLength: (int64)(len(gqlRespPayload)),
		}

		roundTripper := &automock.RoundTrip{}
		roundTripper.On("RoundTrip", req).Return(&resp, nil).Once()

		preAuditlogSvc := &automock.PreAuditlogService{}
		preAuditlogSvc.On("PreLog", mock.Anything, mock.MatchedBy(func(msg proxy.AuditlogMessage) bool { return msg.Claims == fixClaims() })).Return(nil).Once()

		postAuditlogSvc := &automock.AuditlogService{}
		postAuditlogSvc.On("Log", mock.Anything, mock.MatchedBy(func(msg proxy.AuditlogMessage) bool { return msg.Claims == fixClaims() })).Return(errors.New("auditlog issue")).Once()

		transport := proxy.NewTransport(postAuditlogSvc, preAuditlogSvc, roundTripper)

		//WHEN
		output, err := transport.RoundTrip(req)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, output)
		roundTripper.AssertExpectations(t)
		preAuditlogSvc.AssertExpectations(t)
		postAuditlogSvc.AssertExpectations(t)
	})

	t.Run("Fails when pre-audit log is not successful", func(t *testing.T) {
		//GIVEN
		gqlReq := fixGraphQLMutation()
		gqlReqPayload, err := json.Marshal(&gqlReq)
		require.NoError(t, err)

		claims := fixBearerHeader(t)
		req := httptest.NewRequest("POST", "http://localhost", bytes.NewBuffer(gqlReqPayload))
		req.Header = http.Header{
			"Authorization": []string{claims},
		}

		roundTripper := &automock.RoundTrip{}
		postAuditlogSvc := &automock.AuditlogService{}

		preAuditlogSvc := &automock.PreAuditlogService{}
		preAuditlogSvc.On("PreLog", mock.Anything, mock.MatchedBy(func(msg proxy.AuditlogMessage) bool { return msg.Claims == fixClaims() })).Return(errors.New("auditlog issue"))

		transport := proxy.NewTransport(postAuditlogSvc, preAuditlogSvc, roundTripper)

		//WHEN
		_, err = transport.RoundTrip(req)

		//THEN
		require.Error(t, err)
		preAuditlogSvc.AssertExpectations(t)
		postAuditlogSvc.AssertNotCalled(t, "Log")
		roundTripper.AssertNotCalled(t, "RoundTrip")
	})
}

func fixTokenClaims(t *testing.T) proxy.Claims {
	tenantJSON, err := json.Marshal(map[string]string{"consumerTenant": "e36c520b-caa2-4677-b289-8a171184192b", "externalTenant": "externalTenantName"})
	require.NoError(t, err)

	return proxy.Claims{
		Tenant:       string(tenantJSON),
		ConsumerID:   ConsumerId,
		ConsumerType: "Application",
		Scopes:       "scopes",
	}
}
func fixClaims() proxy.Claims {
	return proxy.Claims{
		Tenant:         "e36c520b-caa2-4677-b289-8a171184192b",
		ConsumerTenant: "e36c520b-caa2-4677-b289-8a171184192b",
		Scopes:         "scopes",
		ConsumerID:     ConsumerId,
		ConsumerType:   "Application",
	}
}

func fixBearerHeaderWithTokenClaims(t *testing.T, claims interface{}) string {
	marshalledClaims, err := json.Marshal(&claims)
	require.NoError(t, err)

	header := `{"alg": "HS256","typ": "JWT"}`

	tokenClaims := base64.RawURLEncoding.EncodeToString(marshalledClaims)
	tokenHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	return fmt.Sprintf("Bearer %s", fmt.Sprintf("%s.%s.", tokenHeader, tokenClaims))
}

func fixBearerHeader(t *testing.T) string {
	claims := fixTokenClaims(t)

	return fixBearerHeaderWithTokenClaims(t, claims)
}

func fixGraphQLResponse() model.GraphqlResponse {
	return model.GraphqlResponse{
		Errors: nil,
		Data:   "payload",
	}
}

func fixGraphQLResponseWithLength(length int) model.GraphqlResponse {
	return model.GraphqlResponse{
		Errors: nil,
		Data:   strings.Repeat("a", length),
	}
}

func fixGraphQLMutation() map[string]interface{} {
	gqlBody := make(map[string]interface{}, 0)
	gqlBody["mutation"] = "some-mutation"
	return gqlBody
}
