package broker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/stretchr/testify/assert"
)

const (
	fixInstanceID = "72b83910-ac12-4dcb-b91d-960cca2b36abx"
	fixRuntimeID  = "24da44ea-0295-4b1c-b5c1-6fd26efa4f24"
	fixOpID       = "04f91bff-9e17-45cb-a246-84d511274ef1"

	gcpPlanID   = "ca6e5357-707f-4565-bbbd-b3ab732597c6"
	azurePlanID = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
)

func TestClient_Deprovision(t *testing.T) {
	t.Run("should return deprovisioning operation ID on success", func(t *testing.T) {
		// given
		testServer := fixHTTPServer(false)
		defer testServer.Close()

		config := Config{
			URL: testServer.URL,
		}
		client := NewClient(context.Background(), config)
		client.setHttpClient(testServer.Client())

		instance := internal.Instance{
			InstanceID:    fixInstanceID,
			RuntimeID:     fixRuntimeID,
			ServicePlanID: azurePlanID,
		}

		// when
		opID, err := client.Deprovision(instance)

		// then
		assert.NoError(t, err)
		assert.Equal(t, fixOpID, opID)
	})

	t.Run("should return error on failed request execution", func(t *testing.T) {
		// given
		testServer := fixHTTPServer(true)
		defer testServer.Close()

		config := Config{
			URL: testServer.URL,
		}
		client := NewClient(context.Background(), config)
		client.setHttpClient(testServer.Client())

		instance := internal.Instance{
			InstanceID:    fixInstanceID,
			RuntimeID:     fixRuntimeID,
			ServicePlanID: gcpPlanID,
		}

		// when
		opID, err := client.Deprovision(instance)

		// then
		assert.Error(t, err)
		assert.Len(t, opID, 0)
	})
}

func fixHTTPServer(withFailure bool) *httptest.Server {
	if withFailure {
		r := mux.NewRouter()
		r.HandleFunc("/oauth/v2/service_instances/{instance_id}", deprovisionFailure).Methods(http.MethodDelete)
		return httptest.NewServer(r)
	}

	r := mux.NewRouter()
	r.HandleFunc("/oauth/v2/service_instances/{instance_id}", deprovision).Methods(http.MethodDelete)
	return httptest.NewServer(r)
}

func deprovision(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	_, okServiceID := params["service_id"]
	if !okServiceID {
		w.WriteHeader(http.StatusBadRequest)
	}
	_, okPlanID := params["plan_id"]
	if !okPlanID {
		w.WriteHeader(http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf(`{"operation": "%s"}`, fixOpID)))
}

func deprovisionFailure(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}
