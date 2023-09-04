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
func (s ServicePlan) GetResourceID() string {
	return s.ID
}

// GetResourceType gets the return type of the ServicePlan
func (s ServicePlan) GetResourceType() string {
	return ServicePlanType
}

// GetResourceURLPath gets the ServicePlan URL Path
func (s ServicePlan) GetResourceURLPath() string {
	return paths.ServicePlansPath
}

// ServicePlans represents a collection of Service Plan
type ServicePlans struct {
	NumItems int           `json:"num_items"`
	Items    []ServicePlan `json:"items"`
}

// Match matches a ServicePlan based on some criteria
func (sp ServicePlans) Match(params resources.ResourceMatchParameters) (string, error) {
	servicePlanParams, ok := params.(ServicePlanMatchParameters)
	if !ok {
		return "", errors.New("while type asserting ResourceMatchParameters to ServicePlanMatchParameters")
	}

	planName := servicePlanParams.PlanName
	offeringID := servicePlanParams.OfferingID

	for _, item := range sp.Items {
		if item.CatalogName == planName && item.ServiceOfferingId == offeringID {
			return item.ID, nil
		}
	}
	return "", errors.Errorf("couldn't find service plan for catalog name: %s and offering ID: %s", planName, offeringID)
}

// MatchMultiple matches several ServicePlans based on some criteria
func (sp ServicePlans) MatchMultiple(params resources.ResourceMatchParameters) ([]string, error) {
	return nil, nil // implement me when needed
}

// GetType gets the type of the ServicePlans
func (sp ServicePlans) GetType() string {
	return ServicePlansType
}

// ServicePlanMatchParameters holds all the necessary fields that are used when matching ServicePlans
type ServicePlanMatchParameters struct {
	PlanName   string
	OfferingID string
}

// GetURLPath gets the URL Path of the ServicePlan
func (spa ServicePlanMatchParameters) GetURLPath() string {
	return paths.ServicePlansPath
}
