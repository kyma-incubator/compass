package types

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/paths"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
)

const (
	// ServiceInstancesType represents the type of the ServiceInstances struct; used primarily for logging purposes
	ServiceInstancesType = "service instances"
	// ServiceInstanceType represents the type of the ServiceInstance struct; used primarily for logging purposes
	ServiceInstanceType = "service instance"
)

// ServiceInstanceReqBody is the request body when a Service Instance is being created
type ServiceInstanceReqBody struct {
	Name          string          `json:"name"`
	ServicePlanID string          `json:"service_plan_id"`
	Parameters    json.RawMessage `json:"parameters,omitempty"` // todo::: differs from service to service. Most probably the necessary data will be provided as arbitrary json in the TM notification body?
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

// ServiceInstances represents a collection of Service Instance
type ServiceInstances struct {
	NumItems int                `json:"num_items"`
	Items    []*ServiceInstance `json:"items"`
}

// Match matches a ServiceInstance based on some criteria
func (si *ServiceInstances) Match(params resources.ResourceMatchParameters) (string, error) {
	serviceInstanceParams, ok := params.(*ServiceInstanceMatchParameters)
	if !ok {
		return "", errors.New("while type asserting ResourceMatchParameters to ServiceInstanceMatchParameters")
	}

	serviceInstanceName := serviceInstanceParams.ServiceInstanceName
	for _, item := range si.Items {
		if item.Name == serviceInstanceName {
			return item.ID, nil
		}
	}
	return "", nil // for ServiceInstances we don't want to fail if nothing is found
}

// MatchMultiple matches several ServiceInstances based on some criteria
func (si *ServiceInstances) MatchMultiple(args resources.ResourceMatchParameters) []string {
	return nil // implement me when needed
}

// GetType gets the type of the ServiceInstances
func (si *ServiceInstances) GetType() string {
	return ServiceInstancesType
}

// ServiceInstanceMatchParameters holds all the necessary fields that are used when matching ServiceInstances
type ServiceInstanceMatchParameters struct {
	ServiceInstanceName string
}

// GetURLPath gets the URL Path of the ServiceInstance
func (sia *ServiceInstanceMatchParameters) GetURLPath() string {
	return paths.ServiceInstancesPath
}
