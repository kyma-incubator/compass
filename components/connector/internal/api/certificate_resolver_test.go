package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"

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
		CommonName: "commonname",
		CSRSubjectConsts: certificates.CSRSubjectConsts{
			Country:            "country",
			Organization:       "organization",
			OrganizationalUnit: "organizationalunit",
			Locality:           "locality",
			Province:           "province",
		},
	}
	directorURL = "https://compass-gateway.kyma.local/director/graphql"
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
		authenticator.On("AuthenticateToken", context.TODO()).Return(subject.CommonName, nil)

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(encodedChain, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService, subject.CSRSubjectConsts, directorURL)

		// when
		certificationResult, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), CSR)

		// then
		require.NoError(t, err)
		assert.Equal(t, certChainBase64, certificationResult.CertificateChain)
		assert.Equal(t, caCertificate, certificationResult.CaCertificate)
		assert.Equal(t, clientCertificate, certificationResult.ClientCertificate)
	})

	t.Run("should return error when unauthenticated call", func(t *testing.T) {
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
		authenticator.On("AuthenticateToken", context.TODO()).Return("", fmt.Errorf("error"))

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(encodedChain, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService, subject.CSRSubjectConsts, directorURL)

		// when
		_, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), CSR)

		// then
		require.Error(t, err)
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
		authenticator.On("AuthenticateToken", context.TODO()).Return(subject.CommonName, nil)

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(encodedChain, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService, subject.CSRSubjectConsts, directorURL)

		// when
		_, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), "not base 64 csr")

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to sign CSR", func(t *testing.T) {
		// given
		tokenService := &tokensMocks.Service{}
		authenticator := &authenticationMocks.Authenticator{}
		authenticator.On("AuthenticateToken", context.TODO()).Return(subject.CommonName, nil)

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(certificates.EncodedCertificateChain{}, apperrors.Internal("error"))

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService, subject.CSRSubjectConsts, directorURL)

		// when
		_, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), CSR)

		// then
		require.Error(t, err)
	})
}

func TestCertificateResolver_Configuration(t *testing.T) {

	t.Run("should return configuration", func(t *testing.T) {
		// given
		authenticator := &authenticationMocks.Authenticator{}
		authenticator.On("AuthenticateToken", context.Background()).Return(subject.CommonName, nil)
		tokenService := &tokensMocks.Service{}
		tokenService.On("CreateToken", subject.CommonName, tokens.CSRToken).Return(token, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, nil, subject.CSRSubjectConsts, directorURL)

		// when
		configurationResult, err := certificateResolver.Configuration(context.Background())

		// then
		require.NoError(t, err)
		assert.Equal(t, token, configurationResult.Token.Token)
		assert.Equal(t, directorURL, configurationResult.ManagementPlaneInfo.DirectorURL)
		assert.Equal(t, expectedSubject(subject.CSRSubjectConsts, subject.CommonName), configurationResult.CertificateSigningRequestInfo.Subject)
		assert.Equal(t, "rsa2048", configurationResult.CertificateSigningRequestInfo.KeyAlgorithm)
	})

	t.Run("should return error when failed to generate token", func(t *testing.T) {
		// given
		authenticator := &authenticationMocks.Authenticator{}
		authenticator.On("AuthenticateToken", context.Background()).Return(subject.CommonName, nil)
		tokenService := &tokensMocks.Service{}
		tokenService.On("CreateToken", subject.CommonName, tokens.CSRToken).Return("", apperrors.Internal("error"))

		certificateResolver := NewCertificateResolver(authenticator, tokenService, nil, subject.CSRSubjectConsts, directorURL)

		// when
		configurationResult, err := certificateResolver.Configuration(context.Background())

		// then
		require.Error(t, err)
		require.Nil(t, configurationResult)
	})

	t.Run("should return error when failed to authenticate", func(t *testing.T) {
		// given
		authenticator := &authenticationMocks.Authenticator{}
		authenticator.On("AuthenticateToken", context.Background()).Return("", apperrors.Forbidden("Error"))
		tokenService := &tokensMocks.Service{}

		certificateResolver := NewCertificateResolver(authenticator, tokenService, nil, subject.CSRSubjectConsts, directorURL)

		// when
		configurationResult, err := certificateResolver.Configuration(context.Background())

		// then
		require.Error(t, err)
		require.Nil(t, configurationResult)
	})

}

func expectedSubject(c certificates.CSRSubjectConsts, commonName string) string {
	return fmt.Sprintf("O=%s,OU=%s,L=%s,ST=%s,C=%s,CN=%s", c.Organization, c.OrganizationalUnit, c.Locality, c.Province, c.Country, commonName)
}
