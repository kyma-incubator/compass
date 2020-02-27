package api

import (
	"crypto/rsa"
	"net/http"
	"net/url"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/connectivity-adapter/test/testkit/director"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/connectivity-adapter/test/testkit"
	"github.com/kyma-incubator/compass/tests/connectivity-adapter/test/testkit/connector"
	"github.com/stretchr/testify/require"
)

const (
	TestApp     = "mytestapp"
	TestRuntime = "mytestrunt"
)

func TestConnector(t *testing.T) {
	appInput := directorSchema.ApplicationRegisterInput{
		Name:           TestApp,
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: &directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	descr := "test"
	runtimeInput := directorSchema.RuntimeInput{
		Name:        TestRuntime,
		Description: &descr,
		Labels: &directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	config, err := testkit.ReadConfiguration()
	require.NoError(t, err)

	client, err := director.NewClient(
		config.DirectorUrl,
		config.DirectorHealthzUrl,
		config.Tenant,
		[]string{"application:read", "application:write", "runtime:write", "runtime:read", "eventing:manage"})
	require.NoError(t, err)

	appID, err := client.CreateApplication(appInput)
	require.NoError(t, err)

	defer func() {
		err = client.DeleteApplication(appID)
		require.NoError(t, err)
	}()

	runtimeID, err := client.CreateRuntime(runtimeInput)
	require.NoError(t, err)

	defer func() {
		err = client.DeleteRuntime(runtimeID)
		require.NoError(t, err)
	}()

	err = client.SetDefaultEventing(runtimeID, appID, config.EventsBaseURL)
	require.NoError(t, err)

	t.Run("Connector Service flow for Application", func(t *testing.T) {

		certificateGenerationSuite(t, client, appID, config)
		certificateRotationSuite(t, client, appID, config)
		certificateRevocationSuite(t, client, appID, config.Tenant, config.SkipSslVerify)

		appMgmInfoEndpointSuite(t, client, appID, config, TestApp)
		appCsrInfoEndpointSuite(t, client, appID, config, TestApp)
	})
}

func certificateGenerationSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration) {

	client := connector.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)

	clientKey := connector.CreateKey(t)

	t.Run("should create client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := connector.DecodeAndParseCerts(t, crtResponse)

		// then
		clientsCrt := certificates.CRTChain[0]
		connector.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
	})

	t.Run("should create two certificates in a chain", func(t *testing.T) {
		// when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := connector.DecodeAndParseCerts(t, crtResponse)

		// then
		require.Equal(t, 2, len(certificates.CRTChain))
	})

	t.Run("client cert should be signed by server cert", func(t *testing.T) {
		//when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := connector.DecodeAndParseCerts(t, crtResponse)

		//then
		connector.CheckIfCertIsSigned(t, certificates.CRTChain)
	})

	t.Run("should respond with client certificate together with CA crt", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := connector.DecodeAndParseCerts(t, crtResponse)

		// then
		clientsCrt := certificates.CRTChain[0]
		connector.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
		require.Equal(t, certificates.ClientCRT, clientsCrt)

		caCrt := certificates.CRTChain[1]
		require.Equal(t, certificates.CaCRT, caCrt)
	})

	t.Run("should validate CSR subject", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// given
		infoResponse.Certificate.Subject = "subject=OU=Test,O=Test,L=Wrong,ST=Wrong,C=PL,CN=Wrong"
		csr := connector.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
		csrBase64 := connector.EncodeBase64(csr)

		// when
		_, err := client.CreateCertChain(t, csrBase64, infoResponse.CertUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.StatusCode)
		require.Equal(t, http.StatusBadRequest, err.ErrorResponse.Code)
		require.Contains(t, err.ErrorResponse.Error, "CSR: Invalid common name provided.")
	})

	t.Run("should return error for wrong token on info endpoint", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		wrongUrl := replaceToken(t, tokenResponse.URL, "incorrect-token")

		// when
		_, err := client.GetInfo(t, wrongUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
		require.Equal(t, http.StatusForbidden, err.ErrorResponse.Code)
		require.Equal(t, "invalid token or certificate", err.ErrorResponse.Error)
	})

	t.Run("should return error for wrong token on client-certs", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)

		wrongUrl := replaceToken(t, infoResponse.CertUrl, "incorrect-token")

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// when
		_, err := client.CreateCertChain(t, "csr", wrongUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
		require.Equal(t, http.StatusForbidden, err.ErrorResponse.Code)
		require.Equal(t, "invalid token or certificate", err.ErrorResponse.Error)
	})

	t.Run("should return error on wrong CSR on client-certs", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// when
		_, err := client.CreateCertChain(t, "wrong-csr", infoResponse.CertUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.StatusCode)
		require.Equal(t, http.StatusBadRequest, err.ErrorResponse.Code)
		require.Contains(t, err.ErrorResponse.Error, "Error while parsing base64 content")
	})

}

func appCsrInfoEndpointSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration, appName string) {

	t.Run("should use default values to build CSR info response", func(t *testing.T) {
		// given
		client := connector.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)
		expectedMetadataURL := config.ConnectivityAdapterMtlsUrl
		expectedEventsURL := config.EventsBaseURL

		expectedMetadataURL += "/" + appName + "/v1/metadata/services"
		expectedEventsURL += "/" + appName + "/v1/events"

		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		assert.Equal(t, expectedEventsURL, infoResponse.Api.RuntimeURLs.EventsUrl)
		assert.Equal(t, expectedMetadataURL, infoResponse.Api.RuntimeURLs.MetadataUrl)
	})
}

func appMgmInfoEndpointSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration, appName string) {

	client := connector.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)

	clientKey := connector.CreateKey(t)

	t.Run("should use default values to build management info", func(t *testing.T) {
		// given
		expectedMetadataURL := config.ConnectivityAdapterMtlsUrl
		expectedEventsURL := config.EventsBaseURL

		expectedMetadataURL += "/" + appName + "/v1/metadata/services"
		expectedEventsURL += "/" + appName + "/v1/events"

		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		certificates := connector.DecodeAndParseCerts(t, crtResponse)
		client := connector.NewSecuredConnectorClient(config.SkipSslVerify, clientKey, certificates.ClientCRT.Raw, config.Tenant)

		// when
		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)

		// then
		assert.Equal(t, expectedMetadataURL, mgmInfoResponse.URLs.MetadataUrl)
		assert.Equal(t, expectedEventsURL, mgmInfoResponse.URLs.EventsUrl)
		assert.Equal(t, appName, mgmInfoResponse.ClientIdentity.Application)
		assert.NotEmpty(t, mgmInfoResponse.Certificate.Subject)
		assert.Equal(t, connector.Extensions, mgmInfoResponse.Certificate.Extensions)
		assert.Equal(t, connector.KeyAlgorithm, mgmInfoResponse.Certificate.KeyAlgorithm)

		assert.Empty(t, mgmInfoResponse.ClientIdentity.Group)

		assert.Empty(t, mgmInfoResponse.ClientIdentity.Tenant)

	})
}

func certificateRotationSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration) {
	client := connector.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)

	clientKey := connector.CreateKey(t)

	t.Run("should renew client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)
		require.NotEmpty(t, infoResponse.Certificate)

		certificates := connector.DecodeAndParseCerts(t, crtResponse)
		client := connector.NewSecuredConnectorClient(config.SkipSslVerify, clientKey, certificates.ClientCRT.Raw, config.Tenant)

		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RenewCertUrl)
		require.NotEmpty(t, mgmInfoResponse.Certificate)
		require.Equal(t, infoResponse.Certificate, mgmInfoResponse.Certificate)

		csr := connector.CreateCsr(t, mgmInfoResponse.Certificate.Subject, clientKey)
		csrBase64 := connector.EncodeBase64(csr)

		certificateResponse, errorResponse := client.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertUrl, csrBase64)

		// then
		require.Nil(t, errorResponse)

		certificates = connector.DecodeAndParseCerts(t, certificateResponse)
		clientWithRenewedCert := connector.NewSecuredConnectorClient(config.SkipSslVerify, clientKey, certificates.ClientCRT.Raw, config.Tenant)

		mgmInfoResponse, errorResponse = clientWithRenewedCert.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)
	})
}

func certificateRevocationSuite(t *testing.T, directorClient director.Client, appID, tenant string, skipVerify bool) {
	client := connector.NewConnectorClient(directorClient, appID, tenant, skipVerify)

	clientKey := connector.CreateKey(t)

	t.Run("should revoke client certificate with external API", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		// when
		certificates := connector.DecodeAndParseCerts(t, crtResponse)
		client := connector.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw, tenant)

		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RevokeCertURL)

		// when
		errorResponse = client.RevokeCertificate(t, mgmInfoResponse.URLs.RevokeCertURL)

		// then
		require.Nil(t, errorResponse)

		// when
		csr := connector.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
		csrBase64 := connector.EncodeBase64(csr)

		_, errorResponse = client.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertUrl, csrBase64)

		// then
		require.NotNil(t, errorResponse)
		require.Equal(t, http.StatusForbidden, errorResponse.StatusCode)
	})
}

func createCertificateChain(t *testing.T, connectorClient connector.ConnectorClient, key *rsa.PrivateKey) (*connector.CrtResponse, *connector.InfoResponse) {
	// when
	tokenResponse := connectorClient.CreateToken(t)

	// then
	require.NotEmpty(t, tokenResponse.Token)
	require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

	// when
	infoResponse, errorResponse := connectorClient.GetInfo(t, tokenResponse.URL)

	// then
	require.Nil(t, errorResponse)
	require.NotEmpty(t, infoResponse.CertUrl)
	require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

	// given
	csr := connector.CreateCsr(t, infoResponse.Certificate.Subject, key)
	csrBase64 := connector.EncodeBase64(csr)

	// when
	crtResponse, errorResponse := connectorClient.CreateCertChain(t, csrBase64, infoResponse.CertUrl)

	// then
	require.Nil(t, errorResponse)

	return crtResponse, infoResponse
}

func replaceToken(t *testing.T, originalUrl string, newToken string) string {
	parsedUrl, err := url.Parse(originalUrl)
	require.NoError(t, err)

	queryParams, err := url.ParseQuery(parsedUrl.RawQuery)
	require.NoError(t, err)

	queryParams.Set("token", newToken)
	parsedUrl.RawQuery = queryParams.Encode()

	return parsedUrl.String()
}
