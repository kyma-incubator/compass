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

func TestClient_CreateDataStream(t *testing.T) {
	// given
	testServer := fixHTTPServer(t)
	defer testServer.Close()

	config := ClientConfig{
		AdminURL:  testServer.URL,
		Namespace: testNamespace,
	}
	client := NewClient(config, testServer.Client(), logrus.New())

	// when
	err := client.CreateDataTenant(DataTenantPayload{})

	// then
	assert.NoError(t, err)
}

func TestClient_CreateMetadataTenant(t *testing.T) {
	// given
	testServer := fixHTTPServer(t)
	defer testServer.Close()

	config := ClientConfig{
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

	config := ClientConfig{
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
	data := make([]MetadataItem, 0)

	r.HandleFunc("/namespaces/{namespace}/dataTenants", func(w http.ResponseWriter, r *http.Request) {
		_, ok := mux.Vars(r)["namespace"]
		if !ok {
			t.Error("key namespace doesn't exist")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodPost)
	r.HandleFunc("/namespaces/{namespace}/dataTenants/{name}/{env}/metadata", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace, okNs := vars["namespace"]
		name, okN := vars["name"]
		env, okEnv := vars["env"]
		if !okNs || !okN || !okEnv {
			t.Error("one of the required key doesn't exist")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var item MetadataItem
		err := json.NewDecoder(r.Body).Decode(&item)
		item.DataTenant = DataTenantItem{
			Namespace: NamespaceItem{
				Name: namespace,
			},
			Name:        name,
			Environment: env,
		}
		if err != nil {
			t.Errorf("cannot decode request body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		data = append(data, item)
		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodPost)
	r.HandleFunc("/namespaces/{namespace}/dataTenants/{name}/{env}/metadata", func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			t.Errorf("%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)
	r.HandleFunc("/namespaces/{namespace}/dataTenants/{name}/{env}/metadata/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key, ok := vars["key"]
		if !ok {
			t.Error("key doesn't exist")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		newData := make([]MetadataItem, 0)
		for _, item := range data {
			if item.Key == key {
				continue
			}
			newData = append(newData, item)
		}
		data = newData
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(r)
}
