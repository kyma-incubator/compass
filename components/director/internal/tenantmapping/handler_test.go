package tenantmapping_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	persistencemock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	target := "http://example.com/foo"
	externalTenantID := "external-" + uuid.New().String()
	tenantID := uuid.New()
	systemAuthID := uuid.New()
	objID := uuid.New()

	t.Run("success for the request parsed as JWT flow", func(t *testing.T) {
		username := "admin"
		scopes := "application:read"
		reqDataMock := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.UsernameKey: username,
				},
			},
		}
		objCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   username,
			ConsumerType: "Static User",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","externalTenant":"` + externalTenantID + `","name":"` + username + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		transactMock := getTransactMock()

		mapperForUserMock := getMapperForUserMock()
		mapperForUserMock.On("GetObjectContext", mock.Anything, reqDataMock, username).Return(objCtxMock, nil).Once()

		handler := tenantmapping.NewHandler(reqDataParserMock, transactMock, mapperForUserMock, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, transactMock, mapperForUserMock)
	})

	t.Run("success for the request parsed as OAuth2 flow", func(t *testing.T) {
		scopes := "application:read"
		reqDataMock := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ClientIDKey: systemAuthID.String(),
				},
			},
		}
		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"client_id":"` + systemAuthID.String() + `","consumerID":"` + objID.String() + `","consumerType":"Integration System","externalTenant":"` + externalTenantID + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		transactMock := getTransactMock()

		mapperForSystemAuthMock := getMapperForSystemAuthMock()
		mapperForSystemAuthMock.On("GetObjectContext", mock.Anything, reqDataMock, systemAuthID.String(), tenantmapping.OAuth2Flow).Return(objCtx, nil).Once()

		handler := tenantmapping.NewHandler(reqDataParserMock, transactMock, nil, mapperForSystemAuthMock)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, transactMock, mapperForSystemAuthMock)
	})

	t.Run("success for the request parsed as Certificate flow", func(t *testing.T) {
		scopes := "application:read"
		reqDataMock := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(tenantmapping.ClientIDCertKey): []string{systemAuthID.String()},
				},
			},
		}
		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","externalTenant":"` + externalTenantID + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":{"Client-Id-From-Certificate":["` + systemAuthID.String() + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		transactMock := getTransactMock()

		mapperForSystemAuthMock := getMapperForSystemAuthMock()
		mapperForSystemAuthMock.On("GetObjectContext", mock.Anything, reqDataMock, systemAuthID.String(), tenantmapping.CertificateFlow).Return(objCtx, nil).Once()

		handler := tenantmapping.NewHandler(reqDataParserMock, transactMock, nil, mapperForSystemAuthMock)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, transactMock, mapperForSystemAuthMock)
	})

	t.Run("success for the request parsed as OneTimeToken flow", func(t *testing.T) {
		scopes := "application:read"
		reqDataMock := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(tenantmapping.ClientIDTokenKey): []string{systemAuthID.String()},
				},
			},
		}
		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","externalTenant":"` + externalTenantID + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":{"Client-Id-From-Token":["` + systemAuthID.String() + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		transactMock := getTransactMock()

		mapperForSystemAuthMock := getMapperForSystemAuthMock()
		mapperForSystemAuthMock.On("GetObjectContext", mock.Anything, reqDataMock, systemAuthID.String(), tenantmapping.OneTimeTokenFlow).Return(objCtx, nil).Once()

		handler := tenantmapping.NewHandler(reqDataParserMock, transactMock, nil, mapperForSystemAuthMock)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, transactMock, mapperForSystemAuthMock)
	})

	t.Run("error when sending different HTTP verb than POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		handler := tenantmapping.NewHandler(nil, nil, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, "Bad request method. Got GET, expected POST", strings.TrimSpace(string(body)))
	})

	t.Run("error when body parser returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(tenantmapping.ReqData{}, errors.New("some error")).Once()

		logger := getLoggerMock(t, "while parsing the request: some error")

		handler := tenantmapping.NewHandler(reqDataParserMock, nil, nil, nil)
		handler.SetLogger(logger)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := tenantmapping.ReqData{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, tenantmapping.ReqData{}, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, logger)
	})

	t.Run("error when transaction begin fails", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.UsernameKey: "test",
				},
			},
		}

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader("{}"))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqData, nil).Once()

		transactMock := &persistencemock.Transactioner{}
		transactMock.On("Begin").Return(nil, errors.New("some error")).Once()

		logger := getLoggerMock(t, "while opening the db transaction: some error")

		handler := tenantmapping.NewHandler(reqDataParserMock, transactMock, nil, nil)
		handler.SetLogger(logger)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := tenantmapping.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, reqData.Body, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, transactMock, logger)
	})

	t.Run("error when GetAuthID returns error when looking for Auth ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(tenantmapping.ReqData{}, nil).Once()

		transactMock := getTransactMock()

		logger := getLoggerMock(t, "while getting object context: while determining the auth ID from the request: unable to find valid auth ID")

		handler := tenantmapping.NewHandler(reqDataParserMock, transactMock, nil, nil)
		handler.SetLogger(logger)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := tenantmapping.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, tenantmapping.ReqBody{}, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, transactMock, logger)
	})
}

func getMapperForUserMock() *automock.ObjectContextForUserProvider {
	provider := &automock.ObjectContextForUserProvider{}
	return provider
}

func getMapperForSystemAuthMock() *automock.ObjectContextForSystemAuthProvider {
	provider := &automock.ObjectContextForSystemAuthProvider{}
	return provider
}

func getTransactMock() *persistencemock.Transactioner {
	persistTx := &persistencemock.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()

	transact := &persistencemock.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("RollbackUnlessCommited", persistTx).Return().Once()

	return transact
}

func getLoggerMock(t *testing.T, expectedMessage string) *automock.Logger {
	logger := &automock.Logger{}
	logger.On("Error", mock.MatchedBy(func(err error) bool {
		require.Error(t, err)
		require.Equal(t, expectedMessage, err.Error())
		return true
	})).Once()
	return logger
}
