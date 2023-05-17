package types

import (
	"encoding/json"
)

type ServiceKeyReqBody struct {
	Name              string          `json:"name"`
	ServiceInstanceId string          `json:"service_instance_id"`
	Parameters        json.RawMessage `json:"parameters,omitempty"` // todo::: differs from service to service. Most probably the necessary data will be provided as arbitrary json in the TN notification body?
}

type ServiceKey struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	ServiceInstanceId string          `json:"service_instance_id"`
	Credentials       json.RawMessage `json:"credentials"`
}

type IASParameters struct {
	ConsumedServices      []ConsumedService `json:"consumed-services"`
	XsuaaCrossConsumption bool              `json:"xsuaa-cross-consumption"`
}

type ConsumedService struct {
	ServiceInstanceName string `json:"service-instance-name"`
}

type ServiceKeys struct {
	NumItems int          `json:"num_items"`
	Items    []ServiceKey `json:"items"`
}
