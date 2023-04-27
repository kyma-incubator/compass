package types

type ServiceInstanceReqBody struct {
	Name          string `json:"name"`
	ServicePlanId string `json:"service_plan_id"`
}

type ServiceInstance struct {
	Id            string                 `json:"id"`
	Name          string                 `json:"name"`
	ServicePlanId string                 `json:"service_plan_id"`
	PlatformId    string                 `json:"platform_id"`
}
