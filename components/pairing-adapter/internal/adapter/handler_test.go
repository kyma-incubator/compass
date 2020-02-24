package adapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/pairing-adapter/internal/adapter"
	"github.com/kyma-incubator/compass/components/pairing-adapter/internal/adapter/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		mockClient := &automock.Client{}
		defer mockClient.AssertExpectations(t)
		givenToken := &adapter.ExternalToken{
			Token: "some-token",
		}
		mockClient.On("Do", mock.Anything, givenReqData()).Return(givenToken, nil)

		sut := adapter.NewHandler(mockClient)
		// WHEN
		// THEN
		rr := httptest.NewRecorder()

		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(givenReqData())
		require.NoError(t, err)
		givenReq, err := http.NewRequest(http.MethodPost, "", buf)
		require.NoError(t, err)
		sut.ServeHTTP(rr, givenReq)
		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

		actualBody := rr.Result().Body
		defer func() {
			require.NoError(t, actualBody.Close())
		}()

		actualToken := adapter.ExternalToken{}
		err = json.NewDecoder(actualBody).Decode(&actualToken)
		require.NoError(t, err)
		assert.Equal(t, *givenToken, actualToken)
	})

	t.Run("errors on wrong input data", func(t *testing.T) {
		// GIVEN
		sut := adapter.NewHandler(nil)
		// WHEN
		// THEN
		rr := httptest.NewRecorder()

		buf := new(bytes.Buffer)
		_, err := buf.WriteString(`<xml></xml>`)
		require.NoError(t, err)
		givenReq, err := http.NewRequest(http.MethodPost, "", buf)
		require.NoError(t, err)
		sut.ServeHTTP(rr, givenReq)
		assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)

	})

	t.Run("error on calling external service", func(t *testing.T) {
		// GIVEN
		mockClient := &automock.Client{}
		defer mockClient.AssertExpectations(t)
		mockClient.On("Do", mock.Anything, givenReqData()).Return(nil, errors.New("some error"))

		sut := adapter.NewHandler(mockClient)
		rr := httptest.NewRecorder()

		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(givenReqData())
		require.NoError(t, err)
		givenReq, err := http.NewRequest(http.MethodPost, "", buf)
		require.NoError(t, err)
		// WHEN
		sut.ServeHTTP(rr, givenReq)
		// THEN
		assert.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
	})

}

func givenReqData() adapter.RequestData {
	return adapter.RequestData{}
}
