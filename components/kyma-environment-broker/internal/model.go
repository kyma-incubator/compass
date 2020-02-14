package internal

import (
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
)

type Instance struct {
	InstanceID      string
	RuntimeID       string
	GlobalAccountID string
	ServiceID       string
	ServicePlanID   string

	DashboardURL           string
	ProvisioningParameters string

	CreatedAt time.Time
	UpdatedAt time.Time
	DelatedAt time.Time
}

type Operation struct {
	ID        string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time

	InstanceID             string
	ProvisionerOperationID string
	State                  domain.LastOperationState
	Description            string
}

// ProvisioningOperation holds all information about provisioning operation
type ProvisioningOperation struct {
	Operation `json:"-"`

	// following fields are serialized to JSON and stored in the storage
	LmsTenantID            string `json:"lms_tenant_id"`
	ProvisioningParameters string `json:"provisioning_parameters"`
}

// NewProvisioningOperation creates a fresh (just starting) instance of the ProvisioningOperation
func NewProvisioningOperation(instanceID, params string) ProvisioningOperation {
	return ProvisioningOperation{
		Operation: Operation{
			ID:          uuid.New().String(),
			Version:     0,
			Description: "Operation created",
			InstanceID:  instanceID,
			State:       domain.InProgress,
		},
		ProvisioningParameters: params,
	}
}

type ComponentConfigurationInputList []*gqlschema.ComponentConfigurationInput

func (l ComponentConfigurationInputList) DeepCopy() []*gqlschema.ComponentConfigurationInput {
	var copiedList []*gqlschema.ComponentConfigurationInput
	for _, component := range l {
		var cpyCfg []*gqlschema.ConfigEntryInput
		for _, cfg := range component.Configuration {
			mapped := &gqlschema.ConfigEntryInput{
				Key:   cfg.Key,
				Value: cfg.Value,
			}
			if cfg.Secret != nil {
				mapped.Secret = ptr.Bool(*cfg.Secret)
			}
			cpyCfg = append(cpyCfg, mapped)
		}

		copiedList = append(copiedList, &gqlschema.ComponentConfigurationInput{
			Component:     component.Component,
			Namespace:     component.Namespace,
			Configuration: cpyCfg,
		})
	}
	return copiedList
}

type ServiceManagerOverride struct {
	CredentialsOverride bool `envconfig:"default=false"`
	URL                 string
	Password            string
	Username            string
}
