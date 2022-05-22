package tests

import (
	"crypto/rsa"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/stretchr/testify/assert"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

const (
	TestApp     = "mytestapp"
	TestRuntime = "mytestrunt"
)

var defaultAppNameNormalizer = &normalizer.DefaultNormalizator{}

func TestConnector(t *testing.T) {
	appInput := directorSchema.ApplicationRegisterInput{
		Name:           TestApp,
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	descr := "test"
	runtimeInput := directorSchema.RuntimeRegisterInput{
		Name:        TestRuntime,
		Description: &descr,
		Labels: directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appID, err := directorClient.CreateApplication(appInput)
	defer func() {
		err = directorClient.CleanupApplication(appID)
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	runtimeID, err := directorClient.CreateRuntime(runtimeInput)
	defer func() {
		err = directorClient.CleanupRuntime(runtimeID)
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	err = directorClient.SetDefaultEventing(runtimeID, appID, testConfig.EventsBaseURL)
	require.NoError(t, err)

	t.Run("Connector Service flow for Application", func(t *testing.T) {
		certificateGenerationSuite(t, directorClient, appID)
		certificateRotationSuite(t, directorClient, appID)
		certificateRevocationSuite(t, directorClient, appID)

		for _, shouldNormalizeEventURL := range []bool{true, false} {
			err := directorClient.SetRuntimeLabel(runtimeID, "isNormalized", strconv.FormatBool(shouldNormalizeEventURL))
			require.NoError(t, err)

			for _, appNameLabelExists := range []bool{true, false} {
				var err error
				if appNameLabelExists {
					err = directorClient.SetApplicationLabel(appID, "name", defaultAppNameNormalizer.Normalize(TestApp))
				} else {
					err = directorClient.DeleteApplicationLabel(appID, "name")
				}
				require.NoError(t, err)

				appMgmInfoEndpointSuite(t, directorClient, appID, TestApp, appNameLabelExists, shouldNormalizeEventURL)
				appCsrInfoEndpointSuite(t, directorClient, appID, TestApp, appNameLabelExists, shouldNormalizeEventURL)
			}
		}
	})
}

func certificateGenerationSuite(t *testing.T, directorClient clients.Client, appID string) {
	client := clients.NewConnectorClient(directorClient, appID, testConfig.Tenant, testConfig.SkipSSLValidation)

	clientKey := certs.CreateKey(t)

	t.Run("should create client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := certs.DecodeAndParseCerts(t, crtResponse)

		// then
		clientsCrt := certificates.CRTChain[0]
		certs.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
	})

	t.Run("should create two certificates in a chain", func(t *testing.T) {
		// when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := certs.DecodeAndParseCerts(t, crtResponse)

		// then
		require.Equal(t, 2, len(certificates.CRTChain))
	})

	t.Run("client cert should be signed by server cert", func(t *testing.T) {
		//when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := certs.DecodeAndParseCerts(t, crtResponse)

		//then
		certs.CheckIfCertIsSigned(t, certificates.CRTChain)
	})

	t.Run("should respond with client certificate together with CA crt", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := certs.DecodeAndParseCerts(t, crtResponse)

		// then
		clientsCrt := certificates.CRTChain[0]
		certs.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
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
		require.NotEmpty(t, infoResponse.CertURL)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// given
		infoResponse.Certificate.Subject = "subject=OU=Test,O=Test,L=Wrong,ST=Wrong,C=PL,CN=Wrong"
		csr := certs.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
		csrBase64 := certs.EncodeBase64(csr)

		// when
		_, err := client.CreateCertChain(t, csrBase64, infoResponse.CertURL)

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
		require.NotEmpty(t, infoResponse.CertURL)

		wrongUrl := replaceToken(t, infoResponse.CertURL, "incorrect-token")

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertURL)
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
		require.NotEmpty(t, infoResponse.CertURL)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// when
		_, err := client.CreateCertChain(t, "wrong-csr", infoResponse.CertURL)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.StatusCode)
		require.Equal(t, http.StatusBadRequest, err.ErrorResponse.Code)
		require.Contains(t, err.ErrorResponse.Error, "Error while parsing base64 content")
	})

}

func appCsrInfoEndpointSuite(t *testing.T, directorClient clients.Client, appID string, appName string, appNameLabelExists, shouldNormalizeEventURL bool) {
	t.Run("should use default values to build CSR info response", func(t *testing.T) {
		// given
		client := clients.NewConnectorClient(directorClient, appID, testConfig.Tenant, testConfig.SkipSSLValidation)
		expectedMetadataURL := testConfig.ConnectivityAdapterMtlsUrl
		expectedEventsURL := testConfig.EventsBaseURL

		normalizedAppName := defaultAppNameNormalizer.Normalize(appName)
		if appNameLabelExists {
			expectedMetadataURL += "/" + normalizedAppName + "/v1/metadata/services"
		} else {
			expectedMetadataURL += "/" + appName + "/v1/metadata/services"
		}

		if shouldNormalizeEventURL {
			expectedEventsURL += "/" + normalizedAppName + "/v1/events"
		} else {
			expectedEventsURL += "/" + appName + "/v1/events"
		}

		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		assert.Equal(t, expectedEventsURL, infoResponse.Api.RuntimeURLs.EventsURL)
		assert.Equal(t, expectedMetadataURL, infoResponse.Api.RuntimeURLs.MetadataURL)
	})
}

func appMgmInfoEndpointSuite(t *testing.T, directorClient clients.Client, appID string, appName string, appNameLabelExists, shouldNormalizeEventURL bool) {
	client := clients.NewConnectorClient(directorClient, appID, testConfig.Tenant, testConfig.SkipSSLValidation)

	clientKey := certs.CreateKey(t)

	t.Run("should use default values to build management info", func(t *testing.T) {
		// given
		expectedMetadataURL := testConfig.ConnectivityAdapterMtlsUrl
		expectedEventsURL := testConfig.EventsBaseURL

		normalizedAppName := defaultAppNameNormalizer.Normalize(appName)
		if appNameLabelExists {
			expectedMetadataURL += "/" + normalizedAppName + "/v1/metadata/services"
		} else {
			expectedMetadataURL += "/" + appName + "/v1/metadata/services"
		}

		if shouldNormalizeEventURL {
			expectedEventsURL += "/" + normalizedAppName + "/v1/events"
		} else {
			expectedEventsURL += "/" + appName + "/v1/events"
		}

		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		crtChainBytes := certs.DecodeBase64Cert(t, crtResponse.CRTChain)
		client, err := clients.NewSecuredClient(testConfig.SkipSSLValidation, clientKey, crtChainBytes, testConfig.Tenant)
		require.NoError(t, err)

		// when
		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)

		// then
		assert.Equal(t, expectedMetadataURL, mgmInfoResponse.URLs.MetadataURL)
		assert.Equal(t, expectedEventsURL, mgmInfoResponse.URLs.EventsURL)

		if appNameLabelExists {
			assert.Equal(t, normalizedAppName, mgmInfoResponse.ClientIdentity.Application)
		} else {
			assert.Equal(t, appName, mgmInfoResponse.ClientIdentity.Application)
		}

		assert.NotEmpty(t, mgmInfoResponse.Certificate.Subject)
		assert.Equal(t, clients.Extensions, mgmInfoResponse.Certificate.Extensions)
		assert.Equal(t, clients.KeyAlgorithm, mgmInfoResponse.Certificate.KeyAlgorithm)

		assert.Empty(t, mgmInfoResponse.ClientIdentity.Group)

		assert.Empty(t, mgmInfoResponse.ClientIdentity.Tenant)

	})
}

