package oathkeeper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	connector "github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	mocks "github.com/kyma-incubator/compass/components/director/internal/oathkeeper/automock"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	timeMock "github.com/kyma-incubator/compass/components/director/pkg/time/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	clientID               = "id"
	token                  = "tokenValue"
	csrTokenExpiration     = time.Duration(100)
	appTokenExpiration     = time.Duration(100)
	runtimeTokenExpiration = time.Duration(100)
)

func TestValidationHydrator_ResolveConnectorTokenHeader(t *testing.T) {

	createAuthRequestWithTokenHeader := func(t *testing.T, session string, tokenValue string) *http.Request {

		marshalledSession, err := json.Marshal(session)
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		req.Header.Add(connector.ConnectorTokenHeader, tokenValue)
		return req
	}

	createAuthRequestWithTokenQueryParam := func(t *testing.T, session interface{}, tokenValue string) *http.Request {
		marshalledSession, err := json.Marshal(session)
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, "?token="+tokenValue, bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		return req
	}

	t.Run("should fail when db transaction open fails", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatFailsOnBegin()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		// WHEN
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		req := createAuthRequestWithTokenHeader(t, "", token)
		w := httptest.NewRecorder()
		// THEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "unexpected error occurred while resolving one time token")
	})

	t.Run("should fail when session cannot be decoded", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		req := createAuthRequestWithTokenHeader(t, "", token)
		w := httptest.NewRecorder()
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		// THEN
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "failed to decode Authentication Session from body")
	})

	t.Run("should resolve token from query params and add header to response", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, "")
		w := httptest.NewRecorder()
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err := json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
	})

	t.Run("should resolve token from query params and add header to response", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		tokenService.On("GetByToken", mock.Anything, token).Return(nil, errors.New("error"))
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err := json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
	})

	t.Run("should fail when can't get OneTimeToken from systemAuth", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		afterOneDay := time.Now().AddDate(0, 0, +1)
		systemAuth := &model.SystemAuth{
			Value: nil,
		}
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		tokenService.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		timeService.On("Now").Return(afterOneDay)
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err := json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
	})

	t.Run("should fail when can't compute expiration time", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		afterOneDay := time.Now().AddDate(0, 0, +1)
		systemAuth := &model.SystemAuth{
			Value: &model.Auth{
				OneTimeToken: &model.OneTimeToken{
					Type: "InvalidType",
				},
			},
		}
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		tokenService.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		timeService.On("Now").Return(afterOneDay)
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err := json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
	})

	t.Run("should fail when the token is expired", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		now := time.Now()
		afterOneDay := now.AddDate(0, 0, +1)
		systemAuth := &model.SystemAuth{
			Value: &model.Auth{
				OneTimeToken: &model.OneTimeToken{
					CreatedAt: time.Now(),
					Type:      tokens.ApplicationToken,
				},
			},
		}
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		tokenService.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		timeService.On("Now").Return(afterOneDay)
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err := json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
	})

	t.Run("should fail when invalidating token fails", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		beforeOneDay := time.Now().AddDate(0, 0, -1)
		systemAuth := &model.SystemAuth{
			ID: clientID,
			Value: &model.Auth{
				OneTimeToken: &model.OneTimeToken{
					CreatedAt: time.Now(),
					Type:      tokens.RuntimeToken,
				},
			},
		}
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		tokenService.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		tokenService.On("InvalidateToken", mock.Anything, mock.Anything).Return(errors.New("error when invalidating the token"))
		timeService.On("Now").Return(beforeOneDay)
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		//THEN
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should fail when db transaction commit fails", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		beforeOneDay := time.Now().AddDate(0, 0, -1)
		systemAuth := &model.SystemAuth{
			ID: clientID,
			Value: &model.Auth{
				OneTimeToken: &model.OneTimeToken{
					CreatedAt: time.Now(),
					Type:      tokens.CSRToken,
				},
			},
		}
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		tokenService.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		tokenService.On("InvalidateToken", mock.Anything, mock.Anything).Return(nil)
		timeService.On("Now").Return(beforeOneDay)
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		// THEN
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should succeed when token is resolved successfully", func(t *testing.T) {
		// GIVEN
		tokenService := &mocks.Service{}
		timeService := &timeMock.Service{}
		mockedTx, transact := txtest.NewTransactionContextGenerator(errors.New("err")).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer transact.AssertExpectations(t)
		beforeOneDay := time.Now().AddDate(0, 0, -1)
		systemAuth := &model.SystemAuth{
			ID: clientID,
			Value: &model.Auth{
				OneTimeToken: &model.OneTimeToken{
					CreatedAt: time.Now(),
					Type:      tokens.CSRToken,
				},
			},
		}
		validationHydrator := NewValidationHydrator(tokenService, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		tokenService.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		tokenService.On("InvalidateToken", mock.Anything, mock.Anything).Return(nil)
		timeService.On("Now").Return(beforeOneDay)
		// WHEN
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err := json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{clientID}, authSession.Header[connector.ClientIdFromTokenHeader])
	})
}

func emptyAuthSession() connector.AuthenticationSession {
	return connector.AuthenticationSession{}
}
