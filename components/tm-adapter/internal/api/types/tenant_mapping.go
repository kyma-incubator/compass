package types

import "encoding/json"

type TenantMapping struct {
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
	AssignedTenant AssignedTenant `json:"assignedTenant"`
	Context        Context        `json:"context"`
}

type ReceiverTenant struct {
	ApplicationURL      string `json:"applicationUrl"`
	ApplicationTenantID string `json:"applicationTenantId"`
	DeploymentRegion    string `json:"deploymentRegion"`
}

type AssignedTenant struct {
	ApplicationURL string          `json:"applicationUrl"`
	Configuration  json.RawMessage `json:"configuration"`
}

type Context struct {
	Operation string `json:"operation"`
}
