package oathkeeper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mocks2 "github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper/mocks"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens/mocks"
)

const (
	token    = "abcd-token"
	clientId = "abcd-client-id"
	hash     = "qwertyuiop"
)

var (
	tokenData = tokens.TokenData{
		Type:     tokens.ApplicationToken,
		ClientId: clientId,
	}
)

func TestValidationHydrator_ResolveConnectorTokenHeader(t *testing.T) {
	marshalledSession, err := json.Marshal(emptyAuthSession())
	require.NoError(t, err)

	t.Run("should resolve token and add header to response", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		req.Header.Add(ConnectorTokenHeader, token)
		rr := httptest.NewRecorder()

		tokenService := &mocks.Service{}
		tokenService.On("Resolve", token).Return(tokenData, nil)
		tokenService.On("Delete", token).Return(nil)

		validator := NewValidationHydrator(tokenService, nil)

		// when
		validator.ResolveConnectorTokenHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{clientId}, authSession.Header[ClientIdFromTokenHeader])
	})

	t.Run("should not modify authentication session if failed to resolved token", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		req.Header.Add(ConnectorTokenHeader, token)
		rr := httptest.NewRecorder()

		tokenService := &mocks.Service{}
		tokenService.On("Resolve", token).Return(tokens.TokenData{}, apperrors.NotFound("error"))

		validator := NewValidationHydrator(tokenService, nil)

		// when
		validator.ResolveConnectorTokenHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
	})

	t.Run("should not modify authentication session if no token provided", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		tokenService := &mocks.Service{}

		validator := NewValidationHydrator(tokenService, nil)

		// when
		validator.ResolveConnectorTokenHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
	})

	t.Run("should return error when failed to unmarshal authentication session", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte("wrong body")))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		validator := NewValidationHydrator(nil, nil)

		// when
		validator.ResolveConnectorTokenHeader(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestValidationHydrator_ResolveIstioCertHeader(t *testing.T) {
	marshalledSession, err := json.Marshal(emptyAuthSession())
	require.NoError(t, err)

	t.Run("should resolve cert header and add header to response", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certHeaderParser := &mocks2.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return(clientId, hash, true)

		validator := NewValidationHydrator(nil, certHeaderParser)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{clientId}, authSession.Header[ClientIdFromCertificateHeader])
	})

	t.Run("should not modify authentication session if no valid cert header found", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certHeaderParser := &mocks2.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return("", "", false)

		validator := NewValidationHydrator(nil, certHeaderParser)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
	})

	t.Run("should return error when failed to unmarshal authentication session", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte("wrong body")))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		validator := NewValidationHydrator(nil, nil)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func emptyAuthSession() AuthenticationSession {
	return AuthenticationSession{
		Subject: "client",
		Extra:   nil,
		Header:  http.Header{},
	}
}
