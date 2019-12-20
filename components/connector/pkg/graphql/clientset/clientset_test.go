package clientset

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
)

func Test_Clientset(t *testing.T) {

	clientId := "abcd-efgh"

	// given
	var err error
	token, err := tokenService.CreateToken(clientId, tokens.ApplicationToken)
	require.NoError(t, err)

	clientSet := NewConnectorClientSet(WithSkipTLSVerify(true))

	// when
	certificate, err := clientSet.GenerateCertificateForToken(token, externalAPIUrl)

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, certificate)
	assert.Equal(t, 2, len(certificate.Certificate))
	assert.NotEmpty(t, certificate.PrivateKey)

	// given
	certSecuredClient := clientSet.CertificateSecuredClient(externalAPIUrl, certificate)

	// when
	configuration, err := certSecuredClient.Configuration()

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, configuration)

	// when
	_, csr, err := NewCSR(configuration.CertificateSigningRequestInfo.Subject, nil)
	require.NoError(t, err)

	certResponse, err := certSecuredClient.SignCSR(encodeCSR(csr))

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, certResponse.CertificateChain)
	assert.NotEmpty(t, certResponse.ClientCertificate)

	// when
	revokeResponse, err := certSecuredClient.RevokeCertificate()

	// then
	require.NoError(t, err)
	require.True(t, revokeResponse)
	revocationCM, err := k8sClientSet.CoreV1().ConfigMaps("default").Get(testConfigMapName, v1.GetOptions{})
	require.NoError(t, err)
	assert.Len(t, revocationCM.Data, 1)
}
