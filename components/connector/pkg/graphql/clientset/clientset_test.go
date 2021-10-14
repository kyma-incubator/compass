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
	var err error
	token := "mock-token"

	clientSet := NewConnectorClientSet(WithSkipTLSVerify(true))

	// ctx := context.WithValue(context.Background(), authentication.ConsumerType, "Application")
	// ctx = context.WithValue(ctx, authentication.TenantKey, "tenant")
	ctx := context.TODO()

	// when
	certificate, err := clientSet.GenerateCertificateForToken(ctx, token, externalAPIUrl)

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, certificate)
	assert.Equal(t, 2, len(certificate.Certificate))
	assert.NotEmpty(t, certificate.PrivateKey)

	// given
	certSecuredClient := clientSet.CertificateSecuredClient(externalAPIUrl, certificate)

	// when
	configuration, err := certSecuredClient.Configuration(ctx)

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, configuration)

	// when
	_, csr, err := NewCSR(configuration.CertificateSigningRequestInfo.Subject, nil)
	require.NoError(t, err)

	certResponse, err := certSecuredClient.SignCSR(ctx, encodeCSR(csr))

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, certResponse.CertificateChain)
	assert.NotEmpty(t, certResponse.ClientCertificate)

	// when
	revokeResponse, err := certSecuredClient.RevokeCertificate(ctx)

	// then
	require.NoError(t, err)
	require.True(t, revokeResponse)
	revocationCM, err := k8sClientSet.CoreV1().ConfigMaps("default").Get(ctx, testConfigMapName, v1.GetOptions{})
	require.NoError(t, err)
	assert.Len(t, revocationCM.Data, 1)
}
