package types

import (
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/paths"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
	"github.com/pkg/errors"
)

const (
	// ServiceOfferingsType represents the type of the ServiceOfferings struct; used primarily for logging purposes
	ServiceOfferingsType = "service offerings"
	// ServiceOfferingType represents the type of the ServiceOffering struct; used primarily for logging purposes
	ServiceOfferingType = "service offering"
)

// ServiceOffering represents a Service Offering
type ServiceOffering struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BrokerId    string `json:"broker_id"`
	CatalogId   string `json:"catalog_id"`
	CatalogName string `json:"catalog_name"`
}

// GetResourceID gets the ServiceOffering ID
func (s *ServiceOffering) GetResourceID() string {
	return s.ID
}

// GetResourceType gets the return type of the ServiceOffering
func (s *ServiceOffering) GetResourceType() string {
	return ServiceOfferingType
}

// GetResourceURLPath gets the ServiceOffering URL Path
func (s *ServiceOffering) GetResourceURLPath() string {
	return paths.ServiceOfferingsPath
}

// ServiceOfferings represents a collection of Service Offering
type ServiceOfferings struct {
	NumItems int                `json:"num_items"`
	Items    []*ServiceOffering `json:"items"`
}

// GetType gets the type of the ServiceOfferings
func (so *ServiceOfferings) GetType() string {
	return ServiceOfferingsType
}

func (so *ServiceOfferings) GetURLPath() string {
	return paths.ServiceOfferingsPath
}

// ServiceOfferingMatchParameters holds all the necessary fields that are used when matching ServiceOfferings
type ServiceOfferingMatchParameters struct {
	CatalogName string
}

// Match matches a ServiceOffering based on some criteria
func (sp *ServiceOfferingMatchParameters) Match(resources resources.Resources) (string, error) {
	serviceOfferings, ok := resources.(*ServiceOfferings)
	if !ok {
		return "", errors.New("while type asserting Resources to ServiceOfferings")
	}

	for _, so := range serviceOfferings.Items {
		if so.CatalogName == sp.CatalogName {
			return so.ID, nil
		}
	}
	return "", errors.Errorf("couldn't find service offering for catalog name: %s", sp.CatalogName)
}

// MatchMultiple matches several ServiceOfferings based on some criteria
func (sp *ServiceOfferingMatchParameters) MatchMultiple(resources resources.Resources) ([]string, error) {
	return nil, nil // implement me when needed
}
