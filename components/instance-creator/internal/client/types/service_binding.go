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

// ServiceBindingReqBody is the request body when a Service Bjd] dj g is being created
type ServiceBindingReqBody struct {
	Name             string          `json:"name"`
	ServiceBindingID string          `json:"service_instance_id"`
	Parameters       json.RawMessage `json:"parameters,omitempty"`
}

// GetResourceName gets the ServiceBinding name from the request body
func (rb *ServiceBindingReqBody) GetResourceName() string {
	return rb.Name
}

// ServiceBinding represents a Service Binding
type ServiceBinding struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	ServiceInstanceID string          `json:"service_instance_id"`
	Credentials       json.RawMessage `json:"credentials"`
}

// GetResourceID gets the ServiceBinding ID
func (s *ServiceBinding) GetResourceID() string {
	return s.ID
}

// GetResourceType gets the return type of the ServiceBinding
func (s *ServiceBinding) GetResourceType() string {
	return ServiceBindingType
}

// GetResourceURLPath gets the ServiceBinding URL Path
func (s *ServiceBinding) GetResourceURLPath() string {
	return paths.ServiceBindingsPath
}

// GetResourceName gets the ServiceBinding Name
func (s *ServiceBinding) GetResourceName() string {
	return s.Name
}

// ServiceBindings represents a collection of Service Binding
type ServiceBindings struct {
	NumItems int               `json:"num_items"`
	Items    []*ServiceBinding `json:"items"`
}

// GetType gets the type of the ServiceBindings
func (sbs *ServiceBindings) GetType() string {
	return ServiceBindingsType
}

// GetURLPath gets the URL Path of the ServiceBinding
func (sbs *ServiceBindings) GetURLPath() string {
	return paths.ServiceBindingsPath
}

// GetIDs gets the IDs of all ServiceBindings
func (sbs *ServiceBindings) GetIDs() []string {
	ids := make([]string, 0, sbs.NumItems)
	for _, sb := range sbs.Items {
		ids = append(ids, sb.ID)
	}
	return ids
}

// ServiceBindingMatchParameters holds all the necessary fields that are used when matching ServiceBindings
type ServiceBindingMatchParameters struct {
	ServiceInstanceID   string
	ServiceInstancesIDs []string
}

// Match matches a ServiceBinding based on some criteria
func (sbp *ServiceBindingMatchParameters) Match(resources resources.Resources) (string, error) {
	return "", nil // implement me when needed
}

// MatchMultiple matches several ServiceBindings based on some criteria
func (sbp *ServiceBindingMatchParameters) MatchMultiple(resources resources.Resources) ([]string, error) {
	serviceBindings, ok := resources.(*ServiceBindings)
	if !ok {
		return nil, errors.New("while type asserting Resources to ServiceBindings")
	}
	serviceBindingIDs := make([]string, 0, serviceBindings.NumItems)
	if sbp.ServiceInstanceID != "" {
		for _, sb := range serviceBindings.Items {
			if sb.ServiceInstanceID == sbp.ServiceInstanceID {
				serviceBindingIDs = append(serviceBindingIDs, sb.ID)
			}
		}
	}
	if len(sbp.ServiceInstancesIDs) > 0 {
		// Add all service bindings for all sbp.serviceInstancesIDs
		for _, serviceInstanceID := range sbp.ServiceInstancesIDs {
			for _, sb := range serviceBindings.Items {
				if sb.ServiceInstanceID == serviceInstanceID {
					serviceBindingIDs = append(serviceBindingIDs, sb.ID)
				}
			}
		}
	}
	return serviceBindingIDs, nil
}
