package proxy_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createDefaultAdapterConfig() proxy.AdapterConfig {
	return proxy.AdapterConfig{
		MsgBodySizeLimit: 1000*1000 - 1024,
	}
}

func TestAdapterTransport(t *testing.T) {
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

		transport := proxy.NewAdapterTransport(nil, nil, roundTripper, createDefaultAdapterConfig())

		//WHEN
		_, err := transport.RoundTrip(req)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Succeeds on HTTP PUT request", func(t *testing.T) {
		//GIVEN
		gqlResp := fixGraphQLResponse()
		gqlPayload, err := json.Marshal(&gqlResp)
		require.NoError(t, err)

		token := fixBearerHeader(t)
		req := httptest.NewRequest("PUT", "http://localhost", bytes.NewBuffer(gqlPayload))
		req.Header = http.Header{
			"Authorization": []string{token},
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

		transport := proxy.NewAdapterTransport(postAuditlogSvc, preAuditlogSvc, roundTripper, createDefaultAdapterConfig())

		//WHEN
		output, err := transport.RoundTrip(req)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, output)
		roundTripper.AssertExpectations(t)
		preAuditlogSvc.AssertExpectations(t)
		postAuditlogSvc.AssertExpectations(t)
	})

	t.Run("Splits large request bodies to multiple shards", func(t *testing.T) {
		//GIVEN
		gqlResp := fixGraphQLResponseWithLength(900)
		gqlPayload, err := json.Marshal(&gqlResp)
		require.NoError(t, err)

		token := fixBearerHeader(t)
		req := httptest.NewRequest("PUT", "http://localhost", bytes.NewBuffer(gqlPayload))
		req.Header = http.Header{
			"Authorization": []string{token},
		}
		resp := http.Response{
			StatusCode:    http.StatusCreated,
			Body:          ioutil.NopCloser(bytes.NewBuffer(gqlPayload)),
			ContentLength: (int64)(len(gqlPayload)),
		}

		roundTripper := &automock.RoundTrip{}
		roundTripper.On("RoundTrip", req).Return(&resp, nil).Once()

		preAuditlogSvc := &automock.PreAuditlogService{}
		preAuditlogSvc.On("PreLog", mock.Anything, mock.MatchedBy(func(msg proxy.AuditlogMessage) bool { return msg.Claims == fixClaims() })).Return(nil).Twice()

		postAuditlogSvc := &automock.AuditlogService{}
		postAuditlogSvc.On("Log", mock.Anything, mock.MatchedBy(func(msg proxy.AuditlogMessage) bool { return msg.Claims == fixClaims() })).Return(nil).Twice()

		transport := proxy.NewAdapterTransport(postAuditlogSvc, preAuditlogSvc, roundTripper, proxy.AdapterConfig{
			MsgBodySizeLimit: (len(gqlPayload) + 1) / 2,
		})

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

		token := fixBearerHeader(t)
		req := httptest.NewRequest("PUT", "http://localhost", bytes.NewBuffer(gqlReqPayload))
		req.Header = http.Header{
			"Authorization": []string{token},
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

		transport := proxy.NewAdapterTransport(postAuditlogSvc, preAuditlogSvc, roundTripper, createDefaultAdapterConfig())

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

		token := fixBearerHeader(t)
		req := httptest.NewRequest("PUT", "http://localhost", bytes.NewBuffer(gqlReqPayload))
		req.Header = http.Header{
			"Authorization": []string{token},
		}

		roundTripper := &automock.RoundTrip{}
		postAuditlogSvc := &automock.AuditlogService{}

		preAuditlogSvc := &automock.PreAuditlogService{}
		preAuditlogSvc.On("PreLog", mock.Anything, mock.MatchedBy(func(msg proxy.AuditlogMessage) bool { return msg.Claims == fixClaims() })).Return(errors.New("auditlog issue"))

		transport := proxy.NewAdapterTransport(postAuditlogSvc, preAuditlogSvc, roundTripper, createDefaultAdapterConfig())

		//WHEN
		_, err = transport.RoundTrip(req)

		//THEN
		require.Error(t, err)
		preAuditlogSvc.AssertExpectations(t)
		postAuditlogSvc.AssertNotCalled(t, "Log")
		roundTripper.AssertNotCalled(t, "RoundTrip")
	})

	t.Run("Fails when missing Authorization token", func(t *testing.T) {
		//GIVEN
		gqlResp := fixGraphQLResponse()
		gqlPayload, err := json.Marshal(&gqlResp)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "http://localhost", bytes.NewBuffer(gqlPayload))
		req.Header = http.Header{
			"Authorization": []string{},
		}

		transport := proxy.NewAdapterTransport(nil, nil, nil, createDefaultAdapterConfig())

		//WHEN
		_, err = transport.RoundTrip(req)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "no bearer token")
	})

	t.Run("Fails when invalid token is provided", func(t *testing.T) {
		//GIVEN
		gqlResp := fixGraphQLResponse()
		gqlPayload, err := json.Marshal(&gqlResp)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "http://localhost", bytes.NewBuffer(gqlPayload))
		req.Header = http.Header{
			"Authorization": []string{"token"},
		}

		transport := proxy.NewAdapterTransport(nil, nil, nil, createDefaultAdapterConfig())

		//WHEN
		_, err = transport.RoundTrip(req)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "while parsing bearer token")
	})
}
