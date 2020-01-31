package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	mocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector/automock"
	directorMocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/director/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/model"
	connectorSchema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerManagementInfo(t *testing.T) {

	baseURLs := middlewares.BaseURLs{
		ConnectivityAdapterBaseURL:     "www.connectivity-adapter.com",
		ConnectivityAdapterMTLSBaseURL: "www.connectivity-adapter-mtls.com",
		EventServiceBaseURL:            "www.event-service.com",
	}

	headersFromToken := map[string]string{
		oathkeeper.ClientIdFromTokenHeader: "systemAuthID",
	}

	t.Run("Should get Signing Request Info", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		directorClientProviderMock := &directorMocks.ClientProvider{}
		directorClientMock := &directorMocks.Client{}

		newToken := "new_token"
		directorUrl := "www.director.com"
		certificateSecuredConnectorUrl := "www.connector.com"

		configurationResponse := connectorSchema.Configuration{
			Token: &connectorSchema.Token{
				Token: newToken,
			},
			CertificateSigningRequestInfo: &connectorSchema.CertificateSigningRequestInfo{
				Subject:      "O=Org,OU=OrgUnit,L=Gliwice,ST=Province,C=PL,CN=CommonName",
				KeyAlgorithm: "rsa2048",
			},
			ManagementPlaneInfo: &connectorSchema.ManagementPlaneInfo{
				DirectorURL:                    &directorUrl,
				CertificateSecuredConnectorURL: &certificateSecuredConnectorUrl,
			},
		}

		directorApp := directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name: "myApp",
			},
			EventingConfiguration: directorSchema.ApplicationEventingConfiguration{
				DefaultURL: "www.event-service.com",
			},
		}

		directorClientMock.On("GetApplication", "systemAuthID").Return(directorApp, nil)

		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)

		connectorClientMock.On("Configuration", headersFromToken).Return(configurationResponse, nil)
		handler := NewManagementInfoHandler(connectorClientMock, logrus.New(), "www.connectivity-adapter-mtls.com", directorClientProviderMock)

		req := newRequestWithContext(strings.NewReader(""), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		expectedManagementInfoResponse := model.MgmtInfoReponse{
			ClientIdentity: model.ClientIdentity{
				Application: "myApp",
			},
			URLs: model.MgmtURLs{
				RuntimeURLs: &model.RuntimeURLs{
					EventsURL:     "www.event-service.com/myApp/v1/events",
					EventsInfoURL: "www.event-service.com/myApp/v1/events/subscribed",
					MetadataURL:   "www.connectivity-adapter-mtls.com/myApp/v1/metadata/services",
				},
				RenewCertURL:  "www.connectivity-adapter-mtls.com/v1/applications/certificates/renewals",
				RevokeCertURL: "www.connectivity-adapter-mtls.com/v1/applications/certificates/revocations",
			},
			CertificateInfo: model.CertInfo{
				Subject:      "O=Org,OU=OrgUnit,L=Gliwice,ST=Province,C=PL,CN=CommonName",
				Extensions:   "",
				KeyAlgorithm: "rsa2048",
			},
		}

		// when
		handler.GetManagementInfo(r, req)
		defer closeResponseBody(t, r.Result())

		// then
		responseBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var managementInfoResponse model.MgmtInfoReponse
		err = json.Unmarshal(responseBody, &managementInfoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, expectedManagementInfoResponse, managementInfoResponse)
	})

	t.Run("Should return error when failed to call Compass Director", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		directorClientProviderMock := &directorMocks.ClientProvider{}
		directorClientMock := &directorMocks.Client{}

		directorClientMock.On("GetApplication", "systemAuthID").Return(directorSchema.ApplicationExt{}, apperrors.Internal("error"))

		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)

		req := newRequestWithContext(strings.NewReader(""), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		handler := NewManagementInfoHandler(connectorClientMock, logrus.New(), "", directorClientProviderMock)

		// when
		handler.GetManagementInfo(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
		connectorClientMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to call Compass Connector", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		directorClientProviderMock := &directorMocks.ClientProvider{}
		directorClientMock := &directorMocks.Client{}

		directorClientMock.On("GetApplication", "systemAuthID").Return(directorSchema.ApplicationExt{}, nil)

		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)

		connectorClientMock.On("Configuration", headersFromToken).Return(connectorSchema.Configuration{}, apperrors.Internal("error"))

		req := newRequestWithContext(strings.NewReader(""), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		handler := NewManagementInfoHandler(connectorClientMock, logrus.New(), "", directorClientProviderMock)

		// when
		handler.GetManagementInfo(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
		connectorClientMock.AssertExpectations(t)
	})

	t.Run("Should return error when Authorization context not passed", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		directorClientProviderMock := &directorMocks.ClientProvider{}

		r := httptest.NewRecorder()
		req := newRequestWithContext(strings.NewReader(""), nil, nil)
		handler := NewManagementInfoHandler(connectorClientMock, logrus.New(), "", directorClientProviderMock)

		// when
		handler.GetManagementInfo(r, req)

		// then
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
