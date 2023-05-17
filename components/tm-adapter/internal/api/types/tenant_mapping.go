package types

import "encoding/json"

type TenantMapping struct {
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
	Items          []Item         `json:"items"`
}

type ReceiverTenant struct {
	SubaccountID string `json:"subaccountId"`
}

type Item struct {
	Operation      string          `json:"operation"`
	ApplicationURL string          `json:"applicationUrl"`
	Configuration  json.RawMessage `json:"configuration"`
}