func certificateRotationSuite(t *testing.T, directorClient clients.Client, appID string) {
	client := clients.NewConnectorClient(directorClient, appID, testConfig.Tenant, testConfig.SkipSSLValidation)
	clientKey := certs.CreateKey(t)

	t.Run("should renew client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)
		require.NotEmpty(t, infoResponse.Certificate)

		crtChainBytes := certs.DecodeBase64Cert(t, crtResponse.CRTChain)
		client, err := clients.NewSecuredClient(testConfig.SkipSSLValidation, clientKey, crtChainBytes, testConfig.Tenant)
		require.NoError(t, err)

		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RenewCertURL)
		require.NotEmpty(t, mgmInfoResponse.Certificate)
		require.Equal(t, infoResponse.Certificate, mgmInfoResponse.Certificate)

		csr := certs.CreateCsr(t, mgmInfoResponse.Certificate.Subject, clientKey)
		csrBase64 := certs.EncodeBase64(csr)

		certificateResponse, errorResponse := client.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertURL, csrBase64)

		// then
		require.Nil(t, errorResponse)

		crtChainBytesRenewed := certs.DecodeBase64Cert(t, certificateResponse.CRTChain)
		clientWithRenewedCert, err := clients.NewSecuredClient(testConfig.SkipSSLValidation, clientKey, crtChainBytesRenewed, testConfig.Tenant)
		require.NoError(t, err)

		_, errorResponse = clientWithRenewedCert.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)
	})
}

func certificateRevocationSuite(t *testing.T, directorClient clients.Client, appID string) {
	client := clients.NewConnectorClient(directorClient, appID, testConfig.Tenant, testConfig.SkipSSLValidation)

	clientKey := certs.CreateKey(t)

	t.Run("should revoke client certificate with external API", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		// when
		crtChainBytes := certs.DecodeBase64Cert(t, crtResponse.CRTChain)
		client, err := clients.NewSecuredClient(testConfig.SkipSSLValidation, clientKey, crtChainBytes, testConfig.Tenant)
		require.NoError(t, err)

		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RevokeCertURL)

		// when
		errorResponse = client.RevokeCertificate(t, mgmInfoResponse.URLs.RevokeCertURL)

		// then
		require.Nil(t, errorResponse)

		// when
		csr := certs.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
		csrBase64 := certs.EncodeBase64(csr)

		_, errorResponse = client.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertURL, csrBase64)

		// then
		require.NotNil(t, errorResponse)
		require.Equal(t, http.StatusForbidden, errorResponse.StatusCode)
	})
}

func createCertificateChain(t *testing.T, connectorClient clients.ConnectorClient, key *rsa.PrivateKey) (*model.CrtResponse, *model.InfoResponse) {
	// when
	tokenResponse := connectorClient.CreateToken(t)

	// then
	require.NotEmpty(t, tokenResponse.Token)
	require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

	// when
	infoResponse, errorResponse := connectorClient.GetInfo(t, tokenResponse.URL)

	// then
	require.Nil(t, errorResponse)
	require.NotEmpty(t, infoResponse.CertURL)
	require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

	// given
	csr := certs.CreateCsr(t, infoResponse.Certificate.Subject, key)
	csrBase64 := certs.EncodeBase64(csr)

	// when
	crtResponse, errorResponse := connectorClient.CreateCertChain(t, csrBase64, infoResponse.CertURL)

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
