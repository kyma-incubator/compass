package types

import (
	"encoding/json"
)

// ServiceKeyReqBody is the request body when a Service Key is being created
type ServiceKeyReqBody struct {
	Name              string          `json:"name"`
	ServiceInstanceID string          `json:"service_instance_id"`
	Parameters        json.RawMessage `json:"parameters,omitempty"` // todo::: differs from service to service. Most probably the necessary data will be provided as arbitrary json in the TN notification body?
}

// ServiceKey represents a Service Key
type ServiceKey struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	ServiceInstanceID string          `json:"service_instance_id"`
	Credentials       json.RawMessage `json:"credentials"`
}

// ServiceKeys represents a collection of Service Key
type ServiceKeys struct {
	NumItems int          `json:"num_items"`
	Items    []ServiceKey `json:"items"`
}
