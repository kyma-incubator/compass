package types

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

type TenantMapping struct {
	FormationID     string           `json:"formationId"`
	ReceiverTenant  ReceiverTenant   `json:"receiverTenant"`
	AssignedTenants []AssignedTenant `json:"assignedTenants"`
}

type ReceiverTenant struct {
	ApplicationURL string `json:"applicationUrl"`
}

type Operation string

const (
	OperationAssign   Operation = "assign"
	OperationUnassign Operation = "unassign"
)

type AssignedTenant struct {
	UCLApplicationID string                      `json:"uclApplicationId"`
	LocalTenantID    string                      `json:"localTenantId"`
	Operation        Operation                   `json:"operation"`
	Parameters       AssignedTenantParameters    `json:"parameters"`
	Config           any                         `json:"configuration"`
	Configuration    AssignedTenantConfiguration `json:"-"`
}

func (at *AssignedTenant) SetConfiguration(ctx context.Context) error {
	log := logger.FromContext(ctx)

	if at.Config == nil {
		log.Info().Msg("$.assignedTenants[0].configuration is empty")
		return nil
	}
	b, err := json.Marshal(at.Config)
	if err != nil {
		return errors.Newf("failed to marshal $.assignedTenants[0].configuration: %w", err)
	}
	if err := json.Unmarshal(b, &at.Configuration); err != nil {
		log.Info().Msgf("$.assignedTenants[0].configuration doesn't contain apis: %s", err)
		return nil
	}

	return nil
}

type AssignedTenantParameters struct {
	ClientID string `json:"technicalIntegrationId"`
}

type AssignedTenantConfiguration struct {
	ConsumedAPIs []string `json:"apis"`
}

func (tm TenantMapping) Validate() error {
	if _, err := uuid.Parse(tm.FormationID); err != nil {
		return errors.New("$.formationId is not a valid uuid")
	}
	if tm.ReceiverTenant.ApplicationURL == "" {
		return errors.New("$.receiverTenant.applicationUrl is required")
	}
	if _, err := uuid.Parse(tm.AssignedTenants[0].UCLApplicationID); err != nil {
		return errors.New("$.assignedTenants[0].uclApplicationId is not a valid uuid")
	}
	if tm.AssignedTenants[0].LocalTenantID == "" {
		return errors.New("$.assignedTenants[0].localTenantId is required")
	}
	if tm.AssignedTenants[0].Operation != OperationAssign && tm.AssignedTenants[0].Operation != OperationUnassign {
		return errors.New("$.assignedTenants[0].operation can only be assign or unassign")
	}
	if tm.AssignedTenants[0].Parameters.ClientID == "" {
		return errors.New("$.assignedTenants[0].parameters.technicalIntegrationId is required")
	}
	return nil
}
