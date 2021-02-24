package oathkeeper

//
//import (
//	"bytes"
//	"encoding/json"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
//	"github.com/kyma-project/cli/pkg/installation/mocks"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//	"github.com/stretchr/testify/require"
//)
//
//func TestValidationHydrator_ResolveConnectorTokenHeader(t *testing.T) {
//	marshalledSession, err := json.Marshal(emptyAuthSession())
//	require.NoError(t, err)
//
//	createAuthRequestWithTokenHeader := func(t *testing.T) *http.Request {
//		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
//		require.NoError(t, err)
//		req.Header.Add(ConnectorTokenHeader, token)
//		return req
//	}
//
//	createAuthRequestWithTokenQueryParam := func(t *testing.T) *http.Request {
//		req, err := http.NewRequest(http.MethodPost, "?token="+token, bytes.NewBuffer(marshalledSession))
//		require.NoError(t, err)
//		return req
//	}
//
//	t.Run("should resolve token from header and add header to response", func(t *testing.T) {
//		// given
//		req := createAuthRequestWithTokenHeader(t)
//		rr := httptest.NewRecorder()
//
//		tokenService := &mocks.Service{}
//		tokenService.On("Resolve", token).Return(tokenData, nil)
//		tokenService.On("Delete", token).Return(nil)
//
//		validator := NewValidationHydrator(tokenService, nil, nil)
//
//		// when
//		validator.ResolveConnectorTokenHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		var authSession AuthenticationSession
//		err = json.NewDecoder(rr.Body).Decode(&authSession)
//		require.NoError(t, err)
//
//		assert.Equal(t, []string{clientId}, authSession.Header[ClientIdFromTokenHeader])
//		mock.AssertExpectationsForObjects(t, tokenService)
//	})
//
//	t.Run("should resolve token from query params and add header to response", func(t *testing.T) {
//		// given
//		req := createAuthRequestWithTokenQueryParam(t)
//		rr := httptest.NewRecorder()
//
//		tokenService := &mocks.Service{}
//		tokenService.On("Resolve", token).Return(tokenData, nil)
//		tokenService.On("Delete", token).Return(nil)
//
//		validator := NewValidationHydrator(tokenService, nil, nil)
//
//		// when
//		validator.ResolveConnectorTokenHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		var authSession AuthenticationSession
//		err = json.NewDecoder(rr.Body).Decode(&authSession)
//		require.NoError(t, err)
//
//		assert.Equal(t, []string{clientId}, authSession.Header[ClientIdFromTokenHeader])
//		mock.AssertExpectationsForObjects(t, tokenService)
//	})
//
//	t.Run("should not modify authentication session if failed to resolved token", func(t *testing.T) {
//		// given
//		req := createAuthRequestWithTokenHeader(t)
//		rr := httptest.NewRecorder()
//
//		tokenService := &mocks.Service{}
//		tokenService.On("Resolve", token).Return(tokens.TokenData{}, apperrors.NotFound("error"))
//
//		validator := NewValidationHydrator(tokenService, nil, nil)
//
//		// when
//		validator.ResolveConnectorTokenHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		var authSession AuthenticationSession
//		err = json.NewDecoder(rr.Body).Decode(&authSession)
//		require.NoError(t, err)
//
//		assert.Equal(t, emptyAuthSession(), authSession)
//		mock.AssertExpectationsForObjects(t, tokenService)
//	})
//
//	t.Run("should not modify authentication session if no token provided", func(t *testing.T) {
//		// given
//		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
//		require.NoError(t, err)
//		rr := httptest.NewRecorder()
//
//		tokenService := &mocks.Service{}
//
//		validator := NewValidationHydrator(tokenService, nil, nil)
//
//		// when
//		validator.ResolveConnectorTokenHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		var authSession AuthenticationSession
//		err = json.NewDecoder(rr.Body).Decode(&authSession)
//		require.NoError(t, err)
//
//		assert.Equal(t, emptyAuthSession(), authSession)
//		mock.AssertExpectationsForObjects(t, tokenService)
//	})
//
//	t.Run("should return error when failed to unmarshal authentication session", func(t *testing.T) {
//		// given
//		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte("wrong body")))
//		require.NoError(t, err)
//		rr := httptest.NewRecorder()
//
//		validator := NewValidationHydrator(nil, nil, nil)
//
//		// when
//		validator.ResolveConnectorTokenHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusBadRequest, rr.Code)
//	})
//}
