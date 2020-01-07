package internal

type Instance struct {
	InstanceID      string
	RuntimeID       string
	GlobalAccountID string
	ServiceID       string
	ServicePlanID   string

	DashboardURL           string
	ProvisioningParameters string
}
