package types

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/paths"
)

const (
	// ServiceBindingsType represents the type of the ServiceBindings struct; used primarily for logging purposes
	ServiceBindingsType = "service bindings"
	// ServiceBindingType represents the type of the ServiceBinding struct; used primarily for logging purposes
	ServiceBindingType = "service binding"
)

// TODO:: rename key->binding
// ServiceKeyReqBody is the request body when a Service Key is being created
type ServiceKeyReqBody struct { // TODO:: check if it needs also 'bind_resource' in the request body
	Name         string          `json:"name"`
	ServiceKeyID string          `json:"service_instance_id"`
	Parameters   json.RawMessage `json:"parameters,omitempty"` // TODO:: differs from service to service. Most probably the necessary data will be provided as arbitrary json in the TN notification body?
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

// GetResourceName gets the ServiceKey Name
func (s *ServiceKey) GetResourceName() string {
	return s.Name
}

// ServiceKeys represents a collection of Service Key
type ServiceKeys struct {
	NumItems int           `json:"num_items"`
	Items    []*ServiceKey `json:"items"`
}

// GetType gets the type of the ServiceKeys
func (sk *ServiceKeys) GetType() string {
	return ServiceBindingsType
}

// GetURLPath gets the URL Path of the ServiceKey
func (sk *ServiceKeys) GetURLPath() string {
	return paths.ServiceBindingsPath
}

// GetIDs gets the IDs of all ServiceKeys
func (sks *ServiceKeys) GetIDs() []string {
	ids := make([]string, 0, sks.NumItems)
	for _, sk := range sks.Items {
		ids = append(ids, sk.ID)
	}
	return ids
}

// ServiceKeyMatchParameters holds all the necessary fields that are used when matching ServiceKeys
type ServiceKeyMatchParameters struct {
	ServiceInstanceID   string
	ServiceInstancesIDs []string
}

// Match matches a ServiceKey based on some criteria
func (skp *ServiceKeyMatchParameters) Match(resources resources.Resources) (string, error) {
	return "", nil // implement me when needed
}

// MatchMultiple matches several ServiceKeys based on some criteria
func (skp *ServiceKeyMatchParameters) MatchMultiple(resources resources.Resources) ([]string, error) {
	serviceKeys, ok := resources.(*ServiceKeys)
	if !ok {
		return nil, errors.New("while type asserting Resources to ServiceKeys")
	}
	serviceKeyIDs := make([]string, 0, serviceKeys.NumItems)
	if skp.ServiceInstanceID != "" {
		for _, sk := range serviceKeys.Items {
			if sk.ServiceInstanceID == skp.ServiceInstanceID {
				serviceKeyIDs = append(serviceKeyIDs, sk.ID)
			}
		}
	}
	if len(skp.ServiceInstancesIDs) > 0 {
		// Add all service bindings for all skp.serviceInstancesIDs
		for _, serviceInstanceID := range skp.ServiceInstancesIDs {
			for _, sk := range serviceKeys.Items {
				if sk.ServiceInstanceID == serviceInstanceID {
					serviceKeyIDs = append(serviceKeyIDs, sk.ID)
				}
			}
		}
	}
	return serviceKeyIDs, nil
}
