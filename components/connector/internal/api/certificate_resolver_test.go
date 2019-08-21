package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	authenticationMocks "github.com/kyma-incubator/compass/components/connector/internal/authentication/mocks"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	certificatesMocks "github.com/kyma-incubator/compass/components/connector/internal/certificates/mocks"
	tokensMocks "github.com/kyma-incubator/compass/components/connector/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	CSR           = "Q1NSCg=="
	decodedCSR, _ = decodeStringFromBase64(CSR)
	subject       = certificates.CSRSubject{
		CommonName:         "commonname",
		Country:            "country",
		Organization:       "organization",
		OrganizationalUnit: "organizationalunit",
		Locality:           "locality",
		Province:           "province",
	}
)

func TestCertificateResolver_SignCertificateSigningRequest(t *testing.T) {
	t.Run("should sign client certificate", func(t *testing.T) {
		// given
		certChainBase64 := "certChainBase64"
		caCertificate := "caCertificate"
		clientCertificate := "clientCertificate"

		encodedChain := certificates.EncodedCertificateChain{
			CertificateChain:  certChainBase64,
			CaCertificate:     caCertificate,
			ClientCertificate: clientCertificate,
		}

		tokenService := &tokensMocks.Service{}
		authenticator := &authenticationMocks.Authenticator{}

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(encodedChain, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService)

		// when
		certificationResult, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), CSR)

		// then
		require.NoError(t, err)
		assert.Equal(t, certChainBase64, certificationResult.Certificate)
		assert.Equal(t, caCertificate, certificationResult.CaCertificate)
		assert.Equal(t, clientCertificate, certificationResult.ClientCertificate)
	})

	t.Run("should return error when failed to decode base64", func(t *testing.T) {
		// given
		certChainBase64 := "certChainBase64"
		caCertificate := "caCertificate"
		clientCertificate := "clientCertificate"

		encodedChain := certificates.EncodedCertificateChain{
			CertificateChain:  certChainBase64,
			CaCertificate:     caCertificate,
			ClientCertificate: clientCertificate,
		}

		tokenService := &tokensMocks.Service{}
		authenticator := &authenticationMocks.Authenticator{}

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(encodedChain, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService)

		// when
		_, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), "not base 64 csr")

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to sign CSR", func(t *testing.T) {
		// given
		tokenService := &tokensMocks.Service{}
		authenticator := &authenticationMocks.Authenticator{}

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(certificates.EncodedCertificateChain{}, apperrors.Internal("error"))

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService)

		// when
		_, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), CSR)

		// then
		require.Error(t, err)
	})
}
