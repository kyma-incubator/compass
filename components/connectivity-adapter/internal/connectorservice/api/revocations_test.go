package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlerRevocations(t *testing.T) {
	headersFromCertificate := map[string]string{
		oathkeeper.ClientIdFromCertificateHeader: "myapp",
	}

	t.Run("Should revoke certificate", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Revoke", mock.Anything, headersFromCertificate).Return(nil)

		connectorClientProviderMock := &mocks.ClientProvider{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		handler := NewRevocationsHandler(connectorClientProviderMock)

		req := newRequestWithContext(strings.NewReader(""), headersFromCertificate)

		r := httptest.NewRecorder()

		// when
		handler.RevokeCertificate(r, req)

		// then
		require.Equal(t, http.StatusCreated, r.Code)
	})

	t.Run("Should return error when failed to call Connector", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Revoke", mock.Anything, headersFromCertificate).Return(apperrors.Internal("error"))

		connectorClientProviderMock := &mocks.ClientProvider{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		handler := NewRevocationsHandler(connectorClientProviderMock)

		req := newRequestWithContext(strings.NewReader(""), headersFromCertificate)

		r := httptest.NewRecorder()

		// when
		handler.RevokeCertificate(r, req)

		// then
		require.Equal(t, http.StatusInternalServerError, r.Code)
	})

	t.Run("Should return error when Authorization context not passed", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}

		connectorClientProviderMock := &mocks.ClientProvider{}
		connectorClientProviderMock.On("Client", mock.AnythingOfType("*http.Request")).Return(connectorClientMock)

		r := httptest.NewRecorder()
		req := newRequestWithContext(strings.NewReader(""), nil)
		handler := NewRevocationsHandler(connectorClientProviderMock)

		// when
		handler.RevokeCertificate(r, req)

		// then
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

}
