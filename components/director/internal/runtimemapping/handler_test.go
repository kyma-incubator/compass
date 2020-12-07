package runtimemapping

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"

	"github.com/form3tech-oss/jwt-go"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/runtimemapping/automock"
)

func TestHandler(t *testing.T) {
	txCtxGenerator := txtest.NewTransactionContextGenerator(errors.New("some-error"))

	authorizationHeader := fmt.Sprintf("Bearer token-value")
	issuer := "https://dex.domain.local"
	claimsMock := &jwt.MapClaims{"iss": issuer, "email": "me@domain.local", "groups": "admin-group", "name": "John Doe"}
	tenantID := uuid.New().String()

	t.Run("success for the request with valid token", func(t *testing.T) {
		// GIVEN
		extTenantID := uuid.New().String()
		expectedBody := "{\"subject\":\"\",\"extra\":{\"email\":\"me@domain.local\",\"groups\":\"admin-group\",\"name\":\"John Doe\"},\"header\":{\"Tenant\":[\"" + extTenantID + "\"]}}"
		reqDataParser := oathkeeper.NewReqDataParser()
		persistenceMock, transactMock := txCtxGenerator.ThatSucceeds()

		runtimeSvcMock := &automock.RuntimeService{}
		runtimeSvcMock.On("GetByTokenIssuer", mock.Anything, issuer).Return(&model.Runtime{Tenant: tenantID}, nil)

		tokenVerifierMock := &automock.TokenVerifier{}
		tokenVerifierMock.On("Verify", mock.Anything, authorizationHeader).Return(claimsMock, nil)

		tenantSvcMock := &automock.TenantService{}
		tenantSvcMock.On("GetExternalTenant", mock.Anything, tenantID).Return(extTenantID, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://domain.local", strings.NewReader("{}"))
		req.Header.Add("Authorization", authorizationHeader)

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
		req = req.WithContext(ctx)

		// WHEN
		handler := NewHandler(reqDataParser, transactMock, tokenVerifierMock, runtimeSvcMock, tenantSvcMock)
		handler.ServeHTTP(w, req)
		resp := w.Result()

		// THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, expectedBody, strings.TrimSpace(string(body)))
		require.Equal(t, 0, len(hook.Entries))

		mock.AssertExpectationsForObjects(t, transactMock, persistenceMock, tokenVerifierMock, runtimeSvcMock, tenantSvcMock)
	})

	t.Run("when sending different HTTP verb than POST", func(t *testing.T) {
		// GIVEN
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://domain.local", strings.NewReader(""))

		// WHEN
		handler := NewHandler(nil, nil, nil, nil, nil)
		handler.ServeHTTP(w, req)
		resp := w.Result()

		// THEN
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "Bad request method. Got GET, expected POST", strings.TrimSpace(string(body)))
	})

	t.Run("when unable to parse the request body should log the error", func(t *testing.T) {
		// GIVEN
		expectedBody := "{\"subject\":\"\",\"extra\":null,\"header\":null}"
		reqDataParser := oathkeeper.NewReqDataParser()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://domain.local", strings.NewReader(""))

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
		req = req.WithContext(ctx)

		// WHEN
		handler := NewHandler(reqDataParser, nil, nil, nil, nil)
		handler.ServeHTTP(w, req)
		resp := w.Result()

		// THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, expectedBody, strings.TrimSpace(string(body)))
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, "An error has occurred while parsing the request.", hook.LastEntry().Message)
		errMsg, ok := hook.LastEntry().Data["error"].(error)
		assert.True(t, ok)
		require.Equal(t, "Internal Server Error: request body is empty", errMsg.Error())
	})

	t.Run("when token verifier returns error", func(t *testing.T) {
		// GIVEN
		expectedBody := "{\"subject\":\"\",\"extra\":{},\"header\":{}}"
		reqDataParser := oathkeeper.NewReqDataParser()
		persistenceMock, transactMock := txCtxGenerator.ThatDoesntExpectCommit()

		tokenVerifierMock := &automock.TokenVerifier{}
		tokenVerifierMock.On("Verify", mock.Anything, authorizationHeader).Return(nil, errors.New("some-error"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://domain.local", strings.NewReader("{}"))
		req.Header.Add("Authorization", authorizationHeader)

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
		req = req.WithContext(ctx)

		// WHEN
		handler := NewHandler(reqDataParser, transactMock, tokenVerifierMock, nil, nil)
		handler.ServeHTTP(w, req)
		resp := w.Result()

		// THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, expectedBody, strings.TrimSpace(string(body)))
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, "An error has occurred while processing the request.", hook.LastEntry().Message)
		errMsg, ok := hook.LastEntry().Data["error"].(error)
		assert.True(t, ok)
		require.Equal(t, "while verifying the token: some-error", errMsg.Error())
		mock.AssertExpectationsForObjects(t, transactMock, persistenceMock, tokenVerifierMock)
	})

	t.Run("when claims have no issuer", func(t *testing.T) {
		// GIVEN
		claimsMock := &jwt.MapClaims{"email": "me@domain.local", "groups": "admin-group", "name": "John Doe"}
		expectedBody := "{\"subject\":\"\",\"extra\":{},\"header\":{}}"
		reqDataParser := oathkeeper.NewReqDataParser()
		persistenceMock, transactMock := txCtxGenerator.ThatDoesntExpectCommit()

		tokenVerifierMock := &automock.TokenVerifier{}
		tokenVerifierMock.On("Verify", mock.Anything, authorizationHeader).Return(claimsMock, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://domain.local", strings.NewReader("{}"))
		req.Header.Add("Authorization", authorizationHeader)

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
		req = req.WithContext(ctx)

		// WHEN
		handler := NewHandler(reqDataParser, transactMock, tokenVerifierMock, nil, nil)
		handler.ServeHTTP(w, req)
		resp := w.Result()

		// THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, expectedBody, strings.TrimSpace(string(body)))
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, "An error has occurred while processing the request.", hook.LastEntry().Message)
		errMsg, ok := hook.LastEntry().Data["error"].(error)
		assert.True(t, ok)
		require.Equal(t, "unable to get the issuer: Internal Server Error: no issuer claim found", errMsg.Error())
		mock.AssertExpectationsForObjects(t, transactMock, persistenceMock, tokenVerifierMock)
	})

	t.Run("when runtime verifier returns error", func(t *testing.T) {
		// GIVEN
		expectedBody := "{\"subject\":\"\",\"extra\":{},\"header\":{}}"

		reqDataParser := oathkeeper.NewReqDataParser()
		persistenceMock, transactMock := txCtxGenerator.ThatDoesntExpectCommit()

		runtimeSvcMock := &automock.RuntimeService{}
		runtimeSvcMock.On("GetByTokenIssuer", mock.Anything, issuer).Return(nil, errors.New("some-error"))

		tokenVerifierMock := &automock.TokenVerifier{}
		tokenVerifierMock.On("Verify", mock.Anything, authorizationHeader).Return(claimsMock, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://domain.local", strings.NewReader("{}"))
		req.Header.Add("Authorization", authorizationHeader)

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
		req = req.WithContext(ctx)

		// WHEN
		handler := NewHandler(reqDataParser, transactMock, tokenVerifierMock, runtimeSvcMock, nil)
		handler.ServeHTTP(w, req)
		resp := w.Result()

		// THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, expectedBody, strings.TrimSpace(string(body)))
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, "An error has occurred while processing the request.", hook.LastEntry().Message)
		errMsg, ok := hook.LastEntry().Data["error"].(error)
		assert.True(t, ok)
		require.Equal(t, "when getting the runtime: some-error", errMsg.Error())
		mock.AssertExpectationsForObjects(t, transactMock, persistenceMock, tokenVerifierMock, runtimeSvcMock)
	})

	t.Run("when mapping to external tenant returns error", func(t *testing.T) {
		// GIVEN
		expectedBody := "{\"subject\":\"\",\"extra\":{},\"header\":{}}"
		reqDataParser := oathkeeper.NewReqDataParser()
		persistenceMock, transactMock := txCtxGenerator.ThatDoesntExpectCommit()

		runtimeSvcMock := &automock.RuntimeService{}
		runtimeSvcMock.On("GetByTokenIssuer", mock.Anything, issuer).Return(&model.Runtime{Tenant: tenantID}, nil)

		tokenVerifierMock := &automock.TokenVerifier{}
		tokenVerifierMock.On("Verify", mock.Anything, authorizationHeader).Return(claimsMock, nil)

		tenantSvcMock := &automock.TenantService{}
		tenantSvcMock.On("GetExternalTenant", mock.Anything, tenantID).Return("", errors.New("some-error"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://domain.local", strings.NewReader("{}"))
		req.Header.Add("Authorization", authorizationHeader)

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
		req = req.WithContext(ctx)

		// WHEN
		handler := NewHandler(reqDataParser, transactMock, tokenVerifierMock, runtimeSvcMock, tenantSvcMock)
		handler.ServeHTTP(w, req)
		resp := w.Result()

		// THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, expectedBody, strings.TrimSpace(string(body)))
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, "An error has occurred while processing the request.", hook.LastEntry().Message)
		errMsg, ok := hook.LastEntry().Data["error"].(error)
		assert.True(t, ok)
		require.Equal(t, "unable to fetch external tenant based on runtime tenant: some-error", errMsg.Error())
		mock.AssertExpectationsForObjects(t, transactMock, persistenceMock, tokenVerifierMock, runtimeSvcMock, tenantSvcMock)
	})
}
