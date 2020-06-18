package internal

import (
	"reflect"
)

type ProvisioningParameters struct {
	PlanID     string                    `json:"plan_id"`
	ServiceID  string                    `json:"service_id"`
	ErsContext ERSContext                `json:"ers_context"`
	Parameters ProvisioningParametersDTO `json:"parameters"`

	// PlatformRegion defines the Platform region send in the request path, terminology:
	//  - `Platform` is a place where KEB is registered and which later sends request to KEB.
	//  - `Region` value is use e.g. for billing integration such as EDP.
	PlatformRegion string `json:"platform_region"`
}

func (p ProvisioningParameters) IsEqual(input ProvisioningParameters) bool {
	if p.PlanID != input.PlanID {
		return false
	}
	if p.ServiceID != input.ServiceID {
		return false
	}

	if !reflect.DeepEqual(p.ErsContext, input.ErsContext) {
		return false
	}

	// TODO: If there are already not resolved operations in the db
	//       KEB will raise an error "provisioning operation already
	//       exist" after updating the image on the environments.
	//       Revert this check after some time, when every operation
	//       has at least the default Gardener Shoot Purpose to:
	//       if !reflect.DeepEqual(p.Parameters, input.Parameters) {
	//          ...
	if !temporaryDeepEqualWithoutPurpose(p.Parameters, input.Parameters) {
		return false
	}

	return true
}

type ProvisioningParametersDTO struct {
	Name                        string   `json:"name"`
	TargetSecret                *string  `json:"targetSecret"`
	VolumeSizeGb                *int     `json:"volumeSizeGb"`
	MachineType                 *string  `json:"machineType"`
	Region                      *string  `json:"region"`
	Purpose                     *string  `json:"purpose"`
	Zones                       []string `json:"zones"`
	AutoScalerMin               *int     `json:"autoScalerMin"`
	AutoScalerMax               *int     `json:"autoScalerMax"`
	MaxSurge                    *int     `json:"maxSurge"`
	MaxUnavailable              *int     `json:"maxUnavailable"`
	OptionalComponentsToInstall []string `json:"components"`
	KymaVersion                 string   `json:"kymaVersion"`
}

type ERSContext struct {
	TenantID        string                  `json:"tenant_id"`
	SubAccountID    string                  `json:"subaccount_id"`
	GlobalAccountID string                  `json:"globalaccount_id"`
	ServiceManager  *ServiceManagerEntryDTO `json:"sm_platform_credentials,omitempty"`
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

func temporaryDeepEqualWithoutPurpose(a0, a1 ProvisioningParametersDTO) bool {
	return a0.Name == a1.Name &&
		a0.KymaVersion == a1.KymaVersion &&
		areStrPtrEqual(a0.TargetSecret, a1.TargetSecret) &&
		areIntPtrEqual(a0.VolumeSizeGb, a1.VolumeSizeGb) &&
		areStrPtrEqual(a0.MachineType, a1.MachineType) &&
		areStrPtrEqual(a0.Region, a1.Region) &&
		areStrArrEqual(a0.Zones, a1.Zones) &&
		areIntPtrEqual(a0.AutoScalerMin, a1.AutoScalerMin) &&
		areIntPtrEqual(a0.AutoScalerMax, a1.AutoScalerMax) &&
		areIntPtrEqual(a0.MaxSurge, a1.MaxSurge) &&
		areIntPtrEqual(a0.MaxUnavailable, a1.MaxUnavailable) &&
		areStrArrEqual(a0.OptionalComponentsToInstall, a1.OptionalComponentsToInstall)
}

func areStrPtrEqual(a0, a1 *string) bool {
	if (a0 == nil && a1 != nil) || (a0 != nil && a1 == nil) {
		return false
	}
	if a0 != nil && a1 != nil && *a0 != *a1 {
		return false
	}
	return true
}
func areIntPtrEqual(a0, a1 *int) bool {
	if (a0 == nil && a1 != nil) || (a0 != nil && a1 == nil) {
		return false
	}
	if a0 != nil && a1 != nil && *a0 != *a1 {
		return false
	}
	return true
}
func areStrArrEqual(a0, a1 []string) bool {
	return reflect.DeepEqual(a0, a1)
}
