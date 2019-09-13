package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	authenticationMocks "github.com/kyma-incubator/compass/components/connector/internal/authentication/mocks"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	certificatesMocks "github.com/kyma-incubator/compass/components/connector/internal/certificates/mocks"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	tokensMocks "github.com/kyma-incubator/compass/components/connector/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	clientId = "clientId"
)

var (
	CSR           = "Q1NSCg=="
	decodedCSR, _ = decodeStringFromBase64(CSR)
	subject       = certificates.CSRSubject{
		CommonName: "clientId",
		CSRSubjectConsts: certificates.CSRSubjectConsts{
			Country:            "country",
			Organization:       "organization",
			OrganizationalUnit: "organizationalunit",
			Locality:           "locality",
			Province:           "province",
		},
	}
	directorURL             = "https://compass-gateway.kyma.local/director/graphql"
	certSecuredConnectorURL = "https://compass-gateway-mtls.kyma.local/connector/graphql"
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
		authenticator.On("Authenticate", context.TODO()).Return(clientId, nil)

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(encodedChain, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService, subject.CSRSubjectConsts, directorURL, certSecuredConnectorURL)

		// when
		certificationResult, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), CSR)

		// then
		require.NoError(t, err)
		assert.Equal(t, certChainBase64, certificationResult.CertificateChain)
		assert.Equal(t, caCertificate, certificationResult.CaCertificate)
		assert.Equal(t, clientCertificate, certificationResult.ClientCertificate)
		mock.AssertExpectationsForObjects(t, tokenService, authenticator)
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
		authenticator.On("Authenticate", context.TODO()).Return("", fmt.Errorf("error"))

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(encodedChain, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService, subject.CSRSubjectConsts, directorURL, certSecuredConnectorURL)

		// when
		_, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), CSR)

		// then
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, tokenService, authenticator)
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
		authenticator.On("Authenticate", context.TODO()).Return(clientId, nil)

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(encodedChain, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService, subject.CSRSubjectConsts, directorURL, certSecuredConnectorURL)

		// when
		_, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), "not base 64 csr")

		// then
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, tokenService, authenticator)
	})

	t.Run("should return error when failed to sign CSR", func(t *testing.T) {
		// given
		tokenService := &tokensMocks.Service{}
		authenticator := &authenticationMocks.Authenticator{}
		authenticator.On("Authenticate", context.TODO()).Return(clientId, nil)

		certService := &certificatesMocks.Service{}
		certService.On("SignCSR", decodedCSR, subject).Return(certificates.EncodedCertificateChain{}, apperrors.Internal("error"))

		certificateResolver := NewCertificateResolver(authenticator, tokenService, certService, subject.CSRSubjectConsts, directorURL, certSecuredConnectorURL)

		// when
		_, err := certificateResolver.SignCertificateSigningRequest(context.TODO(), CSR)

		// then
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, tokenService, authenticator, certService)
	})
}

func TestCertificateResolver_Configuration(t *testing.T) {

	t.Run("should return configuration", func(t *testing.T) {
		// given
		authenticator := &authenticationMocks.Authenticator{}
		authenticator.On("Authenticate", context.Background()).Return(clientId, nil)
		tokenService := &tokensMocks.Service{}
		tokenService.On("CreateToken", subject.CommonName, tokens.CSRToken).Return(token, nil)

		certificateResolver := NewCertificateResolver(authenticator, tokenService, nil, subject.CSRSubjectConsts, directorURL, certSecuredConnectorURL)

		// when
		configurationResult, err := certificateResolver.Configuration(context.Background())

		// then
		require.NoError(t, err)
		assert.Equal(t, token, configurationResult.Token.Token)
		assert.Equal(t, &directorURL, configurationResult.ManagementPlaneInfo.DirectorURL)
		assert.Equal(t, &certSecuredConnectorURL, configurationResult.ManagementPlaneInfo.CertificateSecuredConnectorURL)
		assert.Equal(t, expectedSubject(subject.CSRSubjectConsts, subject.CommonName), configurationResult.CertificateSigningRequestInfo.Subject)
		assert.Equal(t, "rsa2048", configurationResult.CertificateSigningRequestInfo.KeyAlgorithm)
		mock.AssertExpectationsForObjects(t, tokenService, authenticator)
	})

	t.Run("should return error when failed to generate token", func(t *testing.T) {
		// given
		authenticator := &authenticationMocks.Authenticator{}
		authenticator.On("Authenticate", context.Background()).Return(clientId, nil)
		tokenService := &tokensMocks.Service{}
		tokenService.On("CreateToken", subject.CommonName, tokens.CSRToken).Return("", apperrors.Internal("error"))

		certificateResolver := NewCertificateResolver(authenticator, tokenService, nil, subject.CSRSubjectConsts, directorURL, certSecuredConnectorURL)

		// when
		configurationResult, err := certificateResolver.Configuration(context.Background())

		// then
		require.Error(t, err)
		require.Nil(t, configurationResult)
		mock.AssertExpectationsForObjects(t, tokenService, authenticator)
	})

	t.Run("should return error when failed to authenticate", func(t *testing.T) {
		// given
		authenticator := &authenticationMocks.Authenticator{}
		authenticator.On("Authenticate", context.Background()).Return("", apperrors.Forbidden("Error"))
		tokenService := &tokensMocks.Service{}

		certificateResolver := NewCertificateResolver(authenticator, tokenService, nil, subject.CSRSubjectConsts, directorURL, certSecuredConnectorURL)

		// when
		configurationResult, err := certificateResolver.Configuration(context.Background())

		// then
		require.Error(t, err)
		require.Nil(t, configurationResult)
		mock.AssertExpectationsForObjects(t, tokenService, authenticator)
	})

}

func expectedSubject(c certificates.CSRSubjectConsts, commonName string) string {
	return fmt.Sprintf("O=%s,OU=%s,L=%s,ST=%s,C=%s,CN=%s", c.Organization, c.OrganizationalUnit, c.Locality, c.Province, c.Country, commonName)
}
