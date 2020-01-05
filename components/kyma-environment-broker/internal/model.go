package internal

type Instance struct {
	InstanceID      string
	RuntimeID       string
	GlobalAccountID string
	ServiceID       string
	ServicePlanID   string

	DashboardURL string

	ProvisioningParameters *ProvisioningParameters
}

type ProvisioningParameters struct {
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
