package oathkeeper_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"

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

	issuer := "issuer"

	t.Run("should resolve cert header and add header to response", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certHeaderParser := &mocks2.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return(clientId, hash, true)
		certHeaderParser.On("GetIssuer").Return(issuer)
		revokedCertsRepository := &revocationMocks.RevokedCertificatesRepository{}
		revokedCertsRepository.On("Contains", hash).Return(false)

		validator := oathkeeper.NewValidationHydrator(revokedCertsRepository, certHeaderParser)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession oathkeeper.AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{clientId}, authSession.Header[oathkeeper.ClientIdFromCertificateHeader])
		assert.Equal(t, []string{issuer}, authSession.Header[oathkeeper.ClientCertificateIssuerHeader])
		mock.AssertExpectationsForObjects(t, certHeaderParser)
	})

	t.Run("should not modify authentication session if no valid cert header found", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certHeaderParser := &mocks2.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return("", "", false)
		certHeaderParser.On("GetIssuer").Return(issuer)

		validator := oathkeeper.NewValidationHydrator(nil, certHeaderParser)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession oathkeeper.AuthenticationSession
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
		certHeaderParser.On("GetIssuer").Return(issuer)
		revokedCertsRepository := &revocationMocks.RevokedCertificatesRepository{}
		revokedCertsRepository.On("Contains", hash).Return(true)

		validator := oathkeeper.NewValidationHydrator(revokedCertsRepository, certHeaderParser)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession oathkeeper.AuthenticationSession
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

		validator := oathkeeper.NewValidationHydrator(nil, nil)

		// when
		validator.ResolveIstioCertHeader(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func emptyAuthSession() oathkeeper.AuthenticationSession {
	return oathkeeper.AuthenticationSession{
		Subject: "client",
		Extra:   nil,
		Header:  http.Header{},
	}
}
