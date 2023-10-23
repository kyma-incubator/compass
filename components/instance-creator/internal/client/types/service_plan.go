package types

import (
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/paths"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
	"github.com/pkg/errors"
)

const (
	// ServicePlansType represents the type of the ServicePlans struct; used primarily for logging purposes
	ServicePlansType = "service plans"
	// ServicePlanType represents the type of the ServicePlan struct; used primarily for logging purposes
	ServicePlanType = "service plan"
)

// ServicePlan represents a Service Plan
type ServicePlan struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	CatalogId         string `json:"catalog_id"`
	CatalogName       string `json:"catalog_name"`
	ServiceOfferingId string `json:"service_offering_id"`
}

// GetResourceID gets the ServicePlan ID
func (s *ServicePlan) GetResourceID() string {
	return s.ID
}

// GetResourceType gets the return type of the ServicePlan
func (s *ServicePlan) GetResourceType() string {
	return ServicePlanType
}

// GetResourceURLPath gets the ServicePlan URL Path
func (s *ServicePlan) GetResourceURLPath() string {
	return paths.ServicePlansPath
}

// ServicePlans represents a collection of Service Plan
type ServicePlans struct {
	NumItems int            `json:"num_items"`
	Items    []*ServicePlan `json:"items"`
}

// GetType gets the type of the ServicePlans
func (sp *ServicePlans) GetType() string {
	return ServicePlansType
}

// GetURLPath gets the URL Path of the ServicePlan
func (sp *ServicePlans) GetURLPath() string {
	return paths.ServicePlansPath
}

// ServicePlanMatchParameters holds all the necessary fields that are used when matching ServicePlans
type ServicePlanMatchParameters struct {
	PlanName   string
	OfferingID string
}

// Match matches a ServicePlan based on some criteria
func (spp *ServicePlanMatchParameters) Match(resources resources.Resources) (string, error) {
	servicePlans, ok := resources.(*ServicePlans)
	if !ok {
		return "", errors.New("while type asserting Resources to ServicePlans")
	}

	for _, sp := range servicePlans.Items {
		if sp.CatalogName == spp.PlanName && sp.ServiceOfferingId == spp.OfferingID {
			return sp.ID, nil
		}
	}
	return "", errors.Errorf("couldn't find service plan for catalog name: %s and offering ID: %s", spp.PlanName, spp.OfferingID)
}

// MatchMultiple matches several ServicePlans based on some criteria
func (spp *ServicePlanMatchParameters) MatchMultiple(resources resources.Resources) ([]string, error) {
	return nil, nil // implement me when needed
}
