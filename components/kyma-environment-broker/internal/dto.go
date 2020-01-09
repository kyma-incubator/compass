package internal

type ProvisioningParametersDTO struct {
	Name           string  `json:"name"`
	NodeCount      *int    `json:"nodeCount"`
	VolumeSizeGb   *int    `json:"volumeSizeGb"`
	MachineType    *string `json:"machineType"`
	Region         *string `json:"region"`
	Zone           *string `json:"zone"`
	AutoScalerMin  *int    `json:"autoScalerMin"`
	AutoScalerMax  *int    `json:"autoScalerMax"`
	MaxSurge       *int    `json:"maxSurge"`
	MaxUnavailable *int    `json:"maxUnavailable"`
}

type ERSContext struct {
	TenantID        string `json:"tenant_id"`
	SubAccountID    string `json:"subaccount_id"`
	GlobalAccountID string `json:"globalaccount_id"`
}
