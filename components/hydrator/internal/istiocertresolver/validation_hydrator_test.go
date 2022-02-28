package istiocertresolver_test

//
//import (
//	"bytes"
//	"encoding/json"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	revocationMocks "github.com/kyma-incubator/compass/components/connector/internal/revocation/mocks"
//
//	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
//	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper/automock"
//	"github.com/kyma-incubator/compass/components/director/pkg/cert"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//	"github.com/stretchr/testify/require"
//)
//
//func TestValidationHydrator_ResolveIstioCertHeader(t *testing.T) {
//	var certificateData = &oathkeeper.CertificateData{
//		ClientID:         "abcd-client-id",
//		CertificateHash:  "qwertyuiop",
//		AuthSessionExtra: nil,
//	}
//
//	var certificateDataWithExtra = &oathkeeper.CertificateData{
//		ClientID:        "abcd-client-id",
//		CertificateHash: "qwertyuiop",
//		AuthSessionExtra: map[string]interface{}{
//			cert.ConsumerTypeExtraField:  "test_consumer_type",
//			cert.InternalConsumerIDField: "test_internal_consumer_id",
//			cert.AccessLevelsExtraField:  []interface{}{"test_access_level"},
//		},
//	}
//
//	marshalledSession, err := json.Marshal(emptyAuthSession())
//	require.NoError(t, err)
//
//	issuer := "issuer"
//
//	t.Run("should resolve cert header and add header to response", func(t *testing.T) {
//		// given
//		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
//		require.NoError(t, err)
//		rr := httptest.NewRecorder()
//
//		certHeaderParser := &automock.CertificateHeaderParser{}
//		certHeaderParser.On("GetCertificateData", req).Return(certificateData)
//		certHeaderParser.On("GetIssuer").Return(issuer)
//		revokedCertsRepository := &revocationMocks.RevokedCertificatesRepository{}
//		revokedCertsRepository.On("Contains", certificateData.CertificateHash).Return(false)
//
//		validator := oathkeeper.NewValidationHydrator(revokedCertsRepository, certHeaderParser)
//
//		// when
//		validator.ResolveIstioCertHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		var authSession oathkeeper.AuthenticationSession
//		err = json.NewDecoder(rr.Body).Decode(&authSession)
//		require.NoError(t, err)
//
//		assert.Equal(t, []string{certificateData.ClientID}, authSession.Header[oathkeeper.ClientIdFromCertificateHeader])
//		assert.Equal(t, []string{issuer}, authSession.Header[oathkeeper.ClientCertificateIssuerHeader])
//		mock.AssertExpectationsForObjects(t, certHeaderParser, revokedCertsRepository)
//	})
//
//	t.Run("should resolve cert header and append extra", func(t *testing.T) {
//		// given
//		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
//		require.NoError(t, err)
//		rr := httptest.NewRecorder()
//
//		certHeaderParser := &automock.CertificateHeaderParser{}
//		certHeaderParser.On("GetCertificateData", req).Return(certificateDataWithExtra)
//		certHeaderParser.On("GetIssuer").Return(issuer)
//		revokedCertsRepository := &revocationMocks.RevokedCertificatesRepository{}
//		revokedCertsRepository.On("Contains", certificateDataWithExtra.CertificateHash).Return(false)
//
//		validator := oathkeeper.NewValidationHydrator(revokedCertsRepository, certHeaderParser)
//
//		// when
//		validator.ResolveIstioCertHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		var authSession oathkeeper.AuthenticationSession
//		err = json.NewDecoder(rr.Body).Decode(&authSession)
//		require.NoError(t, err)
//
//		assert.Equal(t, []string{certificateDataWithExtra.ClientID}, authSession.Header[oathkeeper.ClientIdFromCertificateHeader])
//		assert.Equal(t, []string{issuer}, authSession.Header[oathkeeper.ClientCertificateIssuerHeader])
//		assert.Equal(t, certificateDataWithExtra.AuthSessionExtra, authSession.Extra)
//		mock.AssertExpectationsForObjects(t, certHeaderParser, revokedCertsRepository)
//	})
//
//	t.Run("should not modify authentication session if no valid cert header found", func(t *testing.T) {
//		// given
//		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
//		require.NoError(t, err)
//		rr := httptest.NewRecorder()
//
//		certHeaderParser := &automock.CertificateHeaderParser{}
//		certHeaderParser.On("GetCertificateData", req).Return(nil)
//		certHeaderParser.On("GetIssuer").Return(issuer)
//
//		validator := oathkeeper.NewValidationHydrator(nil, certHeaderParser)
//
//		// when
//		validator.ResolveIstioCertHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		var authSession oathkeeper.AuthenticationSession
//		err = json.NewDecoder(rr.Body).Decode(&authSession)
//		require.NoError(t, err)
//
//		assert.Equal(t, emptyAuthSession(), authSession)
//		mock.AssertExpectationsForObjects(t, certHeaderParser)
//	})
//
//	t.Run("should not modify authentication session if certificate is revoked", func(t *testing.T) {
//		// given
//		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(marshalledSession))
//		require.NoError(t, err)
//		rr := httptest.NewRecorder()
//
//		certHeaderParser := &automock.CertificateHeaderParser{}
//		certHeaderParser.On("GetCertificateData", req).Return(certificateData)
//		certHeaderParser.On("GetIssuer").Return(issuer)
//		revokedCertsRepository := &revocationMocks.RevokedCertificatesRepository{}
//		revokedCertsRepository.On("Contains", certificateData.CertificateHash).Return(true)
//
//		validator := oathkeeper.NewValidationHydrator(revokedCertsRepository, certHeaderParser)
//
//		// when
//		validator.ResolveIstioCertHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		var authSession oathkeeper.AuthenticationSession
//		err = json.NewDecoder(rr.Body).Decode(&authSession)
//		require.NoError(t, err)
//
//		assert.Equal(t, emptyAuthSession(), authSession)
//		mock.AssertExpectationsForObjects(t, certHeaderParser, revokedCertsRepository)
//	})
//
//	t.Run("should return error when failed to unmarshal authentication session", func(t *testing.T) {
//		// given
//		req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte("wrong body")))
//		require.NoError(t, err)
//		rr := httptest.NewRecorder()
//
//		validator := oathkeeper.NewValidationHydrator(nil, nil)
//
//		// when
//		validator.ResolveIstioCertHeader(rr, req)
//
//		// then
//		assert.Equal(t, http.StatusBadRequest, rr.Code)
//	})
//}
//
//func emptyAuthSession() oathkeeper.AuthenticationSession {
//	return oathkeeper.AuthenticationSession{
//		Subject: "client",
//		Extra:   nil,
//		Header:  http.Header{},
//	}
//}
