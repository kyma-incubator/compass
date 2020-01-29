package internal

import "time"

type Instance struct {
	InstanceID      string
	RuntimeID       string
	GlobalAccountID string
	ServiceID       string
	ServicePlanID   string

	DashboardURL           string
	ProvisioningParameters string

	CreatedAt time.Time
	UpdatedAt time.Time
	DelatedAt time.Time
}
