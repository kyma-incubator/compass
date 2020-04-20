package edp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
	client := NewClient(config, testServer.Client(), logrus.New())

	// when
	err := client.CreateDataTenant(DataTenantPayload{})

	// then
	assert.NoError(t, err)
}

func TestClient_DeleteDataTenant(t *testing.T) {
	// given
	testServer := fixHTTPServer(t)
	defer testServer.Close()

	config := Config{
		AdminURL:  testServer.URL,
		Namespace: testNamespace,
	}
	client := NewClient(config, testServer.Client(), logrus.New())

	err := client.CreateDataTenant(DataTenantPayload{
		Name:        subAccountID,
		Environment: environment,
	})
	assert.NoError(t, err)

	// when
	err = client.DeleteDataTenant(subAccountID, environment)

	// then
	assert.NoError(t, err)
}

func TestClient_CreateMetadataTenant(t *testing.T) {
	// given
	testServer := fixHTTPServer(t)
	defer testServer.Close()

	config := Config{
		AdminURL:  testServer.URL,
		Namespace: testNamespace,
	}
	client := NewClient(config, testServer.Client(), logrus.New())

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
	client := NewClient(config, testServer.Client(), logrus.New())

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

	return httptest.NewServer(r)
}

type server struct {
	t          *testing.T
	metadata   []MetadataItem
	dataTenant map[string]string
}

func newServer(t *testing.T) *server {
	return &server{
		t:          t,
		metadata:   make([]MetadataItem, 0),
		dataTenant: make(map[string]string, 0),
	}
}

func (s *server) checkNamespace(w http.ResponseWriter, r *http.Request) bool {
	namespace, ok := mux.Vars(r)["namespace"]
	if !ok {
		s.t.Error("key namespace doesn't exist")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
	if namespace != testNamespace {
		s.t.Errorf("key namespace is not equal to %s", testNamespace)
		w.WriteHeader(http.StatusNotFound)
		return false
	}

	return true
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
	if !s.checkNamespace(w, r) {
		return
	}

	var dt DataTenantPayload
	err := json.NewDecoder(r.Body).Decode(&dt)
	if err != nil {
		s.t.Errorf("cannot decode request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.dataTenant[dt.Name] = dt.Environment
	w.WriteHeader(http.StatusCreated)
}

func (s *server) deleteDataTenant(w http.ResponseWriter, r *http.Request) {
	if !s.checkNamespace(w, r) {
		return
	}
	name, env, ok := s.fetchNameAndEnv(w, r)
	if !ok {
		return
	}

	for dtName, dtEnv := range s.dataTenant {
		if dtName == name && dtEnv == env {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s *server) createMetadata(w http.ResponseWriter, r *http.Request) {
	if !s.checkNamespace(w, r) {
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
	if !s.checkNamespace(w, r) {
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
