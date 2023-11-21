package types

import (
	"encoding/json"
)

type ServiceKeyReqBody struct {
	Name              string          `json:"name"`
	ServiceInstanceId string          `json:"service_instance_id"`
	Parameters        json.RawMessage `json:"parameters,omitempty"`
}

type ServiceKey struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	ServiceInstanceId string          `json:"service_instance_id"`
	Credentials       json.RawMessage `json:"credentials"`
}

type ServiceKeys struct {
	NumItems int          `json:"num_items"`
	Items    []ServiceKey `json:"items"`
}
