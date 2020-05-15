package internal

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type ProvisionInputCreator interface {
	SetProvisioningParameters(params ProvisioningParametersDTO) ProvisionInputCreator
	SetLabel(key, value string) ProvisionInputCreator
	// Deprecated, use: AppendOverrides
	SetOverrides(component string, overrides []*gqlschema.ConfigEntryInput) ProvisionInputCreator
	AppendOverrides(component string, overrides []*gqlschema.ConfigEntryInput) ProvisionInputCreator
	AppendGlobalOverrides(overrides []*gqlschema.ConfigEntryInput) ProvisionInputCreator
	Create() (gqlschema.ProvisionRuntimeInput, error)
}

type LMSTenant struct {
	ID        string
	Name      string
	Region    string
	CreatedAt time.Time
}

type LMS struct {
	TenantID    string    `json:"tenant_id"`
	Failed      bool      `json:"failed"`
	RequestedAt time.Time `json:"requested_at"`
}

type AvsLifecycleData struct {
	AvsEvaluationInternalId int64 `json:"avs_evaluation_internal_id"`
	AVSEvaluationExternalId int64 `json:"avs_evaluation_external_id"`

	AVSInternalEvaluationDeleted bool `json:"avs_internal_evaluation_deleted"`
	AVSExternalEvaluationDeleted bool `json:"avs_external_evaluation_deleted"`
}

type EventHub struct {
	Deleted bool `json:"event_hub_deleted"`
}

type Instance struct {
	InstanceID      string
	RuntimeID       string
	GlobalAccountID string
	SubAccountID    string
	ServiceID       string
	ServiceName     string
	ServicePlanID   string
	ServicePlanName string

	DashboardURL           string
	ProvisioningParameters string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (instance Instance) GetProvisioningParameters() (ProvisioningParameters, error) {
	var pp ProvisioningParameters

	err := json.Unmarshal([]byte(instance.ProvisioningParameters), &pp)
	if err != nil {
		return pp, errors.Wrap(err, "while unmarshalling provisioning parameters")
	}

	return pp, nil
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

type InstanceWithOperation struct {
	Instance

	Type        sql.NullString
	State       sql.NullString
	Description sql.NullString
}

// ProvisioningOperation holds all information about provisioning operation
type ProvisioningOperation struct {
	Operation `json:"-"`

	// following fields are serialized to JSON and stored in the storage
	Lms                    LMS    `json:"lms"`
	ProvisioningParameters string `json:"provisioning_parameters"`

	// following fields are not stored in the storage
	InputCreator ProvisionInputCreator `json:"-"`

	Avs AvsLifecycleData `json:"avs"`

	RuntimeID string `json:"runtime_id"`
}

// DeprovisioningOperation holds all information about de-provisioning operation
type DeprovisioningOperation struct {
	Operation `json:"-"`

	ProvisioningParameters string           `json:"provisioning_parameters"`
	Avs                    AvsLifecycleData `json:"avs"`
	EventHub               EventHub         `json:"eh"`
	SubAccountID           string           `json:"-"`
	RuntimeID              string           `json:"runtime_id"`
}

// OperationStats provide number of operations per type and state
type OperationStats struct {
	Provisioning   map[domain.LastOperationState]int
	Deprovisioning map[domain.LastOperationState]int
}

// NewProvisioningOperation creates a fresh (just starting) instance of the ProvisioningOperation
func NewProvisioningOperation(instanceID string, parameters ProvisioningParameters) (ProvisioningOperation, error) {
	return NewProvisioningOperationWithID(uuid.New().String(), instanceID, parameters)
}

// NewProvisioningOperationWithID creates a fresh (just starting) instance of the ProvisioningOperation with provided ID
func NewProvisioningOperationWithID(operationID, instanceID string, parameters ProvisioningParameters) (ProvisioningOperation, error) {
	params, err := json.Marshal(parameters)
	if err != nil {
		return ProvisioningOperation{}, errors.Wrap(err, "while marshaling provisioning parameters")
	}

	return ProvisioningOperation{
		Operation: Operation{
			ID:          operationID,
			Version:     0,
			Description: "Operation created",
			InstanceID:  instanceID,
			State:       domain.InProgress,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		ProvisioningParameters: string(params),
	}, nil
}

// NewProvisioningOperationWithID creates a fresh (just starting) instance of the ProvisioningOperation with provided ID
func NewDeprovisioningOperationWithID(operationID, instanceID string) (DeprovisioningOperation, error) {
	return DeprovisioningOperation{
		Operation: Operation{
			ID:          operationID,
			Version:     0,
			Description: "Operation created",
			InstanceID:  instanceID,
			State:       domain.InProgress,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}, nil
}

func (po *ProvisioningOperation) GetProvisioningParameters() (ProvisioningParameters, error) {
	var pp ProvisioningParameters

	err := json.Unmarshal([]byte(po.ProvisioningParameters), &pp)
	if err != nil {
		return pp, errors.Wrap(err, "while unmarshaling provisioning parameters")
	}

	return pp, nil
}

func (po *ProvisioningOperation) SetProvisioningParameters(parameters ProvisioningParameters) error {
	params, err := json.Marshal(parameters)
	if err != nil {
		return errors.Wrap(err, "while marshaling provisioning parameters")
	}

	po.ProvisioningParameters = string(params)
	return nil
}

func (do *DeprovisioningOperation) GetProvisioningParameters() (ProvisioningParameters, error) {
	var pp ProvisioningParameters

	err := json.Unmarshal([]byte(do.ProvisioningParameters), &pp)
	if err != nil {
		return pp, errors.Wrap(err, "while unmarshaling provisioning parameters")
	}

	return pp, nil
}

func (do *DeprovisioningOperation) SetProvisioningParameters(parameters ProvisioningParameters) error {
	params, err := json.Marshal(parameters)
	if err != nil {
		return errors.Wrap(err, "while marshaling provisioning parameters")
	}

	do.ProvisioningParameters = string(params)
	return nil
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
			SourceURL:     component.SourceURL,
			Configuration: cpyCfg,
		})
	}
	return copiedList
}
