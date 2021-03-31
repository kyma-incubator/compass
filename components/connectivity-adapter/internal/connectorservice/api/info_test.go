package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	connectorMock "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector/automock"
	directorMock "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/director/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"
	connectorSchema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_Info(t *testing.T) {

	headersFromToken := map[string]string{
		oathkeeper.ClientIdFromTokenHeader: "myapp",
	}

	appName := "myappname"
	normalizedAppName := "mp-myappname"
	application := graphql.ApplicationExt{
		Application:           graphql.Application{Name: appName},
		EventingConfiguration: graphql.ApplicationEventingConfiguration{DefaultURL: "https://default-events-url.com"},
	}

	applicationWithLabels := graphql.ApplicationExt{
		Application:           application.Application,
		EventingConfiguration: application.EventingConfiguration,
		Labels: graphql.Labels{
			appNameLabel: normalizedAppName,
		},
	}

	emptyResFunction := func(applicationName string, eventServiceBaseURL, tenant string, configuration connectorSchema.Configuration) (i interface{}, e error) {
		return connectorSchema.Configuration{}, nil
	}

	t.Run("Should get Signing Request Info for application without name label", func(t *testing.T) {
		// given
		connectorClientMock := &connectorMock.Client{}

		connectorClientProviderMock := &connectorMock.ClientProvider{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		directorClientProviderMock := &directorMock.ClientProvider{}
		directorClientMock := &directorMock.Client{}
		directorClientMock.On("GetApplication", mock.Anything, "myapp").Return(application, nil)
		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)

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

		connectorClientMock.On("Configuration", mock.Anything, headersFromToken).Return(configurationResponse, nil)
		csrInfoHandler := NewInfoHandler(connectorClientProviderMock, directorClientProviderMock, model.NewCSRInfoResponseProvider("www.connectivity-adapter.com", "www.connectivity-adapter-mtls.com"))

		req := newRequestWithContext(strings.NewReader(""), headersFromToken)

		r := httptest.NewRecorder()

		expectedInfoResponse := model.CSRInfoResponse{
			CsrURL: "www.connectivity-adapter.com/v1/applications/certificates?token=new_token",
			API: model.Api{
				RuntimeURLs: &model.RuntimeURLs{
					EventsURL:     "https://default-events-url.com",
					EventsInfoURL: "https://default-events-url.com/subscribed",
					MetadataURL:   fmt.Sprintf("www.connectivity-adapter-mtls.com/%s/v1/metadata/services", appName),
				},
				InfoURL:         "www.connectivity-adapter-mtls.com/v1/applications/management/info",
				CertificatesURL: "www.connectivity-adapter.com/v1/applications/certificates",
			},
			CertificateInfo: model.CertInfo{
				Subject:      "O=Org,OU=OrgUnit,L=Gliwice,ST=Province,C=PL,CN=CommonName",
				Extensions:   "",
				KeyAlgorithm: "rsa2048",
			},
		}

		// when
		csrInfoHandler.GetInfo(r, req)
		defer closeResponseBody(t, r.Result())

		// then
		responseBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var infoResponse model.CSRInfoResponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, r.Code)
		assert.EqualValues(t, expectedInfoResponse, infoResponse)
	})

	t.Run("Should get Signing Request Info for application with name label", func(t *testing.T) {
		// given
		connectorClientMock := &connectorMock.Client{}

		connectorClientProviderMock := &connectorMock.ClientProvider{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		directorClientProviderMock := &directorMock.ClientProvider{}
		directorClientMock := &directorMock.Client{}
		directorClientMock.On("GetApplication", mock.Anything, "myapp").Return(applicationWithLabels, nil)
		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)

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

		connectorClientMock.On("Configuration", mock.Anything, headersFromToken).Return(configurationResponse, nil)
		csrInfoHandler := NewInfoHandler(connectorClientProviderMock, directorClientProviderMock, model.NewCSRInfoResponseProvider("www.connectivity-adapter.com", "www.connectivity-adapter-mtls.com"))

		req := newRequestWithContext(strings.NewReader(""), headersFromToken)

		r := httptest.NewRecorder()

		expectedInfoResponse := model.CSRInfoResponse{
			CsrURL: "www.connectivity-adapter.com/v1/applications/certificates?token=new_token",
			API: model.Api{
				RuntimeURLs: &model.RuntimeURLs{
					EventsURL:     "https://default-events-url.com",
					EventsInfoURL: "https://default-events-url.com/subscribed",
					MetadataURL:   fmt.Sprintf("www.connectivity-adapter-mtls.com/%s/v1/metadata/services", normalizedAppName),
				},
				InfoURL:         "www.connectivity-adapter-mtls.com/v1/applications/management/info",
				CertificatesURL: "www.connectivity-adapter.com/v1/applications/certificates",
			},
			CertificateInfo: model.CertInfo{
				Subject:      "O=Org,OU=OrgUnit,L=Gliwice,ST=Province,C=PL,CN=CommonName",
				Extensions:   "",
				KeyAlgorithm: "rsa2048",
			},
		}

		// when
		csrInfoHandler.GetInfo(r, req)
		defer closeResponseBody(t, r.Result())

		// then
		responseBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var infoResponse model.CSRInfoResponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, r.Code)
		assert.EqualValues(t, expectedInfoResponse, infoResponse)
	})

	t.Run("Should get Management Info for application without name label", func(t *testing.T) {
		// given
		connectorClientProviderMock := &connectorMock.ClientProvider{}
		directorClientProviderMock := &directorMock.ClientProvider{}
		connectorClientMock := &connectorMock.Client{}
		directorClientMock := &directorMock.Client{}

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

		appName := "myApp"
		directorApp := directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name: appName,
			},
			EventingConfiguration: directorSchema.ApplicationEventingConfiguration{
				DefaultURL: "www.event-service.com/myApp/v1/events",
			},
		}

		directorClientMock.On("GetApplication", mock.Anything, "myapp").Return(directorApp, nil)

		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		connectorClientMock.On("Configuration", mock.Anything, headersFromToken).Return(configurationResponse, nil)
		mgmtInfoHandler := NewInfoHandler(connectorClientProviderMock, directorClientProviderMock, model.NewManagementInfoResponseProvider("www.connectivity-adapter-mtls.com"))

		req := newRequestWithContext(strings.NewReader(""), headersFromToken)

		r := httptest.NewRecorder()

		expectedManagementInfoResponse := model.MgmtInfoReponse{
			ClientIdentity: model.ClientIdentity{
				Application: appName,
			},
			URLs: model.MgmtURLs{
				RuntimeURLs: &model.RuntimeURLs{
					EventsURL:     "www.event-service.com/myApp/v1/events",
					EventsInfoURL: "www.event-service.com/myApp/v1/events/subscribed",
					MetadataURL:   fmt.Sprintf("www.connectivity-adapter-mtls.com/%s/v1/metadata/services", appName),
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
		mgmtInfoHandler.GetInfo(r, req)
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

	t.Run("Should get Management Info for application with name label", func(t *testing.T) {
		// given
		connectorClientProviderMock := &connectorMock.ClientProvider{}
		directorClientProviderMock := &directorMock.ClientProvider{}
		connectorClientMock := &connectorMock.Client{}
		directorClientMock := &directorMock.Client{}

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

		appName := "myApp"
		normalizedAppName := "mp-myapp"
		directorApp := directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name: appName,
			},
			EventingConfiguration: directorSchema.ApplicationEventingConfiguration{
				DefaultURL: "www.event-service.com/myApp/v1/events",
			},
			Labels: graphql.Labels{
				appNameLabel: normalizedAppName,
			},
		}

		directorClientMock.On("GetApplication", mock.Anything, "myapp").Return(directorApp, nil)

		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		connectorClientMock.On("Configuration", mock.Anything, headersFromToken).Return(configurationResponse, nil)
		mgmtInfoHandler := NewInfoHandler(connectorClientProviderMock, directorClientProviderMock, model.NewManagementInfoResponseProvider("www.connectivity-adapter-mtls.com"))

		req := newRequestWithContext(strings.NewReader(""), headersFromToken)

		r := httptest.NewRecorder()

		expectedManagementInfoResponse := model.MgmtInfoReponse{
			ClientIdentity: model.ClientIdentity{
				Application: normalizedAppName,
			},
			URLs: model.MgmtURLs{
				RuntimeURLs: &model.RuntimeURLs{
					EventsURL:     "www.event-service.com/myApp/v1/events",
					EventsInfoURL: "www.event-service.com/myApp/v1/events/subscribed",
					MetadataURL:   fmt.Sprintf("www.connectivity-adapter-mtls.com/%s/v1/metadata/services", normalizedAppName),
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
		mgmtInfoHandler.GetInfo(r, req)
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
		directorClientProviderMock := &directorMock.ClientProvider{}
		connectorClientProviderMock := &connectorMock.ClientProvider{}
		directorClientMock := &directorMock.Client{}
		directorClientMock.On("GetApplication", mock.Anything, mock.AnythingOfType("string")).Return(graphql.ApplicationExt{}, apperrors.Internal("error"))
		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)

		connectorClientMock := &connectorMock.Client{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

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
		connectorClientMock.On("Configuration", mock.Anything, headersFromToken).Return(configurationResponse, nil)

		req := newRequestWithContext(strings.NewReader(""), headersFromToken)
		r := httptest.NewRecorder()

		handler := NewInfoHandler(connectorClientProviderMock, directorClientProviderMock, emptyResFunction)

		// when
		handler.GetInfo(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
		directorClientMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to call Compass Connector", func(t *testing.T) {
		// given
		connectorClientMock := &connectorMock.Client{}
		connectorClientMock.On("Configuration", mock.Anything, headersFromToken).Return(connectorSchema.Configuration{}, apperrors.Internal("error"))
		connectorClientProviderMock := &connectorMock.ClientProvider{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		directorClientProviderMock := &directorMock.ClientProvider{}
		directorClientMock := &directorMock.Client{}
		directorClientMock.On("GetApplication", mock.Anything, mock.AnythingOfType("string")).Return(graphql.ApplicationExt{}, nil)
		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)

		req := newRequestWithContext(strings.NewReader(""), headersFromToken)

		r := httptest.NewRecorder()

		handler := NewInfoHandler(connectorClientProviderMock, directorClientProviderMock, emptyResFunction)

		// when
		handler.GetInfo(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
		connectorClientMock.AssertExpectations(t)
	})

	t.Run("Should return error when Authorization context not passed", func(t *testing.T) {
		// given
		connectorClientMock := &connectorMock.Client{}
		directorClientProviderMock := &directorMock.ClientProvider{}
		connectorClientProviderMock := &connectorMock.ClientProvider{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		r := httptest.NewRecorder()
		req := newRequestWithContext(strings.NewReader(""), nil)

		handler := NewInfoHandler(connectorClientProviderMock, directorClientProviderMock, emptyResFunction)

		// when
		handler.GetInfo(r, req)

		// then
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	t.Run("Should return error when failed to build response", func(t *testing.T) {
		// given
		connectorClientMock := &connectorMock.Client{}
		connectorClientMock.On("Configuration", mock.Anything, headersFromToken).Return(connectorSchema.Configuration{}, apperrors.Internal("error"))
		connectorClientProviderMock := &connectorMock.ClientProvider{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		directorClientProviderMock := &directorMock.ClientProvider{}
		directorClientMock := &directorMock.Client{}
		directorClientMock.On("GetApplication", mock.Anything, mock.AnythingOfType("string")).Return(graphql.ApplicationExt{}, nil)
		directorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(directorClientMock)

		req := newRequestWithContext(strings.NewReader(""), headersFromToken)

		r := httptest.NewRecorder()

		errorResFunction := func(applicationName string, eventServiceBaseURL, tenant string, configuration connectorSchema.Configuration) (i interface{}, e error) {
			return connectorSchema.Configuration{}, errors.New("some error")
		}

		handler := NewInfoHandler(connectorClientProviderMock, directorClientProviderMock, errorResFunction)

		// when
		handler.GetInfo(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
		connectorClientMock.AssertExpectations(t)
	})
}

func newRequestWithContext(body io.Reader, headers map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", body)

	newContext := req.Context()

	if headers != nil {
		newContext = middlewares.PutIntoContext(newContext, middlewares.AuthorizationHeadersKey, middlewares.AuthorizationHeaders(headers))
	}

	return req.WithContext(newContext)
}

func closeResponseBody(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}
