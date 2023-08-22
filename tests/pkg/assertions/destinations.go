package assertions

import (
	"encoding/json"
	directordestinationcreator "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"
	esmdestinationcreator "github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func AssertNoDestinationIsFound(t *testing.T, client *clients.DestinationClient, destinationName string) {
	_ = client.GetDestinationByName(t, destinationName, http.StatusNotFound)
}

func AssertNoDestinationCertificateIsFound(t *testing.T, client *clients.DestinationClient, certificateName string) {
	_ = client.GetDestinationCertificateByName(t, certificateName, http.StatusNotFound)
}

func AssertNoAuthDestination(t *testing.T, client *clients.DestinationClient, noAuthDestinationName, noAuthDestinationURL string) {
	noAuthDestBytes := client.GetDestinationByName(t, noAuthDestinationName, http.StatusOK)
	var noAuthDest esmdestinationcreator.NoAuthenticationDestination
	err := json.Unmarshal(noAuthDestBytes, &noAuthDest)
	require.NoError(t, err)
	require.Equal(t, noAuthDestinationName, noAuthDest.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, noAuthDest.Type)
	require.Equal(t, noAuthDestinationURL, noAuthDest.URL)
	require.Equal(t, directordestinationcreator.AuthTypeNoAuth, noAuthDest.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, noAuthDest.ProxyType)
}

func AssertBasicDestination(t *testing.T, client *clients.DestinationClient, basicDestinationName, basicDestinationURL string) {
	basicDestBytes := client.GetDestinationByName(t, basicDestinationName, http.StatusOK)
	var basicDest esmdestinationcreator.BasicDestination
	err := json.Unmarshal(basicDestBytes, &basicDest)
	require.NoError(t, err)
	require.Equal(t, basicDestinationName, basicDest.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, basicDest.Type)
	require.Equal(t, basicDestinationURL, basicDest.URL)
	require.Equal(t, directordestinationcreator.AuthTypeBasic, basicDest.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, basicDest.ProxyType)
}

func AssertSAMLAssertionDestination(t *testing.T, client *clients.DestinationClient, samlAssertionDestinationName, samlAssertionCertName, samlAssertionDestinationURL, app1BaseURL string) {
	samlAssertionDestBytes := client.GetDestinationByName(t, samlAssertionDestinationName, http.StatusOK)
	var samlAssertionDest esmdestinationcreator.SAMLAssertionDestination
	err := json.Unmarshal(samlAssertionDestBytes, &samlAssertionDest)
	require.NoError(t, err)
	require.Equal(t, samlAssertionDestinationName, samlAssertionDest.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, samlAssertionDest.Type)
	require.Equal(t, samlAssertionDestinationURL, samlAssertionDest.URL)
	require.Equal(t, directordestinationcreator.AuthTypeSAMLAssertion, samlAssertionDest.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, samlAssertionDest.ProxyType)
	require.Equal(t, app1BaseURL, samlAssertionDest.Audience)
	require.Equal(t, samlAssertionCertName+directordestinationcreator.JavaKeyStoreFileExtension, samlAssertionDest.KeyStoreLocation)
}

func AssertClientCertAuthDestination(t *testing.T, client *clients.DestinationClient, clientCertAuthDestinationName, clientCertAuthCertName, clientCertAuthDestinationURL string) {
	clientCertAuthDestBytes := client.GetDestinationByName(t, clientCertAuthDestinationName, http.StatusOK)
	var clientCertAuthDest esmdestinationcreator.ClientCertificateAuthenticationDestination
	err := json.Unmarshal(clientCertAuthDestBytes, &clientCertAuthDest)
	require.NoError(t, err)
	require.Equal(t, clientCertAuthDestinationName, clientCertAuthDest.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, clientCertAuthDest.Type)
	require.Equal(t, clientCertAuthDestinationURL, clientCertAuthDest.URL)
	require.Equal(t, directordestinationcreator.AuthTypeClientCertificate, clientCertAuthDest.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, clientCertAuthDest.ProxyType)
	require.Equal(t, clientCertAuthCertName+directordestinationcreator.JavaKeyStoreFileExtension, clientCertAuthDest.KeyStoreLocation)
}

func AssertDestinationCertificate(t *testing.T, client *clients.DestinationClient, certificateName string) {
	certBytes := client.GetDestinationCertificateByName(t, certificateName, http.StatusOK)
	var destCertificate esmdestinationcreator.DestinationSvcCertificateResponse
	err := json.Unmarshal(certBytes, &destCertificate)
	require.NoError(t, err)
	require.Equal(t, certificateName, destCertificate.Name)
	require.NotEmpty(t, destCertificate.Content)
}
