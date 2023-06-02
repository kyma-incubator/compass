package types

import "encoding/json"

type ServiceInstanceReqBody struct {
	Name          string          `json:"name"`
	ServicePlanId string          `json:"service_plan_id"`
	Parameters    json.RawMessage `json:"parameters,omitempty"` // todo::: differs from service to service. Most probably the necessary data will be provided as arbitrary json in the TN notification body?
}

type ServiceInstance struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ServicePlanId string `json:"service_plan_id"`
	PlatformId    string `json:"platform_id"`
}

type ServiceInstances struct {
	NumItems int               `json:"num_items"`
	Items    []ServiceInstance `json:"items"`
}

type IASParameters struct {
	ConsumedServices      []ConsumedService `json:"consumed-services"`
	XsuaaCrossConsumption bool              `json:"xsuaa-cross-consumption"`
}
