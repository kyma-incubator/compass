package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerRevocations(t *testing.T) {
	headersFromCertificate := map[string]string{
		oathkeeper.ClientIdFromCertificateHeader: "myapp",
	}

	t.Run("Should revoke certificate", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Revoke", headersFromCertificate).Return(nil)
		handler := NewRevocationsHandler(connectorClientMock, logrus.New())

		req := newRequestWithContext(strings.NewReader(""), headersFromCertificate, nil)

		r := httptest.NewRecorder()

		// when
		handler.RevokeCertificate(r, req)

		// then
		require.Equal(t, http.StatusCreated, r.Code)
	})

	t.Run("Should return error when failed to call Connector", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Revoke", headersFromCertificate).Return(apperrors.Internal("error"))
		handler := NewRevocationsHandler(connectorClientMock, logrus.New())

		req := newRequestWithContext(strings.NewReader(""), headersFromCertificate, nil)

		r := httptest.NewRecorder()

		// when
		handler.RevokeCertificate(r, req)

		// then
		require.Equal(t, http.StatusInternalServerError, r.Code)
	})

	t.Run("Should return error when Authorization context not passed", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}

		r := httptest.NewRecorder()
		req := newRequestWithContext(strings.NewReader(""), nil, nil)
		handler := NewRevocationsHandler(connectorClientMock, logrus.New())

		// when
		handler.RevokeCertificate(r, req)

		// then
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

}
