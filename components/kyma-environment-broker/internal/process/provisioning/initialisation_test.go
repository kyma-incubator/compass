package provisioning

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	provisionerAutomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	statusOperationID            = "17f3ddba-1132-466d-a3c5-920f544d7ea6"
	statusInstanceID             = "9d75a545-2e1e-4786-abd8-a37b14e185b9"
	statusRuntimeID              = "ef4e3210-652c-453e-8015-bba1c1cd1e1c"
	statusGlobalAccountID        = "abf73c71-a653-4951-b9c2-a26d6c2cccbd"
	statusProvisionerOperationID = "e04de524-53b3-4890-b05a-296be393e4ba"

	dashboardURL = "http://runtime.com"
)

func TestInitialisationStep_Run(t *testing.T) {
	// given
	log := logrus.New()
	memoryStorage := storage.NewMemoryStorage()

	operation := fixOperationRuntimeStatus(t)
	err := memoryStorage.Operations().InsertProvisioningOperation(operation)
	assert.NoError(t, err)

	instance := fixInstanceRuntimeStatus()
	err = memoryStorage.Instances().Insert(instance)
	assert.NoError(t, err)

	provisionerClient := &provisionerAutomock.Client{}
	provisionerClient.On("RuntimeOperationStatus", statusGlobalAccountID, statusProvisionerOperationID).Return(gqlschema.OperationStatus{
		ID:        ptr.String(statusProvisionerOperationID),
		Operation: "",
		State:     gqlschema.OperationStateSucceeded,
		Message:   nil,
		RuntimeID: nil,
	}, nil)

	directorClient := &automock.DirectorClient{}
	directorClient.On("GetConsoleURL", statusGlobalAccountID, statusRuntimeID).Return(dashboardURL, nil)

	step := NewInitialisationStep(memoryStorage.Operations(), memoryStorage.Instances(), provisionerClient, directorClient, nil)

	// when
	operation, repeat, err := step.Run(operation, log)

	// then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
	assert.Equal(t, domain.Succeeded, operation.State)

	updatedInstance, err := memoryStorage.Instances().GetByID(statusInstanceID)
	assert.NoError(t, err)
	assert.Equal(t, dashboardURL, updatedInstance.DashboardURL)
}

func fixOperationRuntimeStatus(t *testing.T) internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:                     statusOperationID,
			InstanceID:             statusInstanceID,
			ProvisionerOperationID: statusProvisionerOperationID,
			Description:            "",
			UpdatedAt:              time.Now(),
		},
		ProvisioningParameters: fixProvisioningParametersRuntimeStatus(t),
	}
}

func fixProvisioningParametersRuntimeStatus(t *testing.T) string {
	parameters := internal.ProvisioningParameters{
		PlanID:    broker.GcpPlanID,
		ServiceID: "",
		ErsContext: internal.ERSContext{
			GlobalAccountID: statusGlobalAccountID,
		},
	}

	rawParameters, err := json.Marshal(parameters)
	if err != nil {
		t.Errorf("cannot marshal provisioning parameters: %s", err)
	}

	return string(rawParameters)
}

func fixInstanceRuntimeStatus() internal.Instance {
	return internal.Instance{
		InstanceID:      statusInstanceID,
		RuntimeID:       statusRuntimeID,
		DashboardURL:    "",
		GlobalAccountID: statusGlobalAccountID,
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
		DelatedAt:       time.Time{},
	}
}
