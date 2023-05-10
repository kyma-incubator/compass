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
	Configuration json.RawMessage `json:"configuration"`
}
