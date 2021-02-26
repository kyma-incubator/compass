package oathkeeper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	connector "github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	mocks "github.com/kyma-incubator/compass/components/director/internal/oathkeeper/automock"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	persistenceMocks "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	timeMock "github.com/kyma-incubator/compass/components/director/pkg/time/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
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
		tokenSerivce := &mocks.Service{}
		transact := &persistenceMocks.Transactioner{}
		timeService := &timeMock.Service{}

		validationHydrator := NewValidationHydrator(tokenSerivce, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		req := createAuthRequestWithTokenHeader(t, "", token)
		w := httptest.NewRecorder()
		transact.On("Begin").Return(nil, errors.New("err"))
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, string(w.Body.Bytes()), "unexpected error occured while resolving one time token")
	})

	t.Run("should fail when session cannot be decoded", func(t *testing.T) {
		tokenSerivce := &mocks.Service{}
		transact := &persistenceMocks.Transactioner{}
		timeService := &timeMock.Service{}

		validationHydrator := NewValidationHydrator(tokenSerivce, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		req := createAuthRequestWithTokenHeader(t, "", token)
		w := httptest.NewRecorder()
		per := &persistenceMocks.PersistenceTx{}
		transact.On("Begin").Return(per, nil)
		transact.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, string(w.Body.Bytes()), "failed to decode Authentication Session from body")
	})

	t.Run("should resolve token from query params and add header to response", func(t *testing.T) {
		tokenSerivce := &mocks.Service{}
		transact := &persistenceMocks.Transactioner{}
		timeService := &timeMock.Service{}

		validationHydrator := NewValidationHydrator(tokenSerivce, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		authenticationSession := connector.AuthenticationSession{
			Subject: "",
			Extra:   nil,
			Header:  nil,
		}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, "")
		w := httptest.NewRecorder()
		per := &persistenceMocks.PersistenceTx{}
		transact.On("Begin").Return(per, nil)
		transact.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should resolve token from query params and add header to response", func(t *testing.T) {
		tokenSerivce := &mocks.Service{}
		transact := &persistenceMocks.Transactioner{}
		timeService := &timeMock.Service{}

		validationHydrator := NewValidationHydrator(tokenSerivce, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		authenticationSession := connector.AuthenticationSession{
			Subject: "",
			Extra:   nil,
			Header:  nil,
		}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		per := &persistenceMocks.PersistenceTx{}
		transact.On("Begin").Return(per, nil)
		transact.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		tokenSerivce.On("GetByToken", mock.Anything, token).Return(nil, errors.New("error"))
		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should fail for expired token", func(t *testing.T) {
		tokenSerivce := &mocks.Service{}
		transact := &persistenceMocks.Transactioner{}
		timeService := &timeMock.Service{}
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
		validationHydrator := NewValidationHydrator(tokenSerivce, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		authenticationSession := connector.AuthenticationSession{
			Subject: "",
			Extra:   nil,
			Header:  nil,
		}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		per := &persistenceMocks.PersistenceTx{}
		transact.On("Begin").Return(per, nil)
		transact.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		tokenSerivce.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		timeService.On("Now").Return(afterOneDay)

		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should fail when invalidating token fails", func(t *testing.T) {
		tokenSerivce := &mocks.Service{}
		transact := &persistenceMocks.Transactioner{}
		timeService := &timeMock.Service{}
		now := time.Now()
		beforeOneDay := now.AddDate(0, 0, -1)
		systemAuth := &model.SystemAuth{
			ID: "id",
			Value: &model.Auth{
				OneTimeToken: &model.OneTimeToken{
					CreatedAt: time.Now(),
					Type:      tokens.RuntimeToken,
				},
			},
		}
		validationHydrator := NewValidationHydrator(tokenSerivce, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		authenticationSession := connector.AuthenticationSession{
			Subject: "",
			Extra:   nil,
			Header:  nil,
		}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		per := &persistenceMocks.PersistenceTx{}
		transact.On("Begin").Return(per, nil)
		transact.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		tokenSerivce.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		tokenSerivce.On("InvalidateToken", mock.Anything, mock.Anything).Return(errors.New("error when invalidating the token"))
		timeService.On("Now").Return(beforeOneDay)

		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should fail when db transaction commit fails", func(t *testing.T) {
		tokenSerivce := &mocks.Service{}
		transact := &persistenceMocks.Transactioner{}
		timeService := &timeMock.Service{}
		now := time.Now()
		beforeOneDay := now.AddDate(0, 0, -1)
		systemAuth := &model.SystemAuth{
			ID: "id",
			Value: &model.Auth{
				OneTimeToken: &model.OneTimeToken{
					CreatedAt: time.Now(),
					Type:      tokens.CSRToken,
				},
			},
		}
		validationHydrator := NewValidationHydrator(tokenSerivce, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		authenticationSession := connector.AuthenticationSession{
			Subject: "",
			Extra:   nil,
			Header:  nil,
		}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		per := &persistenceMocks.PersistenceTx{}
		transact.On("Begin").Return(per, nil)
		transact.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		tokenSerivce.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		tokenSerivce.On("InvalidateToken", mock.Anything, mock.Anything).Return(nil)
		per.On("Commit").Return(errors.New("error during transaction commit"))
		timeService.On("Now").Return(beforeOneDay)

		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should succeed when token is resolved successfully", func(t *testing.T) {
		tokenSerivce := &mocks.Service{}
		transact := &persistenceMocks.Transactioner{}
		timeService := &timeMock.Service{}
		now := time.Now()
		beforeOneDay := now.AddDate(0, 0, -1)
		systemAuth := &model.SystemAuth{
			ID: "id",
			Value: &model.Auth{
				OneTimeToken: &model.OneTimeToken{
					CreatedAt: time.Now(),
					Type:      tokens.CSRToken,
				},
			},
		}
		validationHydrator := NewValidationHydrator(tokenSerivce, transact, timeService, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration)

		authenticationSession := connector.AuthenticationSession{
			Subject: "",
			Extra:   nil,
			Header:  nil,
		}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()
		per := &persistenceMocks.PersistenceTx{}
		transact.On("Begin").Return(per, nil)
		transact.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		tokenSerivce.On("GetByToken", mock.Anything, token).Return(systemAuth, nil)
		tokenSerivce.On("InvalidateToken", mock.Anything, mock.Anything).Return(nil)
		per.On("Commit").Return(nil)
		timeService.On("Now").Return(beforeOneDay)

		validationHydrator.ResolveConnectorTokenHeader(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
