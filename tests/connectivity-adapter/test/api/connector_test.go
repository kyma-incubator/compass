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
		Description:    ptr.String("my first wordpress application"),
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

	client, err := director.NewClient(config.DirectorUrl, config.Tenant, []string{"application:read", "application:write", "runtime:write", "runtime:read", "eventing:manage"})
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
		appName := "mytestapp"
		certificateGenerationSuite(t, client, appID, config)

		appMgmInfoEndpointSuite(t, client, appID, config, appName)
		appCsrInfoEndpointSuite(t, client, appID, config, appName)
		certificateRotationSuite(t, client, appID, config)

		subjectGenerationSuite(t, client, appID, config)

		certificateRevocationSuite(t, client, appID, config.Tenant, config.SkipSslVerify, createApplicationRevocationUrl(config))
	})
}

func certificateGenerationSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration) {

	client := testkit.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should create client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)

		// then
		clientsCrt := certificates.CRTChain[0]
		testkit.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
	})

	t.Run("should create two certificates in a chain", func(t *testing.T) {
		// when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)

		// then
		require.Equal(t, 2, len(certificates.CRTChain))
	})

	t.Run("client cert should be signed by server cert", func(t *testing.T) {
		//when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)

		//then
		testkit.CheckIfCertIsSigned(t, certificates.CRTChain)
	})

	t.Run("should respond with client certificate together with CA crt", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)

		// then
		clientsCrt := certificates.CRTChain[0]
		testkit.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
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
		csr := testkit.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
		csrBase64 := testkit.EncodeBase64(csr)

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

		wrongUrl := replaceToken(tokenResponse.URL, "incorrect-token")

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

		wrongUrl := replaceToken(infoResponse.CertUrl, "incorrect-token")

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
		require.Equal(t, "There was an error while parsing the base64 content. An incorrect value was provided.", err.ErrorResponse.Error)
	})

}

func appCsrInfoEndpointSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration, appName string) {

	t.Run("should use default values to build CSR info response", func(t *testing.T) {
		// given
		client := testkit.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)
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

func subjectGenerationSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration) {
	// TODO: check what should be tested here
	client := testkit.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)

	// when
	tokenResponse := client.CreateToken(t)

	// then
	require.NotEmpty(t, tokenResponse.Token)
	require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

	// when
	infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

	// then
	require.Nil(t, errorResponse)
	subject := testkit.ParseSubject(infoResponse.Certificate.Subject)
	require.NotEmpty(t, subject.Organization[0])
	require.NotEmpty(t, subject.OrganizationalUnit[0])
}

func appMgmInfoEndpointSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration, appName string) {

	client := testkit.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)

	clientKey := testkit.CreateKey(t)

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

		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(config.SkipSslVerify, clientKey, certificates.ClientCRT.Raw, config.Tenant)

		// when
		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)

		// then
		assert.Equal(t, expectedMetadataURL, mgmInfoResponse.URLs.MetadataUrl)
		assert.Equal(t, expectedEventsURL, mgmInfoResponse.URLs.EventsUrl)
		assert.Equal(t, appName, mgmInfoResponse.ClientIdentity.Application)
		assert.NotEmpty(t, mgmInfoResponse.Certificate.Subject)
		assert.Equal(t, testkit.Extensions, mgmInfoResponse.Certificate.Extensions)
		assert.Equal(t, testkit.KeyAlgorithm, mgmInfoResponse.Certificate.KeyAlgorithm)

		assert.Empty(t, mgmInfoResponse.ClientIdentity.Group)
		assert.Empty(t, mgmInfoResponse.ClientIdentity.Tenant)

	})
}

func certificateRotationSuite(t *testing.T, directorClient director.Client, appID string, config testkit.Configuration) {
	client := testkit.NewConnectorClient(directorClient, appID, config.Tenant, config.SkipSslVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should renew client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)
		require.NotEmpty(t, infoResponse.Certificate)

		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(config.SkipSslVerify, clientKey, certificates.ClientCRT.Raw, config.Tenant)

		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RenewCertUrl)
		require.NotEmpty(t, mgmInfoResponse.Certificate)
		require.Equal(t, infoResponse.Certificate, mgmInfoResponse.Certificate)

		csr := testkit.CreateCsr(t, mgmInfoResponse.Certificate.Subject, clientKey)
		csrBase64 := testkit.EncodeBase64(csr)

		certificateResponse, errorResponse := client.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertUrl, csrBase64)

		// then
		require.Nil(t, errorResponse)

		certificates = testkit.DecodeAndParseCerts(t, certificateResponse)
		clientWithRenewedCert := testkit.NewSecuredConnectorClient(config.SkipSslVerify, clientKey, certificates.ClientCRT.Raw, config.Tenant)

		mgmInfoResponse, errorResponse = clientWithRenewedCert.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)
	})
}

func certificateRevocationSuite(t *testing.T, directorClient director.Client, appID, tenant string, skipVerify bool, internalRevocationUrl string) {
	client := testkit.NewConnectorClient(directorClient, appID, tenant, skipVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should revoke client certificate with external API", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw, tenant)

		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RevokeCertURL)

		// when
		errorResponse = client.RevokeCertificate(t, mgmInfoResponse.URLs.RevokeCertURL)

		// then
		require.Nil(t, errorResponse)

		// when
		csr := testkit.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
		csrBase64 := testkit.EncodeBase64(csr)

		_, errorResponse = client.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertUrl, csrBase64)

		// then
		require.NotNil(t, errorResponse)
		require.Equal(t, http.StatusForbidden, errorResponse.StatusCode)
	})
}

func createCertificateChain(t *testing.T, connectorClient testkit.ConnectorClient, key *rsa.PrivateKey) (*testkit.CrtResponse, *testkit.InfoResponse) {
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
	csr := testkit.CreateCsr(t, infoResponse.Certificate.Subject, key)
	csrBase64 := testkit.EncodeBase64(csr)

	// when
	crtResponse, errorResponse := connectorClient.CreateCertChain(t, csrBase64, infoResponse.CertUrl)

	// then
	require.Nil(t, errorResponse)

	return crtResponse, infoResponse
}

func replaceToken(originalUrl string, newToken string) string {
	parsedUrl, _ := url.Parse(originalUrl)
	queryParams, _ := url.ParseQuery(parsedUrl.RawQuery)

	queryParams.Set("token", newToken)
	parsedUrl.RawQuery = queryParams.Encode()

	return parsedUrl.String()
}

func createApplicationRevocationUrl(config testkit.Configuration) string {
	return config.ConnectivityAdapterMtlsUrl + "/v1/applications/certificates/revocations"
}
