package clientset

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_Clientset(t *testing.T) {

	// given
	ctx := context.Background()

	var err error
	token := "mock-token"

	clientSet := NewConnectorClientSet(WithSkipTLSVerify(true))

	// when
	certificate, err := clientSet.GenerateCertificateForToken(context.TODO(), token, externalAPIUrl)

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, certificate)
	assert.Equal(t, 2, len(certificate.Certificate))
	assert.NotEmpty(t, certificate.PrivateKey)

	// given
	certSecuredClient := clientSet.CertificateSecuredClient(externalAPIUrl, certificate)

	// when
	configuration, err := certSecuredClient.Configuration(context.TODO())

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, configuration)

	// when
	_, csr, err := NewCSR(configuration.CertificateSigningRequestInfo.Subject, nil)
	require.NoError(t, err)

	certResponse, err := certSecuredClient.SignCSR(context.TODO(), encodeCSR(csr))

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, certResponse.CertificateChain)
	assert.NotEmpty(t, certResponse.ClientCertificate)

	// when
	revokeResponse, err := certSecuredClient.RevokeCertificate(context.TODO())

	// then
	require.NoError(t, err)
	require.True(t, revokeResponse)
	revocationCM, err := k8sClientSet.CoreV1().ConfigMaps("default").Get(ctx, testConfigMapName, v1.GetOptions{})
	require.NoError(t, err)
	assert.Len(t, revocationCM.Data, 1)
}
