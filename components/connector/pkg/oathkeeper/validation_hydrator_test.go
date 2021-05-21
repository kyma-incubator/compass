package oathkeeper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	revocationMocks "github.com/kyma-incubator/compass/components/connector/internal/revocation/mocks"
	mocks2 "github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	clientId = "abcd-client-id"
	hash     = "qwertyuiop"
)

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
		revokedCertsRepository := &revocationMocks.RevokedCertificatesRepository{}
		revokedCertsRepository.On("Contains", hash).Return(false)

		validator := NewValidationHydrator(certHeaderParser, revokedCertsRepository)

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

		validator := NewValidationHydrator(certHeaderParser, nil)

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
		revokedCertsRepository := &revocationMocks.RevokedCertificatesRepository{}
		revokedCertsRepository.On("Contains", hash).Return(true)

		validator := NewValidationHydrator(certHeaderParser, revokedCertsRepository)

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
