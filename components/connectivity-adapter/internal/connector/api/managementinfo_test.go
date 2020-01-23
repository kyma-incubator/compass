package api

import (
	"encoding/json"
	"errors"
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

func TestHandlerManagementInfo(t *testing.T) {

	baseURLs := middlewares.BaseURLs{
		ConnectivityAdapterBaseURL: "www.connectivity-adapter.com",
		EventServiceBaseURL:        "www.event-service.com",
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
		handler := NewManagementInfoHandler(connectorClientMock, logrus.New())

		req := newRequestWithContext(strings.NewReader(""), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		expectedURLs := model.MgmtURLs{
			RuntimeURLs: &model.RuntimeURLs{
				EventsURL:   "www.event-service.com/myapp/v1/events",
				MetadataURL: "www.connectivity-adapter.com/myapp/v1/metadata",
			},
			RenewCertURL:  "www.connectivity-adapter.com/applications/certificates/renewals",
			RevokeCertURL: "www.connectivity-adapter.com/applications/certificates/revocations",
		}

		expectedClientIdentify := model.ClientIdentity{
			Application: "myapp",
		}

		expectedCertInfo := model.CertInfo{
			Subject:      "O=Org,OU=OrgUnit,L=Gliwice,ST=Province,C=PL,CN=CommonName",
			Extensions:   "",
			KeyAlgorithm: "rsa2048",
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

		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, expectedURLs, managementInfoResponse.URLs)
		assert.EqualValues(t, expectedClientIdentify, managementInfoResponse.ClientIdentity)
		assert.EqualValues(t, expectedCertInfo, managementInfoResponse.CertificateInfo)
	})

	t.Run("Should return error when failed to call Compass Connector", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Configuration", headersFromToken).Return(schema.Configuration{}, errors.New("failed to execute graphql query"))

		req := newRequestWithContext(strings.NewReader(""), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		handler := NewManagementInfoHandler(connectorClientMock, logrus.New())

		// when
		handler.GetManagementInfo(r, req)

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

			handler := NewManagementInfoHandler(connectorClientMock, logrus.New())

			// when
			handler.GetManagementInfo(r, tc.request)

			// then
			assert.Equal(t, http.StatusInternalServerError, r.Code)
			connectorClientMock.AssertNotCalled(t, "Configuration")
		})
	}
}
