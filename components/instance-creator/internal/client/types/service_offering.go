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
func (s ServiceOffering) GetResourceID() string {
	return s.ID
}

// GetResourceType gets the return type of the ServiceOffering
func (s ServiceOffering) GetResourceType() string {
	return ServiceOfferingType
}

// GetResourceURLPath gets the ServiceOffering URL Path
func (s ServiceOffering) GetResourceURLPath() string {
	return paths.ServiceOfferingsPath
}

// ServiceOfferings represents a collection of Service Offering
type ServiceOfferings struct {
	NumItems int               `json:"num_items"`
	Items    []ServiceOffering `json:"items"`
}

// Match matches a ServiceOffering based on some criteria
func (so ServiceOfferings) Match(args resources.ResourceArguments) (string, error) {
	catalogName := args.(ServiceOfferingArguments).CatalogName
	for _, item := range so.Items {
		if item.CatalogName == catalogName {
			return item.ID, nil
		}
	}
	return "", errors.Errorf("couldn't find service offering for catalog name: %s", catalogName)
}

// MatchMultiple matches several ServiceOfferings based on some criteria
func (so ServiceOfferings) MatchMultiple(args resources.ResourceArguments) []string {
	return nil // implement me when needed
}

// GetType gets the type of the ServiceOfferings
func (so ServiceOfferings) GetType() string {
	return ServiceOfferingsType
}

// ServiceOfferingArguments holds all the necessary fields that are used when matching ServiceOfferings
type ServiceOfferingArguments struct {
	CatalogName string
}

// GetURLPath gets the URL Path of the ServiceOffering
func (sp ServiceOfferingArguments) GetURLPath() string {
	return paths.ServiceOfferingsPath
}
