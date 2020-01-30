package broker

type ersContext struct {
	TenantID        string `json:"tenant_id"`
	SubAccountID    string `json:"subaccount_id"`
	GlobalAccountID string `json:"globalaccount_id"`
}

type provisionResponse struct {
	Operation string `json:"operation"`
}

type lastOperationResponse struct {
	State string `json:"state"`
}

type instanceDetailsResponse struct {
	DashboardURL string `json:"dashboard_url"`
}

type provisionParameters struct {
	Name       string   `json:"name"`
	Components []string `json:"components"`
}
