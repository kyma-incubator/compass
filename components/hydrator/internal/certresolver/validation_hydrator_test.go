package certresolver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/internal/certresolver"

	"github.com/kyma-incubator/compass/components/hydrator/internal/certresolver/automock"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidationHydrator_ResolveIstioCertHeader(t *testing.T) {
	var certificateData = &certresolver.CertificateData{
		ClientID:         "abcd-client-id",
		CertificateHash:  "qwertyuiop",
		AuthSessionExtra: nil,
	}

	var certificateDataWithExtra = &certresolver.CertificateData{
		ClientID:        "abcd-client-id",
		CertificateHash: "qwertyuiop",
		AuthSessionExtra: map[string]interface{}{
			cert.ConsumerTypeExtraField:  "test_consumer_type",
			cert.InternalConsumerIDField: "test_internal_consumer_id",
			cert.AccessLevelsExtraField:  []interface{}{"test_access_level"},
		},
	}

	marshalledSession, err := json.Marshal(emptyAuthSession())
	require.NoError(t, err)

	issuer := "issuer"

	t.Run("should resolve cert header and add header to response", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certsHash := map[string]string{}

		certHeaderParser := &automock.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return(certificateData)
		certHeaderParser.On("GetIssuer").Return(issuer)

		revokedCertificatesCache := &automock.RevokedCertificatesCache{}
		revokedCertificatesCache.On("Get").Return(certsHash)

		validator := certresolver.NewValidationHydrator(revokedCertificatesCache, certHeaderParser)

		// when
		validator.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession oathkeeper.AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{certificateData.ClientID}, authSession.Header[oathkeeper.ClientIdFromCertificateHeader])
		assert.Equal(t, []string{issuer}, authSession.Header[oathkeeper.ClientCertificateIssuerHeader])
		mock.AssertExpectationsForObjects(t, certHeaderParser, revokedCertificatesCache)
	})

	t.Run("should resolve cert header and append extra", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certsHash := map[string]string{}

		certHeaderParser := &automock.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return(certificateDataWithExtra)
		certHeaderParser.On("GetIssuer").Return(issuer)

		revokedCertificatesCache := &automock.RevokedCertificatesCache{}
		revokedCertificatesCache.On("Get").Return(certsHash)

		validator := certresolver.NewValidationHydrator(revokedCertificatesCache, certHeaderParser)

		// when
		validator.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession oathkeeper.AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, []string{certificateDataWithExtra.ClientID}, authSession.Header[oathkeeper.ClientIdFromCertificateHeader])
		assert.Equal(t, []string{issuer}, authSession.Header[oathkeeper.ClientCertificateIssuerHeader])
		assert.Equal(t, certificateDataWithExtra.AuthSessionExtra, authSession.Extra)
		mock.AssertExpectationsForObjects(t, certHeaderParser, revokedCertificatesCache)
	})

	t.Run("should not modify authentication session if no valid cert header found", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		certHeaderParser := &automock.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return(nil)
		certHeaderParser.On("GetIssuer").Return(issuer)

		revokedCertificatesCache := &automock.RevokedCertificatesCache{}
		revokedCertificatesCache.AssertNotCalled(t, "Get")

		validator := certresolver.NewValidationHydrator(nil, certHeaderParser)

		// when
		validator.ServeHTTP(rr, req)

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

		certsHash := map[string]string{
			certificateData.CertificateHash: certificateData.ClientID,
		}

		certHeaderParser := &automock.CertificateHeaderParser{}
		certHeaderParser.On("GetCertificateData", req).Return(certificateData)
		certHeaderParser.On("GetIssuer").Return(issuer)
		revokedCertificatesCache := &automock.RevokedCertificatesCache{}
		revokedCertificatesCache.On("Get").Return(certsHash)

		validator := certresolver.NewValidationHydrator(revokedCertificatesCache, certHeaderParser)

		// when
		validator.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)

		var authSession oathkeeper.AuthenticationSession
		err = json.NewDecoder(rr.Body).Decode(&authSession)
		require.NoError(t, err)

		assert.Equal(t, emptyAuthSession(), authSession)
		mock.AssertExpectationsForObjects(t, certHeaderParser, revokedCertificatesCache)
	})

	t.Run("should return error when failed to unmarshal authentication session", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte("wrong body")))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		validator := certresolver.NewValidationHydrator(nil, nil)

		// when
		validator.ServeHTTP(rr, req)

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
