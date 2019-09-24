package oathkeeper

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	mocks2 "github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper/mocks"
	revocationMocks "github.com/kyma-incubator/compass/components/connector/internal/revocation/mocks"

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

	createAuthRequestWithToken := func(t *testing.T) *http.Request {
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		req.Header.Add(ConnectorTokenHeader, token)
		return req
	}

	t.Run("should resolve token and add header to response", func(t *testing.T) {
		// given
		req := createAuthRequestWithToken(t)
		rr := httptest.NewRecorder()

		tokenService := &mocks.Service{}
		tokenService.On("Resolve", token).Return(tokenData, nil)
		tokenService.On("Delete", token).Return(nil)

		validator := NewValidationHydrator(tokenService, nil, nil)

		// when
		validator.ResolveConnectorTokenHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{clientId}, authSession.Header[ClientIdFromTokenHeader])
		mock.AssertExpectationsForObjects(t, tokenService)
	})

	t.Run("should not modify authentication session if failed to resolved token", func(t *testing.T) {
		// given
		req := createAuthRequestWithToken(t)
		rr := httptest.NewRecorder()

		tokenService := &mocks.Service{}
		tokenService.On("Resolve", token).Return(tokens.TokenData{}, apperrors.NotFound("error"))

		validator := NewValidationHydrator(tokenService, nil, nil)

		// when
		validator.ResolveConnectorTokenHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
		mock.AssertExpectationsForObjects(t, tokenService)
	})

	t.Run("should not modify authentication session if no token provided", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		tokenService := &mocks.Service{}

		validator := NewValidationHydrator(tokenService, nil, nil)

		// when
		validator.ResolveConnectorTokenHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
		mock.AssertExpectationsForObjects(t, tokenService)
	})

	t.Run("should return error when failed to unmarshal authentication session", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte("wrong body")))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		validator := NewValidationHydrator(nil, nil, nil)

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
		revocationList := &revocationMocks.RevocationListRepository{}
		revocationList.On("Contains", hash).Return(false, nil)

		validator := NewValidationHydrator(nil, certHeaderParser, revocationList)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{clientId}, authSession.Header[ClientIdFromCertificateHeader])
		mock.AssertExpectationsForObjects(t, certHeaderParser)
	})

	t.Run("should not modify authentication session if no valid cert header found", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certHeaderParser := &mocks2.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return("", "", false)

		validator := NewValidationHydrator(nil, certHeaderParser, nil)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
		mock.AssertExpectationsForObjects(t, certHeaderParser)
	})

	t.Run("should not modify authentication session if certificate is revoked", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certHeaderParser := &mocks2.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return(clientId, hash, true)
		revocationList := &revocationMocks.RevocationListRepository{}
		revocationList.On("Contains", hash).Return(true, nil)

		validator := NewValidationHydrator(nil, certHeaderParser, revocationList)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
		mock.AssertExpectationsForObjects(t, certHeaderParser)
	})

	t.Run("should not modify authentication session if failed to read revoked certificates", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certHeaderParser := &mocks2.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return(clientId, hash, true)
		revocationList := &revocationMocks.RevocationListRepository{}
		revocationList.On("Contains", hash).Return(false, errors.Errorf("some error"))

		validator := NewValidationHydrator(nil, certHeaderParser, revocationList)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
		mock.AssertExpectationsForObjects(t, certHeaderParser)
	})

	t.Run("should return error when failed to unmarshal authentication session", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte("wrong body")))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		validator := NewValidationHydrator(nil, nil, nil)

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
