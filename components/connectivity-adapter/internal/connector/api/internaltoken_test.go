package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/stretchr/testify/assert"

	mocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql/automock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandler_InternalToken(t *testing.T) {
	t.Run("Should get token", func(t *testing.T) {
		// given
		application := "myapp"
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Token", application).Return("token", nil)

		handler := NewTokenHandler(connectorClientMock, "www.connectivity-adapter.com", logrus.New())

		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))
		req.Header.Set(ApplicationHeader, application)
		r := httptest.NewRecorder()

		// when
		handler.GetToken(r, req)

		// then
		responseBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var tokenResponse model.TokenResponse
		err = json.Unmarshal(responseBody, &tokenResponse)
		require.NoError(t, err)

		require.Equal(t, http.StatusCreated, r.Code)
		assert.Equal(t, "token", tokenResponse.Token)
		assert.Equal(t, "www.connectivity-adapter.com/v1/applications/signingRequests/info?token=token", tokenResponse.URL)
	})

	t.Run("Should return error when failed to get token", func(t *testing.T) {
		// given
		application := "myapp"
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Token", application).Return("", apperrors.Internal("error"))

		handler := NewTokenHandler(connectorClientMock, "www.connectivity-adapter.com", logrus.New())

		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))
		req.Header.Set(ApplicationHeader, application)
		r := httptest.NewRecorder()

		// when
		handler.GetToken(r, req)

		// then
		require.Equal(t, http.StatusInternalServerError, r.Code)
	})

	t.Run("Should return error when Application Header not passed", func(t *testing.T) {
		// given
		application := "myapp"
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("Token", application).Return("", apperrors.Internal("error"))

		handler := NewTokenHandler(connectorClientMock, "www.connectivity-adapter.com", logrus.New())

		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))

		r := httptest.NewRecorder()

		// when
		handler.GetToken(r, req)

		// then
		require.Equal(t, http.StatusBadRequest, r.Code)
	})
}
