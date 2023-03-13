package types

import "errors"

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
	Operation        Operation                   `json:"operation"`
	AssignmentID     string                      `json:"assignmentId"`
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
	if tm.FormationID == "" {
		return errors.New("$.formationId is required")
	}
	if tm.ReceiverTenant.ApplicationURL == "" {
		return errors.New("$.receiverTenant.applicationUrl is required")
	}
	if tm.AssignedTenants[0].Operation == "" {
		return errors.New("$.assignedTenants[0].operation is required")
	}
	if tm.AssignedTenants[0].AssignmentID == "" {
		return errors.New("$.assignedTenants[0].assignmentId is required")
	}
	if tm.AssignedTenants[0].Parameters.ClientID == "" {
		return errors.New("$.assignedTenants[0].parameters.clientId is required")
	}
	return nil
}
