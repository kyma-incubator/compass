package types

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/paths"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
)

const (
	// ServiceBindingsType represents the type of the ServiceBindings struct; used primarily for logging purposes
	ServiceBindingsType = "service bindings"
	// ServiceBindingType represents the type of the ServiceBinding struct; used primarily for logging purposes
	ServiceBindingType = "service binding"
)

// ServiceKeyReqBody is the request body when a Service Key is being created
type ServiceKeyReqBody struct {
	Name         string          `json:"name"`
	ServiceKeyID string          `json:"service_instance_id"`
	Parameters   json.RawMessage `json:"parameters,omitempty"` // todo::: differs from service to service. Most probably the necessary data will be provided as arbitrary json in the TN notification body?
}

// GetResourceName gets the ServiceKey name from the request body
func (rb *ServiceKeyReqBody) GetResourceName() string {
	return rb.Name
}

// ServiceKey represents a Service Key
type ServiceKey struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	ServiceInstanceID string          `json:"service_instance_id"`
	Credentials       json.RawMessage `json:"credentials"`
}

// GetResourceID gets the ServiceKey ID
func (s *ServiceKey) GetResourceID() string {
	return s.ID
}

// GetResourceType gets the return type of the ServiceKey
func (s *ServiceKey) GetResourceType() string {
	return ServiceBindingType
}

// GetResourceURLPath gets the ServiceKey URL Path
func (s *ServiceKey) GetResourceURLPath() string {
	return paths.ServiceBindingsPath
}

// ServiceKeys represents a collection of Service Key
type ServiceKeys struct {
	NumItems int           `json:"num_items"`
	Items    []*ServiceKey `json:"items"`
}

// Match matches a ServiceKey based on some criteria
func (sk *ServiceKeys) Match(params resources.ResourceMatchParameters) (string, error) {
	return "", nil // implement me when needed
}

// MatchMultiple matches several ServiceKeys based on some criteria
func (sk *ServiceKeys) MatchMultiple(params resources.ResourceMatchParameters) ([]string, error) {
	serviceKeyParams, ok := params.(*ServiceKeyMatchParameters)
	if !ok {
		return nil, errors.New("while type asserting ResourceMatchParameters to ServiceKeyMatchParameters")
	}
	serviceKeyIDs := make([]string, 0, sk.NumItems)
	for _, item := range sk.Items {
		if item.ServiceInstanceID == serviceKeyParams.ServiceInstanceID {
			serviceKeyIDs = append(serviceKeyIDs, item.ID)
		}
	}
	return serviceKeyIDs, nil
}

// GetType gets the type of the ServiceKeys
func (sk *ServiceKeys) GetType() string {
	return ServiceBindingsType
}

// ServiceKeyMatchParameters holds all the necessary fields that are used when matching ServiceKeys
type ServiceKeyMatchParameters struct {
	ServiceInstanceID string
}

// GetURLPath gets the URL Path of the ServiceKey
func (ska *ServiceKeyMatchParameters) GetURLPath() string {
	return paths.ServiceBindingsPath
}
