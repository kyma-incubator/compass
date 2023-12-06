package types

import "encoding/json"

type TenantMapping struct {
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
	AssignedTenant AssignedTenant `json:"assignedTenant"`
	Context        Context        `json:"context"`
}

type ReceiverTenant struct {
	State                string          `json:"state"`
	DeploymentRegion     string          `json:"deploymentRegion"`
	ApplicationTenantID  string          `json:"applicationTenantId"`
	SubaccountID         string          `json:"subaccountId"`
	ApplicationNamespace string          `json:"applicationNamespace"`
	Subdomain            string          `json:"subdomain"`
	Configuration        json.RawMessage `json:"configuration"`
}

type AssignedTenant struct {
	State                string          `json:"state"`
	ApplicationNamespace string          `json:"applicationNamespace"`
	Configuration        json.RawMessage `json:"configuration"`
}

type Context struct {
	Operation   string `json:"operation"`
	FormationID string `json:"uclFormationId"`
}
