package edp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/logger"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	subAccountID  = "72b83910-c2dc-415b-b95d-960cc45b36abx"
	environment   = "test"
	testNamespace = "testNs"
)

func TestClient_CreateDataTenant(t *testing.T) {
	// given
	testServer := fixHTTPServer(t)
	defer testServer.Close()

	config := Config{
		AdminURL:  testServer.URL,
		Namespace: testNamespace,
	}
	client := NewClient(config, logger.NewLogDummy())
	client.setHttpClient(testServer.Client())

	// when
	err := client.CreateDataTenant(DataTenantPayload{
		Name:        subAccountID,
		Environment: environment,
	})

	// then
	assert.NoError(t, err)

	response, err := testServer.Client().Get(fmt.Sprintf("%s/namespaces/%s/dataTenants/%s/%s", testServer.URL, testNamespace, subAccountID, environment))
	assert.NoError(t, err)

	var dt DataTenantItem
	err = json.NewDecoder(response.Body).Decode(&dt)
	assert.NoError(t, err)
	assert.Equal(t, subAccountID, dt.Name)
	assert.Equal(t, environment, dt.Environment)
	assert.Equal(t, testNamespace, dt.Namespace.Name)
}

func TestClient_DeleteDataTenant(t *testing.T) {
	// given
	testServer := fixHTTPServer(t)
	defer testServer.Close()

	config := Config{
		AdminURL:  testServer.URL,
		Namespace: testNamespace,
	}
	client := NewClient(config, logger.NewLogDummy())
	client.setHttpClient(testServer.Client())

	err := client.CreateDataTenant(DataTenantPayload{
		Name:        subAccountID,
		Environment: environment,
	})
	assert.NoError(t, err)

	// when
	err = client.DeleteDataTenant(subAccountID, environment)

	// then
	assert.NoError(t, err)

	response, err := testServer.Client().Get(fmt.Sprintf("%s/namespaces/%s/dataTenants/%s/%s", testServer.URL, testNamespace, subAccountID, environment))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestClient_CreateMetadataTenant(t *testing.T) {
	// given
	testServer := fixHTTPServer(t)
	defer testServer.Close()

	config := Config{
		AdminURL:  testServer.URL,
		Namespace: testNamespace,
	}
	client := NewClient(config, logger.NewLogDummy())
	client.setHttpClient(testServer.Client())

	// when
	err := client.CreateMetadataTenant(subAccountID, environment, MetadataTenantPayload{Key: "tK", Value: "tV"})
	assert.NoError(t, err)

	err = client.CreateMetadataTenant(subAccountID, environment, MetadataTenantPayload{Key: "tK2", Value: "tV2"})
	assert.NoError(t, err)

	// then
	assert.NoError(t, err)

	data, err := client.GetMetadataTenant(subAccountID, environment)
	assert.NoError(t, err)
	assert.Len(t, data, 2)
}

func TestClient_DeleteMetadataTenant(t *testing.T) {
	// given
	key := "tK"
	testServer := fixHTTPServer(t)
	defer testServer.Close()

	config := Config{
		AdminURL:  testServer.URL,
		Namespace: testNamespace,
	}
	client := NewClient(config, logger.NewLogDummy())
	client.setHttpClient(testServer.Client())

	err := client.CreateMetadataTenant(subAccountID, environment, MetadataTenantPayload{Key: key, Value: "tV"})
	assert.NoError(t, err)

	// when
	err = client.DeleteMetadataTenant(subAccountID, environment, key)

	// then
	assert.NoError(t, err)

	data, err := client.GetMetadataTenant(subAccountID, environment)
	assert.NoError(t, err)
	assert.Len(t, data, 0)
}

func fixHTTPServer(t *testing.T) *httptest.Server {
	r := mux.NewRouter()
	srv := newServer(t)

	r.HandleFunc("/namespaces/{namespace}/dataTenants", srv.createDataTenant).Methods(http.MethodPost)
	r.HandleFunc("/namespaces/{namespace}/dataTenants/{name}/{env}", srv.deleteDataTenant).Methods(http.MethodDelete)

	r.HandleFunc("/namespaces/{namespace}/dataTenants/{name}/{env}/metadata", srv.createMetadata).Methods(http.MethodPost)
	r.HandleFunc("/namespaces/{namespace}/dataTenants/{name}/{env}/metadata", srv.getMetadata).Methods(http.MethodGet)
	r.HandleFunc("/namespaces/{namespace}/dataTenants/{name}/{env}/metadata/{key}", srv.deleteMetadata).Methods(http.MethodDelete)

	// enpoints use only for test (exist in real EDP)
	r.HandleFunc("/namespaces/{namespace}/dataTenants/{name}/{env}", srv.getDataTenants).Methods(http.MethodGet)

	return httptest.NewServer(r)
}

type server struct {
	t          *testing.T
	metadata   []MetadataItem
	dataTenant map[string][]byte
}

func newServer(t *testing.T) *server {
	return &server{
		t:          t,
		metadata:   make([]MetadataItem, 0),
		dataTenant: make(map[string][]byte, 0),
	}
}

func (s *server) checkNamespace(w http.ResponseWriter, r *http.Request) (string, bool) {
	namespace, ok := mux.Vars(r)["namespace"]
	if !ok {
		s.t.Error("key namespace doesn't exist")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return "", false
	}
	if namespace != testNamespace {
		s.t.Errorf("key namespace is not equal to %s", testNamespace)
		w.WriteHeader(http.StatusNotFound)
		return namespace, false
	}

	return namespace, true
}

func (s *server) fetchNameAndEnv(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	vars := mux.Vars(r)
	name, okName := vars["name"]
	env, okEnv := vars["env"]

	if !okName || !okEnv {
		s.t.Error("one of the required key doesn't exist")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return "", "", false
	}

	return name, env, true
}

func (s *server) createDataTenant(w http.ResponseWriter, r *http.Request) {
	ns, ok := s.checkNamespace(w, r)
	if !ok {
		return
	}

	var dt DataTenantPayload
	err := json.NewDecoder(r.Body).Decode(&dt)
	if err != nil {
		s.t.Errorf("cannot read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dti := DataTenantItem{
		Namespace: NamespaceItem{
			Name: ns,
		},
		Name:        dt.Name,
		Environment: dt.Environment,
	}

	data, err := json.Marshal(dti)
	if err != nil {
		s.t.Errorf("wrong request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.dataTenant[dt.Name] = data
	w.WriteHeader(http.StatusCreated)
}

func (s *server) deleteDataTenant(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.checkNamespace(w, r); !ok {
		return
	}
	name, _, ok := s.fetchNameAndEnv(w, r)
	if !ok {
		return
	}

	for dtName, _ := range s.dataTenant {
		if dtName == name {
			delete(s.dataTenant, dtName)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// EDP server return 204 if dataTenant not exist already
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) createMetadata(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.checkNamespace(w, r); !ok {
		return
	}
	name, env, ok := s.fetchNameAndEnv(w, r)
	if !ok {
		return
	}

	var item MetadataItem
	err := json.NewDecoder(r.Body).Decode(&item)
	item.DataTenant = DataTenantItem{
		Namespace: NamespaceItem{
			Name: testNamespace,
		},
		Name:        name,
		Environment: env,
	}
	if err != nil {
		s.t.Errorf("cannot decode request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.metadata = append(s.metadata, item)
	w.WriteHeader(http.StatusCreated)
}

func (s *server) deleteMetadata(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.checkNamespace(w, r); !ok {
		return
	}

	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		s.t.Error("key doesn't exist")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	newData := make([]MetadataItem, 0)
	for _, item := range s.metadata {
		if item.Key == key {
			continue
		}
		newData = append(newData, item)
	}

	s.metadata = newData
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) getMetadata(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(s.metadata)
	if err != nil {
		s.t.Errorf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) getDataTenants(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.checkNamespace(w, r); !ok {
		return
	}
	name, _, ok := s.fetchNameAndEnv(w, r)
	if !ok {
		s.t.Error("cannot find name/env query parameters")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if _, ok := s.dataTenant[name]; !ok {
		s.t.Logf("dataTenant with name %s not exist", name)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := w.Write(s.dataTenant[name])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// setHttpClient auxiliary method of testing to get rid of oAuth client wrapper
func (c *Client) setHttpClient(httpClient *http.Client) {
	c.httpClient = httpClient
}
