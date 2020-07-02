package provisioner

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/ptr"
	schema "github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	testAccountID    = "4346c639-32f8-4947-ae95-73bb8efad209"
	testSubAccountID = "42d45043-d0fb-4077-9de0-d7f781949bce"

	provisionRuntimeID            = "4e268c0f-d053-4ab7-b167-6dbc0a0e09a6"
	provisionRuntimeOperationID   = "c89f7862-0ef9-4d4e-bc82-afbc5ac98b8d"
	deprovisionRuntimeOperationID = "f9f7b734-7538-419c-8ac1-37060c60531a"
)

func TestClient_ProvisionRuntime(t *testing.T) {
	t.Run("should trigger provisioning", func(t *testing.T) {
		// Given
		tr := &testResolver{t: t, runtime: &testRuntime{}}
		testServer := fixHTTPServer(tr)
		defer testServer.Close()

		client := NewProvisionerClient(testServer.URL, false)

		// When
		status, err := client.ProvisionRuntime(testAccountID, testSubAccountID, fixProvisionRuntimeInput())

		// Then
		assert.NoError(t, err)
		assert.Equal(t, ptr.String(provisionRuntimeOperationID), status.ID)
		assert.Equal(t, schema.OperationStateInProgress, status.State)
		assert.Equal(t, ptr.String(provisionRuntimeID), status.RuntimeID)

		assert.Equal(t, "test", tr.getRuntime().name)
	})

	t.Run("provisioner should return error", func(t *testing.T) {
		// Given
		tr := &testResolver{t: t, runtime: &testRuntime{}, failed: true}
		testServer := fixHTTPServer(tr)
		defer testServer.Close()

		client := NewProvisionerClient(testServer.URL, false)

		// When
		status, err := client.ProvisionRuntime(testAccountID, testSubAccountID, fixProvisionRuntimeInput())

		// Then
		assert.Error(t, err)
		assert.Empty(t, status)

		assert.Equal(t, "", tr.getRuntime().name)
	})
}

func TestClient_DeprovisionRuntime(t *testing.T) {
	t.Run("should trigger deprovisioning", func(t *testing.T) {
		// Given
		tr := &testResolver{t: t, runtime: &testRuntime{}}
		testServer := fixHTTPServer(tr)
		defer testServer.Close()

		client := NewProvisionerClient(testServer.URL, false)
		operation, err := client.ProvisionRuntime(testAccountID, testSubAccountID, fixProvisionRuntimeInput())
		assert.NoError(t, err)

		// When
		operationId, err := client.DeprovisionRuntime(testAccountID, *operation.RuntimeID)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, deprovisionRuntimeOperationID, operationId)

		assert.Empty(t, tr.getRuntime().runtimeID)
	})

	t.Run("provisioner should return error", func(t *testing.T) {
		// Given
		tr := &testResolver{t: t, runtime: &testRuntime{}}
		testServer := fixHTTPServer(tr)
		defer testServer.Close()

		client := NewProvisionerClient(testServer.URL, false)
		operation, err := client.ProvisionRuntime(testAccountID, testSubAccountID, fixProvisionRuntimeInput())
		assert.NoError(t, err)

		tr.failed = true

		// When
		operationId, err := client.DeprovisionRuntime(testAccountID, *operation.RuntimeID)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "", operationId)

		assert.Equal(t, provisionRuntimeID, tr.getRuntime().runtimeID)
	})
}

func TestClient_ReconnectRuntimeAgent(t *testing.T) {
	t.Run("should reconnect runtime agent", func(t *testing.T) {
		// Given
		tr := &testResolver{t: t, runtime: &testRuntime{}}
		testServer := fixHTTPServer(tr)
		defer testServer.Close()

		client := NewProvisionerClient(testServer.URL, false)
		operation, err := client.ProvisionRuntime(testAccountID, testSubAccountID, fixProvisionRuntimeInput())
		assert.NoError(t, err)

		// When
		operationId, err := client.ReconnectRuntimeAgent(testAccountID, *operation.RuntimeID)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, provisionRuntimeOperationID, operationId)
	})

	t.Run("provisioner should return error", func(t *testing.T) {
		// Given
		tr := &testResolver{t: t, runtime: &testRuntime{}}
		testServer := fixHTTPServer(tr)
		defer testServer.Close()

		client := NewProvisionerClient(testServer.URL, false)
		operation, err := client.ProvisionRuntime(testAccountID, testSubAccountID, fixProvisionRuntimeInput())
		assert.NoError(t, err)

		tr.failed = true

		// When
		operationId, err := client.ReconnectRuntimeAgent(testAccountID, *operation.RuntimeID)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "", operationId)
	})
}

func TestClient_RuntimeOperationStatus(t *testing.T) {
	t.Run("should return runtime operation status", func(t *testing.T) {
		// Given
		tr := &testResolver{t: t, runtime: &testRuntime{}}
		testServer := fixHTTPServer(tr)
		defer testServer.Close()

		client := NewProvisionerClient(testServer.URL, false)
		_, err := client.ProvisionRuntime(testAccountID, testSubAccountID, fixProvisionRuntimeInput())
		assert.NoError(t, err)

		// When
		status, err := client.RuntimeOperationStatus(testAccountID, provisionRuntimeID)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, ptr.String(provisionRuntimeID), status.RuntimeID)
		assert.Equal(t, ptr.String(provisionRuntimeOperationID), status.ID)
		assert.Equal(t, schema.OperationStateInProgress, status.State)
		assert.Equal(t, schema.OperationTypeProvision, status.Operation)
	})

	t.Run("provisioner should return error", func(t *testing.T) {
		// Given
		tr := &testResolver{t: t, runtime: &testRuntime{}}
		testServer := fixHTTPServer(tr)
		defer testServer.Close()

		client := NewProvisionerClient(testServer.URL, false)
		_, err := client.ProvisionRuntime(testAccountID, testSubAccountID, fixProvisionRuntimeInput())
		assert.NoError(t, err)

		tr.failed = true

		// When
		status, err := client.RuntimeOperationStatus(testAccountID, provisionRuntimeID)

		// Then
		assert.Error(t, err)
		assert.Empty(t, status)
	})
}

