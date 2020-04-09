package appinfo

import "time"

type (
	RuntimeDTO struct {
		RuntimeID         string    `json:"runtimeId"`
		GlobalAccountID   string    `json:"globalAccountId"`
		SubAccountID      string    `json:"subaccountId"`
		ServiceInstanceID string    `json:"serviceInstanceId"`
		ServiceClassID    string    `json:"serviceClassId"`
		ServiceClassName  string    `json:"serviceClassName"`
		ServicePlanID     string    `json:"servicePlanId"`
		ServicePlanName   string    `json:"servicePlanName"`
		Status            StatusDTO `json:"status"`
	}

	StatusDTO struct {
		CreatedAt      *time.Time          `json:"createdAt,omitempty"`
		UpdatedAt      *time.Time          `json:"updatedAt,omitempty"`
		DeletedAt      *time.Time          `json:"deletedAt,omitempty"`
		Provisioning   *OperationStatusDTO `json:"provisioning,omitempty"`
		Deprovisioning *OperationStatusDTO `json:"deprovisioning,omitempty"`
	}

	OperationStatusDTO struct {
		State       string `json:"state"`
		Description string `json:"description"`
	}
)
