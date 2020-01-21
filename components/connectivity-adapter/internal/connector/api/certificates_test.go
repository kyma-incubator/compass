package api

//
//import (
//	"bytes"
//	"encoding/json"
//	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api/middlewares"
//	mocks "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql/automock"
//	"net/http"
//	"net/http/httptest"
//	"strings"
//	"testing"
//)
//
//func TestHandler_Certificates(t *testing.T) {
//
//	clientId := "myapp"
//	baseURLs := middlewares.BaseURLs{
//		ConnectivityAdapterBaseURL: "www.connectivity-adapter.com",
//		EventServiceBaseURL:        "www.event-service.com",
//	}
//	signatureRequestRaw := compact([]byte("{\"csr\":\"Q1NSCg==\"}"))
//
//	t.Run("Should sign certificate", func(t *testing.T) {
//		// given
//		connectorClientMock := &mocks.Client{}
//		// func (c client) SignCSR(csr string, clientID string) (schema.CertificationResult, error) {
//		connectorClientMock.On("SignCSR")
//
//
//		handler := NewCertificatesHandler(connectorClientMock)
//		req := newRequestWithContext(&clientId, nil)
//
//		r := httptest.NewRecorder()
//
//		// when
//		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", bytes.NewReader(signatureRequestRaw))
//
//		// then
//
//
//	})
//
//	t.Run("Should return error when failed to call Compass Connector", func(t *testing.T) {
//
//	})
//}
//
//func compact(src []byte) []byte {
//	buffer := new(bytes.Buffer)
//	err := json.Compact(buffer, src)
//	if err != nil {
//		return src
//	}
//	return buffer.Bytes()
//}
