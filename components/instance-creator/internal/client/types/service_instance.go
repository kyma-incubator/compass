package types

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/paths"
)

const (
	// ServiceInstancesType represents the type of the ServiceInstances struct; used primarily for logging purposes
	ServiceInstancesType = "service instances"
	// ServiceInstanceType represents the type of the ServiceInstance struct; used primarily for logging purposes
	ServiceInstanceType = "service instance"
)

// ServiceInstanceReqBody is the request body when a Service Instance is being created
type ServiceInstanceReqBody struct {
	Name          string              `json:"name"`
	ServicePlanID string              `json:"service_plan_id"`
	Parameters    json.RawMessage     `json:"parameters,omitempty"` // TODO:: differs from service to service. Most probably the necessary data will be provided as arbitrary json in the TM notification body?
	Labels        map[string][]string `json:"labels,omitempty"`
}

// GetResourceName gets the ServiceInstance name from the request body
func (rb *ServiceInstanceReqBody) GetResourceName() string {
	return rb.Name
}

// ServiceInstance represents a Service Instance
type ServiceInstance struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ServicePlanID string `json:"service_plan_id"`
	PlatformID    string `json:"platform_id"`
}

// GetResourceID gets the ServiceInstance ID
func (s *ServiceInstance) GetResourceID() string {
	return s.ID
}

// GetResourceType gets the return type of the ServiceInstance
func (s *ServiceInstance) GetResourceType() string {
	return ServiceInstanceType
}

// GetResourceURLPath gets the ServiceInstance URL Path
func (s *ServiceInstance) GetResourceURLPath() string {
	return paths.ServiceInstancesPath
}

// GetResourceName gets the Service
func (s *ServiceInstance) GetResourceName() string {
	return s.Name
}

// ServiceInstances represents a collection of Service Instance
type ServiceInstances struct {
	NumItems int                `json:"num_items"`
	Items    []*ServiceInstance `json:"items"`
}

// GetType gets the type of the ServiceInstances
func (si *ServiceInstances) GetType() string {
	return ServiceInstancesType
}

// GetURLPath gets the URL Path of the ServiceInstance
func (si *ServiceInstances) GetURLPath() string {
	return paths.ServiceInstancesPath
}

// GetIDs gets the IDs of all ServiceInstances
func (sis *ServiceInstances) GetIDs() []string {
	ids := make([]string, 0, sis.NumItems)
	for _, si := range sis.Items {
		ids = append(ids, si.ID)
	}
	return ids
}

// ServiceInstanceMatchParameters holds all the necessary fields that are used when matching ServiceInstances
type ServiceInstanceMatchParameters struct {
	ServiceInstanceName string
}

// Match matches a ServiceInstance based on some criteria
func (sip *ServiceInstanceMatchParameters) Match(resources resources.Resources) (string, error) {
	serviceInstances, ok := resources.(*ServiceInstances)
	if !ok {
		return "", errors.New("while type asserting Resources to ServiceInstances")
	}

	for _, si := range serviceInstances.Items {
		if si.Name == sip.ServiceInstanceName {
			return si.ID, nil
		}
	}
	return "", nil // for ServiceInstances we don't want to fail if nothing is found
}

// MatchMultiple matches several ServiceInstances based on some criteria
func (sip *ServiceInstanceMatchParameters) MatchMultiple(resources resources.Resources) ([]string, error) {
	return nil, nil // implement me when needed
}
