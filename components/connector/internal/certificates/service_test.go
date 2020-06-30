package certificates_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"

	certificatesMocks "github.com/kyma-incubator/compass/components/connector/internal/certificates/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authSecretName   = "nginx-auth-ca"
	rootCASecretName = "rootCA-secret"

	caCertificateSecretKey     = "ca.crt"
	caKeySecretKey             = "ca.key"
	rootCACertificateSecretKey = "cacert"

	appName            = "appName"
	country            = "country"
	organization       = "organization"
	organizationalUnit = "organizationalUnit"
	locality           = "locality"
	province           = "province"
)

var (
	rawCSR = []byte("csr")

	caCrtEncoded = []byte("caCrtEncoded")
	caKeyEncoded = []byte("caKeyEncoded")

	certsSecretData = map[string][]byte{
		caCertificateSecretKey: caCrtEncoded,
		caKeySecretKey:         caKeyEncoded,
	}

	rootCaEncoded = []byte("rootCAEncoded")

	rootCASecretData = map[string][]byte{
		rootCACertificateSecretKey: rootCaEncoded,
	}

	rootCACrt = &x509.Certificate{}
	caCrt     = &x509.Certificate{}
	caKey     = &rsa.PrivateKey{}
	csr       = &x509.CertificateRequest{}

	rootCACrtBytes = []byte("rootCACertificate")
	clientCRT      = []byte("clientCertificate")
	clientCRTBytes = []byte("clientCertificateBytes")
	caCRTBytes     = []byte("caCRTBytes")
	certChain      = append(clientCRTBytes, caCRTBytes...)

	subjectValues = certificates.CSRSubject{
		CommonName: appName,
		CSRSubjectConsts: certificates.CSRSubjectConsts{
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		},
	}
)

