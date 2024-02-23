package destinationfetchersvc_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	destination1Name    = "mys4_1"
	exampleDestination1 = `{
        "Name": "mys4_1",
        "Type": "HTTP",
        "URL": "https://my54321-api.s4.com:443",
        "Authentication": "BasicAuthentication",
        "ProxyType": "Internet",
        "XFSystemName": "Test S4HANA system",
        "HTML5.DynamicDestination": "true",
        "User": "SOME_USER",
        "product.name": "SAP S/4HANA Cloud",
        "WebIDEEnabled": "true",
        "communicationScenarioId": "SAP_COM_0108",
        "Password": "SecretPassword",
        "WebIDEUsage": "odata_gen"
    }`
	exampleDestination2 = `{
        "Name": "mysystem_2",
        "Type": "HTTP",
        "URL": "https://mysystem.com",
        "Authentication": "BasicAuthentication",
        "ProxyType": "Internet",
        "HTML5.DynamicDestination": "true",
        "User": "SOME_USER",
        "Password": "SecretPassword",
		"x-system-id": "system-id",
		"x-system-type": "mysystem"
    }`
)

type destinationHandler struct {
	customTenantDestinationHandler func(w http.ResponseWriter, req *http.Request)
	t                              *testing.T
}

func (dh *destinationHandler) mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/subaccountDestinations", dh.tenantDestinationHandler)
	mux.HandleFunc("/destinations/", dh.fetchDestinationHandler)
	mux.HandleFunc("/oauth/token", dh.tokenHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dh.t.Logf("Unhandled request to mocked destination service %s", r.URL.String())
		w.WriteHeader(http.StatusInternalServerError)
	})
	return mux
}

func (dh *destinationHandler) tenantDestinationHandler(w http.ResponseWriter, req *http.Request) {
	if dh.customTenantDestinationHandler != nil {
		dh.customTenantDestinationHandler(w, req)
		return
	}
	query := req.URL.Query()
	page := query.Get("$page")

	w.Header().Set("Content-Type", "application/json")
	// assuming pageSize is always 1
	var response []byte
	switch page {
	case "1":
		response = []byte(fmt.Sprintf("[%s]", exampleDestination1))
	case "2":
		response = []byte(fmt.Sprintf("[%s]", exampleDestination2))
	default:
		dh.t.Logf("Expected page size to be 1 or 2, got '%s'", page)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Page-Count", "2")
	_, err := w.Write(response)
	assert.NoError(dh.t, err)
}

var defaultDestinations = map[string]string{
	"dest1": `{"name": "dest1", "destinationConfiguration": {}}`,
	"dest2": `{"name": "dest2", "destinationConfiguration": {}}`,
}

func (dh *destinationHandler) fetchDestinationHandler(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	w.Header().Set("Content-Type", "application/json")
	for destinationName, destination := range defaultDestinations {
		if strings.HasSuffix(path, "/"+destinationName) {
			_, err := w.Write([]byte(destination))
			assert.NoError(dh.t, err)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (dh *destinationHandler) tokenHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(`{
			"access_token": "accesstoken",
			"token_type": "tokentype",
			"refresh_token": "refreshtoken",
			"expires_in": 100
		}`))
	assert.NoError(dh.t, err)
}

type destinationServer struct {
	server  *httptest.Server
	handler *destinationHandler
}

func newDestinationServer(t *testing.T) destinationServer {
	destinationHandler := &destinationHandler{t: t}
	httpServer := httptest.NewUnstartedServer(destinationHandler.mux())
	var err error
	httpServer.Listener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return destinationServer{
		server:  httpServer,
		handler: destinationHandler,
	}
}