type testRuntime struct {
	tenant                 string
	clientID               string
	name                   string
	runtimeID              string
	provisionOperationID   string
	deprovisionOperationID string
}

type testResolver struct {
	t       *testing.T
	runtime *testRuntime
	failed  bool
}

func fixHTTPServer(tr *testResolver) *httptest.Server {
	r := mux.NewRouter()

	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountID := r.Header.Get(accountIDKey)
			subAccountID := r.Header.Get(subAccountIDKey)
			if accountID != testAccountID {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			tr.runtime.tenant = accountID
			tr.runtime.clientID = subAccountID

			h.ServeHTTP(w, r)
		})
	})
	r.HandleFunc("/", handler.GraphQL(schema.NewExecutableSchema(schema.Config{Resolvers: tr})))

	return httptest.NewServer(r)
}

func (tr testResolver) Mutation() schema.MutationResolver {
	tr.t.Log("Mutation TestResolver")
	return &testMutationResolver{t: tr.t, runtime: tr.runtime, failed: tr.failed}
}

func (tr testResolver) Query() schema.QueryResolver {
	tr.t.Log("Query TestResolver")
	return &testQueryResolver{t: tr.t, runtime: tr.runtime, failed: tr.failed}
}

func (tr testResolver) getRuntime() *testRuntime {
	return tr.runtime
}

type testMutationResolver struct {
	t       *testing.T
	runtime *testRuntime
	failed  bool
}

func (tmr *testMutationResolver) ProvisionRuntime(_ context.Context, config schema.ProvisionRuntimeInput) (*schema.OperationStatus, error) {
	tmr.t.Log("ProvisionRuntime testMutationResolver")

	if tmr.failed {
		return nil, fmt.Errorf("provision runtime failed for %s", config.RuntimeInput.Name)
	}

	tmr.runtime.name = config.RuntimeInput.Name
	tmr.runtime.runtimeID = provisionRuntimeID
	tmr.runtime.provisionOperationID = provisionRuntimeOperationID

	return &schema.OperationStatus{
		ID:        ptr.String(tmr.runtime.provisionOperationID),
		State:     schema.OperationStateInProgress,
		RuntimeID: ptr.String(tmr.runtime.runtimeID),
	}, nil
}

func (tmr testMutationResolver) UpgradeRuntime(_ context.Context, id string, config schema.UpgradeRuntimeInput) (*schema.OperationStatus, error) {
	return nil, nil
}

func (tmr testMutationResolver) DeprovisionRuntime(_ context.Context, id string) (string, error) {
	tmr.t.Log("DeprovisionRuntime testMutationResolver")

	if tmr.failed {
		return "", fmt.Errorf("deprovision failed for %s", id)
	}

	if tmr.runtime.runtimeID == id {
		tmr.runtime.runtimeID = ""
		tmr.runtime.name = ""
		tmr.runtime.deprovisionOperationID = deprovisionRuntimeOperationID
	}

	return tmr.runtime.deprovisionOperationID, nil
}

func (tmr testMutationResolver) RollBackUpgradeOperation(_ context.Context, id string) (*schema.RuntimeStatus, error) {
	return nil, nil
}

func (tmr testMutationResolver) ReconnectRuntimeAgent(_ context.Context, id string) (string, error) {
	tmr.t.Log("ReconnectRuntimeAgent testMutationResolver")

	if tmr.failed {
		return "", fmt.Errorf("reconnect runtime agent failed for %s", id)
	}

	if tmr.runtime.runtimeID == id {
		return tmr.runtime.provisionOperationID, nil
	}

	return "", nil
}

type testQueryResolver struct {
	t       *testing.T
	runtime *testRuntime
	failed  bool
}

func (tqr testQueryResolver) RuntimeStatus(_ context.Context, id string) (*schema.RuntimeStatus, error) {
	return nil, nil
}

func (tqr testQueryResolver) RuntimeOperationStatus(_ context.Context, id string) (*schema.OperationStatus, error) {
	tqr.t.Log("RuntimeOperationStatus - testQueryResolver")

	if tqr.failed {
		return nil, fmt.Errorf("query about runtime operation status failed for %s", id)
	}

	if tqr.runtime.runtimeID == id {
		return &schema.OperationStatus{
			ID:        ptr.String(tqr.runtime.provisionOperationID),
			Operation: schema.OperationTypeProvision,
			State:     schema.OperationStateInProgress,
			Message:   ptr.String("test message"),
			RuntimeID: ptr.String(tqr.runtime.runtimeID),
		}, nil
	}

	return nil, nil
}

func fixProvisionRuntimeInput() schema.ProvisionRuntimeInput {
	return schema.ProvisionRuntimeInput{
		RuntimeInput: &schema.RuntimeInput{
			Name:        "test",
			Description: nil,
			Labels:      nil,
		},
		ClusterConfig: &schema.ClusterConfigInput{
			GardenerConfig: &schema.GardenerConfigInput{
				ProviderSpecificConfig: &schema.ProviderSpecificInput{},
			},
		},
		KymaConfig: &schema.KymaConfigInput{
			Components: []*schema.ComponentConfigurationInput{
				{
					Component: "test",
				},
			},
		},
	}
}