func TestCertificateService_SignCSR(t *testing.T) {

	t.Run("should create certificate", func(t *testing.T) {
		// given
		cache := certificates.NewCertificateCache()
		cache.Put(authSecretName, certsSecretData)

		certUtils := &certificatesMocks.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("SignCSR", caCrt, csr, caKey).Return(clientCRT, nil)
		certUtils.On("AddCertificateHeaderAndFooter", caCrt.Raw).Return(caCRTBytes)
		certUtils.On("AddCertificateHeaderAndFooter", clientCRT).Return(clientCRTBytes)

		certificatesService := certificates.NewCertificateService(
			cache,
			certUtils,
			authSecretName,
			"",
			caCertificateSecretKey,
			caKeySecretKey,
			rootCACertificateSecretKey)

		// when
		encodedCertChain, apperr := certificatesService.SignCSR(rawCSR, subjectValues)

		// then
		require.NoError(t, apperr)
		assert.NotEmpty(t, encodedCertChain)

		decodedClientCRT, err := decodeBase64(encodedCertChain.ClientCertificate)
		require.NoError(t, err)
		assert.Equal(t, clientCRTBytes, decodedClientCRT)

		decodedChain, err := decodeBase64(encodedCertChain.CertificateChain)
		require.NoError(t, err)
		assert.Equal(t, certChain, decodedChain)

		certUtils.AssertExpectations(t)
	})

	t.Run("should create certificate with additional root certificate", func(t *testing.T) {
		// given
		certChain := append(append(clientCRTBytes, caCRTBytes...), rootCACrtBytes...)

		cache := certificates.NewCertificateCache()
		cache.Put(authSecretName, certsSecretData)
		cache.Put(rootCASecretName, rootCASecretData)

		certUtils := &certificatesMocks.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil).
			On("LoadCert", rootCaEncoded).Return(rootCACrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("SignCSR", caCrt, csr, caKey).Return(clientCRT, nil)
		certUtils.On("AddCertificateHeaderAndFooter", caCrt.Raw).Return(caCRTBytes).Once().
			On("AddCertificateHeaderAndFooter", rootCACrt.Raw).Return(rootCACrtBytes)
		certUtils.On("AddCertificateHeaderAndFooter", clientCRT).Return(clientCRTBytes)

		certificatesService := certificates.NewCertificateService(
			cache,
			certUtils,
			authSecretName,
			rootCASecretName,
			caCertificateSecretKey,
			caKeySecretKey,
			rootCACertificateSecretKey)

		// when
		encodedCertChain, apperr := certificatesService.SignCSR(rawCSR, subjectValues)

		// then
		require.NoError(t, apperr)
		assert.NotEmpty(t, encodedCertChain)

		decodedClientCRT, err := decodeBase64(encodedCertChain.ClientCertificate)
		require.NoError(t, err)
		assert.Equal(t, clientCRTBytes, decodedClientCRT)

		decodedChain, err := decodeBase64(encodedCertChain.CertificateChain)
		require.NoError(t, err)
		assert.Equal(t, certChain, decodedChain)

		certUtils.AssertExpectations(t)
	})

	t.Run("should return Not Found error when secret not found", func(t *testing.T) {
		// given
		cache := certificates.NewCertificateCache()
		certUtils := &certificatesMocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)

		certificatesService := certificates.NewCertificateService(
			cache,
			certUtils,
			authSecretName,
			"",
			caCertificateSecretKey,
			caKeySecretKey,
			rootCACertificateSecretKey)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, subjectValues)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.Empty(t, encodedChain)
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when couldn't load csr", func(t *testing.T) {
		// given
		cache := certificates.NewCertificateCache()

		certUtils := &certificatesMocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(nil, apperrors.Internal("error"))

		certificatesService := certificates.NewCertificateService(
			cache,
			certUtils,
			authSecretName,
			"",
			caCertificateSecretKey,
			caKeySecretKey,
			rootCACertificateSecretKey)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, subjectValues)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when subject check failed", func(t *testing.T) {
		// given
		cache := certificates.NewCertificateCache()

		certUtils := &certificatesMocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(apperrors.Forbidden("error"))

		certificatesService := certificates.NewCertificateService(
			cache,
			certUtils,
			authSecretName,
			"",
			caCertificateSecretKey,
			caKeySecretKey,
			rootCACertificateSecretKey)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, subjectValues)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeForbidden, err.Code())
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when couldn't load cert", func(t *testing.T) {
		// given
		cache := certificates.NewCertificateCache()
		cache.Put(authSecretName, certsSecretData)

		certUtils := &certificatesMocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("LoadCert", caCrtEncoded).Return(nil, apperrors.Internal("error"))

		certificatesService := certificates.NewCertificateService(
			cache,
			certUtils,
			authSecretName,
			"",
			caCertificateSecretKey,
			caKeySecretKey,
			rootCACertificateSecretKey)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, subjectValues)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when couldn't load key", func(t *testing.T) {
		// given
		cache := certificates.NewCertificateCache()
		cache.Put(authSecretName, certsSecretData)

		certUtils := &certificatesMocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(nil, apperrors.Internal("error"))

		certificatesService := certificates.NewCertificateService(
			cache,
			certUtils,
			authSecretName,
			"",
			caCertificateSecretKey,
			caKeySecretKey,
			rootCACertificateSecretKey)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, subjectValues)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when failed to sign CSR", func(t *testing.T) {
		// given
		cache := certificates.NewCertificateCache()
		cache.Put(authSecretName, certsSecretData)

		certUtils := &certificatesMocks.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("SignCSR", caCrt, csr, caKey).Return(nil, apperrors.Internal("error"))

		certificatesService := certificates.NewCertificateService(
			cache,
			certUtils,
			authSecretName,
			"",
			caCertificateSecretKey,
			caKeySecretKey,
			rootCACertificateSecretKey)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, subjectValues)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		certUtils.AssertExpectations(t)
	})
}

func decodeBase64(base64CrtChain string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64CrtChain)
}
