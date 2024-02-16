package ord

import (
	"net/url"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// WellKnownEndpoint is the single entry point for the discovery.
const WellKnownEndpoint = "/.well-known/open-resource-discovery"

// DocumentPerspective represents the perspective of the document
type DocumentPerspective string

const (
	// SystemVersionPerspective represents a dynamic document
	SystemVersionPerspective DocumentPerspective = "system-version"
	// SystemInstancePerspective represents a static document
	SystemInstancePerspective DocumentPerspective = "system-instance"
	// ConfigBaseURLRegex represents the valid structure of the field
	ConfigBaseURLRegex = "^http[s]?:\\/\\/[^:\\/\\s]+\\.[^:\\/\\s\\.]+(:\\d+)?(\\/[a-zA-Z0-9-\\._~]+)*$"
)

// WellKnownConfig represents the whole config object
type WellKnownConfig struct {
	Schema                  string                  `json:"$schema"`
	BaseURL                 string                  `json:"baseUrl"`
	OpenResourceDiscoveryV1 OpenResourceDiscoveryV1 `json:"openResourceDiscoveryV1"`
}

// OpenResourceDiscoveryV1 contains all Documents' details
type OpenResourceDiscoveryV1 struct {
	Documents []DocumentDetails `json:"documents"`
}

// DocumentDetails contains fields related to the fetching of each Document
type DocumentDetails struct {
	URL string `json:"url"`
	// TODO: Currently we cannot differentiate between system instance types reliably, therefore we cannot make use of the systemInstanceAware optimization (store it once per system type and reuse it for each system instance of that type).
	//  Once we have system landscape discovery and stable system types we can make use of this optimization. Until then we store all the information for a system instance as it is provided in the documents.
	//  Therefore we treat every resource as SystemInstanceAware = true
	SystemInstanceAware bool                            `json:"systemInstanceAware"`
	AccessStrategies    accessstrategy.AccessStrategies `json:"accessStrategies"`
	Perspective         DocumentPerspective             `json:"perspective"`
}

// Document represents an ORD Document
type Document struct {
	Schema                string `json:"$schema"`
	OpenResourceDiscovery string `json:"openResourceDiscovery"`
	Description           string `json:"description"`

	// TODO: In the current state of ORD and it's implementation we are missing system landscape discovery and an id correlation in the system instances. Because of that in the first phase we will rely on:
	//  - DescribedSystemInstance is the application in our DB and it's baseURL should match with the one in the webhook.
	DescribedSystemInstance *model.Application `json:"describedSystemInstance"`

	DescribedSystemVersion *model.ApplicationTemplateVersionInput `json:"describedSystemVersion"`

	Perspective DocumentPerspective `json:"-"`

	PolicyLevel       *string `json:"policyLevel"`
	CustomPolicyLevel *string `json:"customPolicyLevel"`

	Packages                []*model.PackageInput               `json:"packages"`
	ConsumptionBundles      []*model.BundleCreateInput          `json:"consumptionBundles"`
	Products                []*model.ProductInput               `json:"products"`
	APIResources            []*model.APIDefinitionInput         `json:"apiResources"`
	EventResources          []*model.EventDefinitionInput       `json:"eventResources"`
	EntityTypes             []*model.EntityTypeInput            `json:"entityTypes"`
	Tombstones              []*model.TombstoneInput             `json:"tombstones"`
	Vendors                 []*model.VendorInput                `json:"vendors"`
	Capabilities            []*model.CapabilityInput            `json:"capabilities"`
	IntegrationDependencies []*model.IntegrationDependencyInput `json:"integrationDependencies"`
	DataProducts            []*model.DataProductInput           `json:"dataProducts"`
}

// Validate validates if the Config object complies with the spec requirements
func (c WellKnownConfig) Validate(baseURL string) error {
	if err := validation.Validate(c.BaseURL, validation.Match(regexp.MustCompile(ConfigBaseURLRegex))); err != nil {
		return err
	}

	if err := validation.Validate(c.OpenResourceDiscoveryV1.Documents, validation.Required); err != nil {
		return err
	}

	for _, docDetails := range c.OpenResourceDiscoveryV1.Documents {
		if err := validateDocDetails(docDetails); err != nil {
			return err
		}
	}

	areDocsWithRelativeURLs, err := checkForRelativeDocURLs(c.OpenResourceDiscoveryV1.Documents)
	if err != nil {
		return err
	}

	if baseURL == "" && areDocsWithRelativeURLs {
		return errors.New("there are relative document URls but no baseURL provided neither in config nor through /well-known endpoint")
	}

	return nil
}

// Documents is a slice of Document objects
type Documents []*Document

// ResourcesFromDB holds some of the ORD data from the database
type ResourcesFromDB struct {
	APIs                    map[string]*model.APIDefinition
	Events                  map[string]*model.EventDefinition
	EntityTypes             map[string]*model.EntityType
	Packages                map[string]*model.Package
	Bundles                 map[string]*model.Bundle
	Capabilities            map[string]*model.Capability
	IntegrationDependencies map[string]*model.IntegrationDependency
	DataProducts            map[string]*model.DataProduct
}

// ResourceIDs holds some of the ORD entities' IDs
type ResourceIDs struct {
	PackageIDs               map[string]bool
	PackagePolicyLevels      map[string]string
	BundleIDs                map[string]bool
	ProductIDs               map[string]bool
	APIIDs                   map[string]bool
	EventIDs                 map[string]bool
	EntityTypeIDs            map[string]bool
	VendorIDs                map[string]bool
	CapabilityIDs            map[string]bool
	IntegrationDependencyIDs map[string]bool
	DataProductIDs           map[string]bool
}

func validateDocDetails(docDetails DocumentDetails) error {
	if err := validation.Validate(docDetails.URL, validation.Required); err != nil {
		return err
	}

	if err := validation.Validate(docDetails.AccessStrategies, validation.Required); err != nil {
		return err
	}

	for _, as := range docDetails.AccessStrategies {
		if err := as.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func checkForRelativeDocURLs(docs []DocumentDetails) (bool, error) {
	for _, doc := range docs {
		parsedDocURL, err := url.ParseRequestURI(doc.URL)
		if err != nil {
			return false, errors.New("error while parsing document url")
		}

		if parsedDocURL.Host == "" {
			return true, nil
		}
	}
	return false, nil
}
