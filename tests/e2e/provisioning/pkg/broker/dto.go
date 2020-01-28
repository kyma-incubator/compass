package broker

type ERSContext struct {
	TenantID        string `json:"tenant_id"`
	SubAccountID    string `json:"subaccount_id"`
	GlobalAccountID string `json:"globalaccount_id"`
}

type ProvisionResponse struct {
	Operation string `json:"operation"`
}
