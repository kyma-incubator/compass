package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func TestHandler_Certificates(t *testing.T) {

	baseURLs := middlewares.BaseURLs{
		ConnectivityAdapterBaseURL: "www.connectivity-adapter.com",
		EventServiceBaseURL:        "www.event-service.com",
	}
	headersFromToken := map[string]string{
		oathkeeper.ClientIdFromTokenHeader: "myapp",
	}
	signatureRequestRaw := compact([]byte("{\"csr\":\"Q1NSCg==\"}"))

	t.Run("Should sign certificate", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("SignCSR", "Q1NSCg==", headersFromToken).Return(schema.CertificationResult{
			CaCertificate:     "ca_cert",
			CertificateChain:  "cert_chain",
			ClientCertificate: "client_cert",
		}, nil)

		handler := NewCertificatesHandler(connectorClientMock, logrus.New())
		req := newRequestWithContext(bytes.NewReader(signatureRequestRaw), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		// when
		handler.SignCSR(r, req)

		// then
		assert.Equal(t, http.StatusCreated, r.Code)

		responseBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var certResponse model.CertResponse
		err = json.Unmarshal(responseBody, &certResponse)
		require.NoError(t, err)

		assert.Equal(t, "ca_cert", certResponse.CaCRT)
		assert.Equal(t, "client_cert", certResponse.ClientCRT)
		assert.Equal(t, "cert_chain", certResponse.CRTChain)
	})

	t.Run("Should return error when failed to call Compass Connector", func(t *testing.T) {
		// given
		connectorClientMock := &mocks.Client{}
		connectorClientMock.On("SignCSR", "Q1NSCg==", headersFromToken).
			Return(schema.CertificationResult{}, errors.New("some error"))

		handler := NewCertificatesHandler(connectorClientMock, logrus.New())
		req := newRequestWithContext(bytes.NewReader(signatureRequestRaw), headersFromToken, &baseURLs)

		r := httptest.NewRecorder()

		// when
		handler.SignCSR(r, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
