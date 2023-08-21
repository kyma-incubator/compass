package types

import "encoding/json"

type TenantMapping struct {
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
	AssignedTenant AssignedTenant `json:"assignedTenant"`
	Context        Context        `json:"context"`
}

type ReceiverTenant struct {
	ApplicationURL string          `json:"applicationUrl"`
	State          string          `json:"state"`
	SubaccountID   string          `json:"applicationTenantId"`
	Configuration  json.RawMessage `json:"configuration"`
}

type AssignedTenant struct {
	ApplicationURL string          `json:"applicationUrl"`
	State          string          `json:"state"`
	Configuration  json.RawMessage `json:"configuration"`
}

type Context struct {
	Operation string `json:"operation"`
}
