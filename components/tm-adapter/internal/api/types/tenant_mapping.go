package types

type TenantMapping struct {
	Context TenantMappingContext `json:"context"`
	Items   []TenantMappingItem  `json:"items"`
}

type TenantMappingContext struct {
	UclFormationId string `json:"uclFormationId"`
	AccountId      string `json:"accountId"`
	CrmId          string `json:"crmId"`
	Platform       string `json:"platform"`
}

type TenantMappingItem struct {
	Operation            string        `json:"operation"`
	UclAssignmentId      string        `json:"uclAssignmentId"`
	DeploymentRegion     string        `json:"deploymentRegion"`
	ApplicationNamespace string        `json:"applicationNamespace"`
	ApplicationUrl       string        `json:"applicationUrl"`
	ApplicationTenantId  string        `json:"applicationTenantId"`
	AssigneeReady        bool          `json:"assigneeReady"`
	UclSystemTenantId    string        `json:"uclSystemTenantId"`
	//Configuration        Configuration `json:"configuration"`
	Configuration        any `json:"configuration"`
}

type Configuration struct {
	Destinations []Destination `json:"destinations"`
	Credentials  Credentials   `json:"credentials"`
}

type Destination struct {
	Name              string `json:"name"`
	SubaccountId      string `json:"subaccountId"`
	ServiceInstanceId string `json:"serviceInstanceId"`
}

type Credentials struct {
	InboundCommunication InboundCommunication `json:"inboundCommunication"`
}

type InboundCommunication struct {
	BasicAuthentication BasicAuthentication `json:"basicAuthentication"`
}

type BasicAuthentication struct {
	Destinations []Destination `json:"destinations"`
}
