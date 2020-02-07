package broker_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
)

const (
	operationID     = "23caac24-c317-47d0-bd2f-6b1bf4bdba99"
	runtimeID       = "b4491027-bdc1-4358-9098-a2f18c86e5c5"
	globalAccountID = "e8f7ec0a-0cd6-41f0-905d-5d1efa9fb6c4"
	instID          = "c39d9b98-5ed9-4a68-b786-f26ce93a734f"
	dashboardURL    = "https://console.e2e-provisioning.gophers.kyma.pro"
)

func TestLastOperation_LastOperation(t *testing.T) {
	// given
	// #setup memory storage
	memoryStorage := storage.NewMemoryStorage()
	err := memoryStorage.Instances().Insert(fixInstance())
	assert.NoError(t, err)

	// #setup provisioner client
	provisionClient := provisioner.NewFakeClient()
	operationMessage := "success"
	provisionClient.SetOperation(operationID, schema.OperationStatus{
		ID:        ptr.String(operationID),
		Operation: schema.OperationTypeProvision,
		State:     schema.OperationStateSucceeded,
		Message:   &operationMessage,
		RuntimeID: ptr.String(runtimeID),
	})

	// #setup director client
	directorClient := director.NewFakeDirectorClient()
	directorClient.SetConsoleURL(runtimeID, dashboardURL)

	// #create LastOperation endpoint
	lastOperationEndpoint := broker.NewLastOperation(
		memoryStorage.Instances(),
		provisionClient,
		directorClient,
		&broker.DumyDumper{},
	)

	// when
	response, err := lastOperationEndpoint.LastOperation(context.TODO(), instID, domain.PollDetails{OperationData: operationID})
	assert.NoError(t, err)

	// then
	assert.Equal(t, domain.LastOperation{
		State:       domain.Succeeded,
		Description: operationMessage,
	}, response)

	instance, err := memoryStorage.Instances().GetByID(instID)
	assert.NoError(t, err)
	assert.Equal(t, dashboardURL, instance.DashboardURL)
}

func fixInstance() internal.Instance {
	return internal.Instance{
		InstanceID:             instID,
		RuntimeID:              runtimeID,
		GlobalAccountID:        globalAccountID,
		ServiceID:              "2222",
		ServicePlanID:          "3333",
		DashboardURL:           "",
		ProvisioningParameters: "",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		DelatedAt:              time.Time{},
	}
}
