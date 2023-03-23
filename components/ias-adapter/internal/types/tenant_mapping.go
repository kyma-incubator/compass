package types

import (
	"errors"

	"github.com/google/uuid"
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
	Configuration    AssignedTenantConfiguration `json:"configuration"`
}

type AssignedTenantParameters struct {
	ClientID string `json:"clientId"`
}

type AssignedTenantConfiguration struct {
	ConsumedAPIs []string `json:"consumedApis"`
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
	if tm.AssignedTenants[0].Operation == "" {
		return errors.New("$.assignedTenants[0].operation is required")
	}
	if tm.AssignedTenants[0].Parameters.ClientID == "" {
		return errors.New("$.assignedTenants[0].parameters.clientId is required")
	}
	return nil
}
