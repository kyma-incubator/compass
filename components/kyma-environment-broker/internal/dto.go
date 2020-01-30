package internal

type ProvisioningParametersDTO struct {
	Name                        string   `json:"name"`
	NodeCount                   *int     `json:"nodeCount"`
	VolumeSizeGb                *int     `json:"volumeSizeGb"`
	MachineType                 *string  `json:"machineType"`
	Region                      *string  `json:"region"`
	Zone                        *string  `json:"zone"`
	AutoScalerMin               *int     `json:"autoScalerMin"`
	AutoScalerMax               *int     `json:"autoScalerMax"`
	MaxSurge                    *int     `json:"maxSurge"`
	MaxUnavailable              *int     `json:"maxUnavailable"`
	OptionalComponentsToInstall []string `json:"components"`
}

type ERSContext struct {
	TenantID        string                 `json:"tenant_id"`
	SubAccountID    string                 `json:"subaccount_id"`
	GlobalAccountID string                 `json:"globalaccount_id"`
	ServiceManager  ServiceManagerEntryDTO `json:"sm_platform_credentials"`
}

type ServiceManagerEntryDTO struct {
	Credentials ServiceManagerCredentials `json:"credentials"`
	URL         string                    `json:"url"`
}

type ServiceManagerCredentials struct {
	BasicAuth ServiceManagerBasicAuth `json:"basic"`
}

type ServiceManagerBasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
