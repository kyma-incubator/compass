package provisioning

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	inputAutomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input/automock"
	provisionerAutomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	kymaVersion            = "1.10"
	instanceID             = "58f8c703-1756-48ab-9299-a847974d1fee"
	operationID            = "fd5cee4d-0eeb-40d0-a7a7-0708e5eba470"
	globalAccountID        = "80ac17bd-33e8-4ffa-8d56-1d5367755723"
	subAccountID           = "12df5747-3efb-4df6-ad6f-4414bb661ce3"
	provisionerOperationID = "1a0ed09b-9bb9-4e6f-a88c-01955c5f1129"
	runtimeID              = "2498c8ee-803a-43c2-8194-6d6dd0354c30"

	serviceManagerURL      = "http://sm.com"
	serviceManagerUser     = "admin"
	serviceManagerPassword = "admin123"
)

func TestCreateRuntimeStep_Run(t *testing.T) {
	// given
	log := logrus.New()
	memoryStorage := storage.NewMemoryStorage()

	operation := fixOperationCreateRuntime(t)
	err := memoryStorage.Operations().InsertProvisioningOperation(operation)
	assert.NoError(t, err)

	provisionerClient := &provisionerAutomock.Client{}
	provisionerClient.On("ProvisionRuntime", globalAccountID, subAccountID, gqlschema.ProvisionRuntimeInput{
		RuntimeInput: &gqlschema.RuntimeInput{
			Name:        "",
			Description: nil,
			Labels: &gqlschema.Labels{
				"broker_instance_id":   []string{instanceID},
				"global_subaccount_id": []string{subAccountID},
			},
		},
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				KubernetesVersion: "1.15.5",
				DiskType:          "pd-standard",
				VolumeSizeGb:      30,
				MachineType:       "n1-standard-4",
				Region:            "europe-west4-a",
				Provider:          "gcp",
				WorkerCidr:        "10.250.0.0/19",
				AutoScalerMin:     2,
				AutoScalerMax:     4,
				MaxSurge:          4,
				MaxUnavailable:    1,
				TargetSecret:      "",
				ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
					GcpConfig: &gqlschema.GCPProviderConfigInput{
						Zone: "europe-west4-b",
					},
				},
				Seed: nil,
			},
			GcpConfig: nil,
		},
		KymaConfig: &gqlschema.KymaConfigInput{
			Version: kymaVersion,
			Components: internal.ComponentConfigurationInputList{
				{
					Component:     "keb",
					Namespace:     "kyma-system",
					Configuration: nil,
				},
			},
			Configuration: nil,
		},
		Credentials: nil,
	}).Return(gqlschema.OperationStatus{
		ID:        ptr.String(provisionerOperationID),
		Operation: "",
		State:     "",
		Message:   nil,
		RuntimeID: nil,
	}, nil)

	provisionerClient.On("RuntimeOperationStatus", globalAccountID, provisionerOperationID).Return(gqlschema.OperationStatus{
		ID:        ptr.String(provisionerOperationID),
		Operation: "",
		State:     "",
		Message:   nil,
		RuntimeID: ptr.String(runtimeID),
	}, nil)

	step := NewCreateRuntimeStep(memoryStorage.Operations(), memoryStorage.Instances(), provisionerClient)

	// when
	entry := log.WithFields(logrus.Fields{"step": "TEST"})
	operation, repeat, err := step.Run(operation, entry)

	// then
	assert.NoError(t, err)
	assert.Equal(t, 1*time.Second, repeat)
	assert.Equal(t, provisionerOperationID, operation.ProvisionerOperationID)

	instance, err := memoryStorage.Instances().GetByID(operation.InstanceID)
	assert.NoError(t, err)
	assert.Equal(t, instance.RuntimeID, runtimeID)
}

func fixOperationCreateRuntime(t *testing.T) internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:          operationID,
			InstanceID:  instanceID,
			Description: "",
			UpdatedAt:   time.Now(),
		},
		ProvisioningParameters: fixProvisioningParameters(t),
		InputCreator:           fixInputCreator(t),
	}
}

func fixProvisioningParameters(t *testing.T) string {
	parameters := internal.ProvisioningParameters{
		PlanID:    broker.GcpPlanID,
		ServiceID: "",
		ErsContext: internal.ERSContext{
			GlobalAccountID: globalAccountID,
			SubAccountID:    subAccountID,
			ServiceManager: &internal.ServiceManagerEntryDTO{
				Credentials: internal.ServiceManagerCredentials{
					BasicAuth: internal.ServiceManagerBasicAuth{
						Username: serviceManagerUser,
						Password: serviceManagerPassword,
					},
				},
				URL: serviceManagerURL,
			},
		},
		Parameters: internal.ProvisioningParametersDTO{
			Region: ptr.String("europe-west4-a"),
		},
	}

	rawParameters, err := json.Marshal(parameters)
	if err != nil {
		t.Errorf("cannot marshal provisioning parameters: %s", err)
	}

	return string(rawParameters)
}

func fixInputCreator(t *testing.T) internal.ProvisionInputCreator {
	optComponentsSvc := &inputAutomock.OptionalComponentService{}

	optComponentsSvc.On("ComputeComponentsToDisable", []string(nil)).Return([]string{})
	optComponentsSvc.On("ExecuteDisablers", internal.ComponentConfigurationInputList{
		{
			Component:     "to-remove-component",
			Namespace:     "kyma-system",
			Configuration: nil,
		},
		{
			Component:     "keb",
			Namespace:     "kyma-system",
			Configuration: nil,
		},
	}).Return(internal.ComponentConfigurationInputList{
		{
			Component:     "keb",
			Namespace:     "kyma-system",
			Configuration: nil,
		},
	}, nil)

	kymaComponentList := []v1alpha1.KymaComponent{
		{
			Name:      "to-remove-component",
			Namespace: "kyma-system",
		},
		{
			Name:      "keb",
			Namespace: "kyma-system",
		},
	}
	ibf := input.NewInputBuilderFactory(optComponentsSvc, kymaComponentList, input.Config{}, kymaVersion)

	creator, found := ibf.ForPlan(broker.GcpPlanID)
	if !found {
		t.Errorf("input creator for %q plan does not exist", broker.GcpPlanID)
	}

	return creator
}
