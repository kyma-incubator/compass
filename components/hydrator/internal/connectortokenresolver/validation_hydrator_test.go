package connectortokenresolver_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	connector "github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/hydrator/internal/connectortokenresolver"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	clientID = "id"
	token    = "YWJj"
)

func TestValidationHydrator_ServeHTTP(t *testing.T) {
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

	t.Run("should fail when session cannot be decoded", func(t *testing.T) {
		// GIVEN
		directorClientMock := &automock.Client{}
		directorClientMock.AssertNotCalled(t, "GetSystemAuthByToken")

		validationHydrator := connectortokenresolver.NewValidationHydrator(directorClientMock)
		req := createAuthRequestWithTokenHeader(t, "", token)
		w := httptest.NewRecorder()
		// WHEN
		validationHydrator.ServeHTTP(w, req)
		// THEN
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "failed to decode Authentication Session from body")
		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("should resolve token from query params and add header to response", func(t *testing.T) {
		// GIVEN
		directorClientMock := &automock.Client{}
		directorClientMock.AssertNotCalled(t, "GetSystemAuthByToken")

		validationHydrator := connectortokenresolver.NewValidationHydrator(directorClientMock)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, "")
		w := httptest.NewRecorder()
		// WHEN
		validationHydrator.ServeHTTP(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err := json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("should resolve token from query params and add header to response", func(t *testing.T) {
		// GIVEN
		directorClientMock := &automock.Client{}
		directorClientMock.On("GetSystemAuthByToken", mock.Anything, token).Return(nil, errors.New("error"))

		validationHydrator := connectortokenresolver.NewValidationHydrator(directorClientMock)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()

		// WHEN
		validationHydrator.ServeHTTP(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err := json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("should fail when invalidating token fails", func(t *testing.T) {
		// GIVEN
		gqlAuth := &graphql.Auth{
			OneTimeToken: &graphql.OneTimeTokenForApplication{
				TokenWithURL: graphql.TokenWithURL{
					Type: "Runtime",
				},
			},
			CertCommonName: str.Ptr(""),
		}

		modelAuth, err := auth.ToModel(gqlAuth)
		require.NoError(t, err)

		sysAuth := &model.SystemAuth{
			ID:    clientID,
			Value: modelAuth,
		}

		directorClientMock := &automock.Client{}
		directorClientMock.On("GetSystemAuthByToken", mock.Anything, token).Return(sysAuth, nil).Once()
		directorClientMock.On("InvalidateSystemAuthOneTimeToken", mock.Anything, clientID).Return(errors.New("error when invalidating the token"))

		validationHydrator := connectortokenresolver.NewValidationHydrator(directorClientMock)
		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()

		// WHEN
		validationHydrator.ServeHTTP(w, req)
		// THEN
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("should succeed when token is resolved successfully", func(t *testing.T) {
		// GIVEN
		gqlAuth := &graphql.Auth{
			OneTimeToken: &graphql.OneTimeTokenForApplication{
				TokenWithURL: graphql.TokenWithURL{
					Type: "Runtime",
				},
			},
			CertCommonName: str.Ptr(""),
		}

		modelAuth, err := auth.ToModel(gqlAuth)
		require.NoError(t, err)

		sysAuth := &model.SystemAuth{
			ID:    clientID,
			Value: modelAuth,
		}

		directorClientMock := &automock.Client{}

		validationHydrator := connectortokenresolver.NewValidationHydrator(directorClientMock)
		directorClientMock.On("GetSystemAuthByToken", mock.Anything, token).Return(sysAuth, nil).Once()
		directorClientMock.On("InvalidateSystemAuthOneTimeToken", mock.Anything, clientID).Return(nil)

		authenticationSession := connector.AuthenticationSession{}
		req := createAuthRequestWithTokenQueryParam(t, authenticationSession, token)
		w := httptest.NewRecorder()

		// WHEN
		validationHydrator.ServeHTTP(w, req)
		// THEN
		assert.Equal(t, http.StatusOK, w.Code)

		var authSession connector.AuthenticationSession
		err = json.NewDecoder(w.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{clientID}, authSession.Header[connector.ClientIdFromTokenHeader])
	})
}

func emptyAuthSession() connector.AuthenticationSession {
	return connector.AuthenticationSession{}
}
