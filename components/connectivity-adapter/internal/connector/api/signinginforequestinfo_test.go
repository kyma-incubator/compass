package api

import (
	"encoding/json"
	"errors"
	mocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_SigningRequestInfo(t *testing.T) {

	t.Run("Should get Signing Request Info", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}

		newToken := "new_token"
		directorUrl := "www.director.com"
		certificateSecuredConnectorUrl := "www.connector.com"

		configurationResponse := schema.Configuration{
			Token: &schema.Token{
				Token: newToken,
			},
			CertificateSigningRequestInfo: &schema.CertificateSigningRequestInfo{
				Subject:      "O=Org,OU=OrgUnit,L=Gliwice,ST=Province,C=PL,CN=CommonName",
				KeyAlgorithm: "rsa2048",
			},
			ManagementPlaneInfo: &schema.ManagementPlaneInfo{
				DirectorURL:                    &directorUrl,
				CertificateSecuredConnectorURL: &certificateSecuredConnectorUrl,
			},
		}

		connectorClientMock.On("Configuration", "myapp").Return(configurationResponse, nil)
		handler := NewSigningRequestInfoHandler(connectorClientMock, "www.baseurl.com")

		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))
		req.Header.Set(oathkeeper.ClientIdFromTokenHeader, "myapp")

		r := httptest.NewRecorder()

		expectedSignUrl := "www.baseurl.com/v1/applications/certificates?token=new_token"
		expectedAPI := model.Api{
			RuntimeURLs: &model.RuntimeURLs{
				EventsURL:   "www.baseurl.com/myapp/v1/events",
				MetadataURL: "www.baseurl.com/myapp/v1/metadata",
			},
			InfoURL:         "www.baseurl.com/v1/applications/management/info",
			CertificatesURL: "www.baseurl.com/v1/applications/certificates",
		}

		expectedCertInfo := model.CertInfo{
			Subject:      "O=Org,OU=OrgUnit,L=Gliwice,ST=Province,C=PL,CN=CommonName",
			Extensions:   "",
			KeyAlgorithm: "rsa2048",
		}

		// when
		handler.GetSigningRequestInfo(r, req)
		defer closeResponseBody(t, r.Result())

		// then
		responseBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var infoResponse model.CSRInfoResponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, expectedSignUrl, infoResponse.CsrURL)
		assert.EqualValues(t, expectedAPI, infoResponse.API)
		assert.EqualValues(t, expectedCertInfo, infoResponse.CertificateInfo)
	})

	t.Run("Should return error when failed to call Compass Connector", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Configuration", "myapp").Return(schema.Configuration{}, errors.New("failed to execute graphql query"))

		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))
		req.Header.Set(oathkeeper.ClientIdFromTokenHeader, "myapp")

		r := httptest.NewRecorder()

		handler := NewSigningRequestInfoHandler(connectorClientMock, "www.baseurl.com")

		// when
		handler.GetSigningRequestInfo(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
		connectorClientMock.AssertExpectations(t)
	})

	// TODO: Check if it is needed (will be covered in middleware test)
	// TODO check what is the response from GraphQL client if the header is missing
	t.Run("Should return error when Client-Id-From-Token not passed", func(t *testing.T) {
		// given

		// when

		// then
	})
}

func closeResponseBody(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}
