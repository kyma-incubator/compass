package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api/middlewares"
	mocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_SigningRequestInfo(t *testing.T) {

	baseURLs := middlewares.BaseURLs{
		ConnectivityAdapterBaseURL:     "www.connectivity-adapter.com",
		ConnectivityAdapterMTLSBaseURL: "www.connectivity-adapter-mtls.com",
		EventServiceBaseURL:            "www.event-service.com",
	}

	headersFromToken := map[string]string{
		oathkeeper.ClientIdFromTokenHeader: "myapp",
	}

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

		connectorClientMock.On("Configuration", headersFromToken).Return(configurationResponse, nil)
		handler := NewSigningRequestInfoHandler(connectorClientMock, logrus.New())

		req := newRequestWithContext(strings.NewReader(""), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		expectedInfoResponse := model.CSRInfoResponse{
			CsrURL: "www.connectivity-adapter.com/v1/applications/certificates?token=new_token",
			API: model.Api{
				RuntimeURLs: &model.RuntimeURLs{
					EventsURL:   "www.event-service.com/myapp/v1/events",
					MetadataURL: "www.connectivity-adapter-mtls.com/myapp/v1/metadata",
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
		assert.EqualValues(t, expectedInfoResponse, infoResponse)
	})

	t.Run("Should return error when failed to call Compass Connector", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Configuration", headersFromToken).Return(schema.Configuration{}, errors.New("failed to execute graphql query"))

		req := newRequestWithContext(strings.NewReader(""), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		handler := NewSigningRequestInfoHandler(connectorClientMock, logrus.New())

		// when
		handler.GetSigningRequestInfo(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
		connectorClientMock.AssertExpectations(t)
	})

	testCases := []struct {
		description string
		request     *http.Request
	}{
		{
			"Should return error when Authorization context not passed",
			newRequestWithContext(strings.NewReader(""), nil, nil),
		},
		{
			"Should return error when Base URL context not passed",
			newRequestWithContext(strings.NewReader(""), headersFromToken, nil),
		},
	}

	for _, tc := range testCases {
		t.Run("Should return error when Authorization context not passed", func(t *testing.T) {
			// given
			connectorClientMock := &mocks.Client{}

			r := httptest.NewRecorder()

			handler := NewSigningRequestInfoHandler(connectorClientMock, logrus.New())

			// when
			handler.GetSigningRequestInfo(r, tc.request)

			// then
			assert.Equal(t, http.StatusInternalServerError, r.Code)
			connectorClientMock.AssertNotCalled(t, "Configuration")
		})
	}
}

func newRequestWithContext(body io.Reader, headers map[string]string, baseURLs *middlewares.BaseURLs) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", body)

	newContext := req.Context()

	if baseURLs != nil {
		newContext = middlewares.PutIntoContext(newContext, middlewares.BaseURLsKey, *baseURLs)
	}

	if headers != nil {
		newContext = middlewares.PutIntoContext(newContext, middlewares.AuthorizationHeadersKey, middlewares.AuthorizationHeaders(headers))
	}

	return req.WithContext(newContext)
}

func closeResponseBody(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}
