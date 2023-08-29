package types

import "encoding/json"

// ServiceInstanceReqBody is the request body when a Service Instance is being created
type ServiceInstanceReqBody struct {
	Name          string          `json:"name"`
	ServicePlanID string          `json:"service_plan_id"`
	Parameters    json.RawMessage `json:"parameters,omitempty"` // todo::: differs from service to service. Most probably the necessary data will be provided as arbitrary json in the TM notification body?
}

// ServiceInstance represents a Service Instance
type ServiceInstance struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ServicePlanID string `json:"service_plan_id"`
	PlatformID    string `json:"platform_id"`
}

// ServiceInstances represents a collection of Service Instance
type ServiceInstances struct {
	NumItems int               `json:"num_items"`
	Items    []ServiceInstance `json:"items"`
}
