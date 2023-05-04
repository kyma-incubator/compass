package types

type TenantMapping struct {
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
}

type ReceiverTenant struct {
	SubaccountID         string `json:"subaccountId"`
}
