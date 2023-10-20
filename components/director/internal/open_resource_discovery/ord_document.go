package ord

import (
	"encoding/json"
	"net/url"
	"path"
	"regexp"

	"dario.cat/mergo"

	"github.com/hashicorp/go-multierror"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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
	Tombstones              []*model.TombstoneInput             `json:"tombstones"`
	Vendors                 []*model.VendorInput                `json:"vendors"`
	Capabilities            []*model.CapabilityInput            `json:"capabilities"`
	IntegrationDependencies []*model.IntegrationDependencyInput `json:"integrationDependencies"`
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
	Packages                map[string]*model.Package
	Bundles                 map[string]*model.Bundle
	Capabilities            map[string]*model.Capability
	IntegrationDependencies map[string]*model.IntegrationDependency
}

// ResourceIDs holds some of the ORD entities' IDs
type ResourceIDs struct {
	PackageIDs               map[string]bool
	PackagePolicyLevels      map[string]string
	BundleIDs                map[string]bool
	ProductIDs               map[string]bool
	APIIDs                   map[string]bool
	EventIDs                 map[string]bool
	VendorIDs                map[string]bool
	CapabilityIDs            map[string]bool
	IntegrationDependencyIDs map[string]bool
}

// Validate validates all the documents for a system instance
func (docs Documents) Validate(calculatedBaseURL string, resourcesFromDB ResourcesFromDB, resourceHashes map[string]uint64, globalResourcesOrdIDs map[string]bool, credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping) error {
	var (
		errs                *multierror.Error
		baseURL             = calculatedBaseURL
		isBaseURLConfigured = len(calculatedBaseURL) > 0
	)

	for _, doc := range docs {
		if !isBaseURLConfigured && (doc.DescribedSystemInstance == nil || doc.DescribedSystemInstance.BaseURL == nil) {
			errs = multierror.Append(errs, errors.New("no baseURL was provided neither from /well-known URL, nor from config, nor from describedSystemInstance"))
			continue
		}

		if len(baseURL) == 0 {
			baseURL = *doc.DescribedSystemInstance.BaseURL
		}

		if doc.DescribedSystemInstance != nil {
			if err := ValidateSystemInstanceInput(doc.DescribedSystemInstance); err != nil {
				errs = multierror.Append(errs, errors.Wrap(err, "error validating system instance"))
			}
		}

		if doc.DescribedSystemVersion != nil {
			if err := ValidateSystemVersionInput(doc.DescribedSystemVersion); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating system version"))
			}
		}
		if doc.DescribedSystemInstance != nil && doc.DescribedSystemInstance.BaseURL != nil && *doc.DescribedSystemInstance.BaseURL != baseURL {
			errs = multierror.Append(errs, errors.Errorf("describedSystemInstance should be the same as the one providing the documents - %s : %s", *doc.DescribedSystemInstance.BaseURL, baseURL))
		}
	}

	resourceIDs := ResourceIDs{
		PackageIDs:          make(map[string]bool),
		PackagePolicyLevels: make(map[string]string),
		BundleIDs:           make(map[string]bool),
		ProductIDs:          make(map[string]bool),
		APIIDs:              make(map[string]bool),
		EventIDs:            make(map[string]bool),
		VendorIDs:           make(map[string]bool),
	}

	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if _, ok := resourceIDs.PackageIDs[pkg.OrdID]; ok {
				errs = multierror.Append(errs, errors.Errorf("found duplicate package with ord id %q", pkg.OrdID))
				continue
			}
			resourceIDs.PackageIDs[pkg.OrdID] = true
			if pkg.PolicyLevel != nil {
				resourceIDs.PackagePolicyLevels[pkg.OrdID] = *pkg.PolicyLevel
			}
		}
	}

	invalidApisIndices := make([]int, 0)
	invalidEventsIndices := make([]int, 0)
	invalidCapabilitiesIndices := make([]int, 0)
	invalidIntegrationDependenciesIndices := make([]int, 0)

	r1, e1 := docs.validateAndCheckForDuplications(SystemVersionPerspective, true, resourcesFromDB, resourceIDs, resourceHashes, credentialExchangeStrategyTenantMappings)
	r2, e2 := docs.validateAndCheckForDuplications(SystemInstancePerspective, true, resourcesFromDB, resourceIDs, resourceHashes, credentialExchangeStrategyTenantMappings)
	r3, e3 := docs.validateAndCheckForDuplications("", false, resourcesFromDB, resourceIDs, resourceHashes, credentialExchangeStrategyTenantMappings)
	errs = multierror.Append(errs, e1)
	errs = multierror.Append(errs, e2)
	errs = multierror.Append(errs, e3)

	if err := mergo.Merge(&resourceIDs, r1); err != nil {
		return err
	}
	if err := mergo.Merge(&resourceIDs, r2); err != nil {
		return err
	}
	if err := mergo.Merge(&resourceIDs, r3); err != nil {
		return err
	}

	// Validate entity relations
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if pkg.Vendor != nil && !resourceIDs.VendorIDs[*pkg.Vendor] && !globalResourcesOrdIDs[*pkg.Vendor] {
				errs = multierror.Append(errs, errors.Errorf("package with id %q has a reference to unknown vendor %q", pkg.OrdID, *pkg.Vendor))
			}
			ordIDs := gjson.ParseBytes(pkg.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					errs = multierror.Append(errs, errors.Errorf("package with id %q has a reference to unknown product %q", pkg.OrdID, productID.String()))
				}
			}
		}
		for _, product := range doc.Products {
			if !resourceIDs.VendorIDs[product.Vendor] && !globalResourcesOrdIDs[product.Vendor] {
				errs = multierror.Append(errs, errors.Errorf("product with id %q has a reference to unknown vendor %q", product.OrdID, product.Vendor))
			}
		}
		for i, api := range doc.APIResources {
			if api.OrdPackageID != nil && !resourceIDs.PackageIDs[*api.OrdPackageID] {
				errs = multierror.Append(errs, errors.Errorf("api with id %q has a reference to unknown package %q", *api.OrdID, *api.OrdPackageID))
				invalidApisIndices = append(invalidApisIndices, i)
			}
			if api.PartOfConsumptionBundles != nil {
				for _, apiBndlRef := range api.PartOfConsumptionBundles {
					if !resourceIDs.BundleIDs[apiBndlRef.BundleOrdID] {
						errs = multierror.Append(errs, errors.Errorf("api with id %q has a reference to unknown bundle %q", *api.OrdID, apiBndlRef.BundleOrdID))
					}
				}
			}

			ordIDs := gjson.ParseBytes(api.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					errs = multierror.Append(errs, errors.Errorf("api with id %q has a reference to unknown product %q", *api.OrdID, productID.String()))
				}
			}
		}

		for i, event := range doc.EventResources {
			if event.OrdPackageID != nil && !resourceIDs.PackageIDs[*event.OrdPackageID] {
				errs = multierror.Append(errs, errors.Errorf("event with id %q has a reference to unknown package %q", *event.OrdID, *event.OrdPackageID))
				invalidEventsIndices = append(invalidEventsIndices, i)
			}
			if event.PartOfConsumptionBundles != nil {
				for _, eventBndlRef := range event.PartOfConsumptionBundles {
					if !resourceIDs.BundleIDs[eventBndlRef.BundleOrdID] {
						errs = multierror.Append(errs, errors.Errorf("event with id %q has a reference to unknown bundle %q", *event.OrdID, eventBndlRef.BundleOrdID))
					}
				}
			}

			ordIDs := gjson.ParseBytes(event.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					errs = multierror.Append(errs, errors.Errorf("event with id %q has a reference to unknown product %q", *event.OrdID, productID.String()))
				}
			}
		}

		for i, capability := range doc.Capabilities {
			if capability.OrdPackageID != nil && !resourceIDs.PackageIDs[*capability.OrdPackageID] {
				errs = multierror.Append(errs, errors.Errorf("capability with id %q has a reference to unknown package %q", *capability.OrdID, *capability.OrdPackageID))
				invalidCapabilitiesIndices = append(invalidCapabilitiesIndices, i)
			}
		}

		for i, integrationDependency := range doc.IntegrationDependencies {
			if integrationDependency.OrdPackageID != nil && !resourceIDs.PackageIDs[*integrationDependency.OrdPackageID] {
				errs = multierror.Append(errs, errors.Errorf("integration dependency with id %q has a reference to unknown package %q", *integrationDependency.OrdID, *integrationDependency.OrdPackageID))
				invalidIntegrationDependenciesIndices = append(invalidIntegrationDependenciesIndices, i)
			}
		}

		doc.APIResources = deleteInvalidInputObjects(invalidApisIndices, doc.APIResources)
		doc.EventResources = deleteInvalidInputObjects(invalidEventsIndices, doc.EventResources)
		doc.Capabilities = deleteInvalidInputObjects(invalidCapabilitiesIndices, doc.Capabilities)
		doc.IntegrationDependencies = deleteInvalidInputObjects(invalidIntegrationDependenciesIndices, doc.IntegrationDependencies)
		invalidApisIndices = nil
		invalidEventsIndices = nil
		invalidCapabilitiesIndices = nil
		invalidIntegrationDependenciesIndices = nil
	}

	return errs.ErrorOrNil()
}

func (docs Documents) validateAndCheckForDuplications(perspectiveConstraint DocumentPerspective, forbidDuplications bool, resourcesFromDB ResourcesFromDB, resourceID ResourceIDs, resourceHashes map[string]uint64, credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping) (ResourceIDs, *multierror.Error) {
	errs := &multierror.Error{}

	resourceIDs := ResourceIDs{
		PackageIDs:               make(map[string]bool),
		PackagePolicyLevels:      resourceID.PackagePolicyLevels,
		BundleIDs:                make(map[string]bool),
		ProductIDs:               make(map[string]bool),
		APIIDs:                   make(map[string]bool),
		EventIDs:                 make(map[string]bool),
		VendorIDs:                make(map[string]bool),
		CapabilityIDs:            make(map[string]bool),
		IntegrationDependencyIDs: make(map[string]bool),
	}
	for _, doc := range docs {
		if doc.Perspective == perspectiveConstraint {
			continue
		}
		invalidPackagesIndices := make([]int, 0)
		invalidBundlesIndices := make([]int, 0)
		invalidProductsIndices := make([]int, 0)
		invalidVendorsIndices := make([]int, 0)
		invalidTombstonesIndices := make([]int, 0)
		invalidApisIndices := make([]int, 0)
		invalidEventsIndices := make([]int, 0)
		invalidCapabilitiesIndices := make([]int, 0)
		invalidIntegrationDependenciesIndices := make([]int, 0)

		if err := validateDocumentInput(doc); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "error validating document"))
		}

		for i, pkg := range doc.Packages {
			if err := validatePackageInputWithSuppressedErrors(pkg, resourcesFromDB.Packages, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating package with ord id %q", pkg.OrdID))
			}
			if err := validatePackageInput(pkg, doc.PolicyLevel); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating package with ord id %q", pkg.OrdID))
				invalidPackagesIndices = append(invalidPackagesIndices, i)
				resourceIDs.PackageIDs[pkg.OrdID] = false
			}
		}

		for i, bndl := range doc.ConsumptionBundles {
			if err := validateBundleInputWithSuppressedErrors(bndl, resourcesFromDB.Bundles, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating bundle with ord id %q", stringPtrToString(bndl.OrdID)))
			}
			if err := validateBundleInput(bndl, credentialExchangeStrategyTenantMappings); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating bundle with ord id %q", stringPtrToString(bndl.OrdID)))
				invalidBundlesIndices = append(invalidBundlesIndices, i)
				continue
			}
			if bndl.OrdID != nil {
				if _, ok := resourceIDs.BundleIDs[*bndl.OrdID]; ok && forbidDuplications {
					errs = multierror.Append(errs, errors.Errorf("found duplicate bundle with ord id %q", *bndl.OrdID))
				}
				resourceIDs.BundleIDs[*bndl.OrdID] = true
			}
		}

		for i, product := range doc.Products {
			if err := validateProductInput(product); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating product with ord id %q", product.OrdID))
				invalidProductsIndices = append(invalidProductsIndices, i)
				continue
			}
			if _, ok := resourceIDs.ProductIDs[product.OrdID]; ok && forbidDuplications {
				errs = multierror.Append(errs, errors.Errorf("found duplicate product with ord id %q", product.OrdID))
			}
			resourceIDs.ProductIDs[product.OrdID] = true
		}

		for i, api := range doc.APIResources {
			if err := validateAPIInputWithSuppressedErrors(api, resourcesFromDB.APIs, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating api with ord id %q", stringPtrToString(api.OrdID)))
			}
			if err := validateAPIInput(api, doc.PolicyLevel); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating api with ord id %q", stringPtrToString(api.OrdID)))
				invalidApisIndices = append(invalidApisIndices, i)
				continue
			}
			if api.OrdID != nil {
				if _, ok := resourceIDs.APIIDs[*api.OrdID]; ok && forbidDuplications {
					errs = multierror.Append(errs, errors.Errorf("found duplicate api with ord id %q", *api.OrdID))
				}
				resourceIDs.APIIDs[*api.OrdID] = true
			}
		}

		for i, event := range doc.EventResources {
			if err := validateEventInputWithSuppressedErrors(event, resourcesFromDB.Events, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating event with ord id %q", stringPtrToString(event.OrdID)))
			}
			if err := validateEventInput(event, doc.PolicyLevel); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating event with ord id %q", stringPtrToString(event.OrdID)))
				invalidEventsIndices = append(invalidEventsIndices, i)
				continue
			}

			if event.OrdID != nil {
				if _, ok := resourceIDs.EventIDs[*event.OrdID]; ok && forbidDuplications {
					errs = multierror.Append(errs, errors.Errorf("found duplicate event with ord id %q", *event.OrdID))
				}

				resourceIDs.EventIDs[*event.OrdID] = true
			}
		}

		for i, capability := range doc.Capabilities {
			if err := validateCapabilityInputWithSuppressedErrors(capability, resourcesFromDB.Capabilities, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating capability with ord id %q", stringPtrToString(capability.OrdID)))
			}
			if err := validateCapabilityInput(capability); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating capability with ord id %q", stringPtrToString(capability.OrdID)))
				invalidCapabilitiesIndices = append(invalidCapabilitiesIndices, i)
				continue
			}
			if capability.OrdID != nil {
				if _, ok := resourceIDs.CapabilityIDs[*capability.OrdID]; ok && forbidDuplications {
					errs = multierror.Append(errs, errors.Errorf("found duplicate capability with ord id %q", *capability.OrdID))
				}
				resourceIDs.CapabilityIDs[*capability.OrdID] = true
			}
		}

		for i, integrationDependency := range doc.IntegrationDependencies {
			if err := validateIntegrationDependencyInputWithSuppressedErrors(integrationDependency, resourcesFromDB.IntegrationDependencies, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating integration dependency with ord id %q", stringPtrToString(integrationDependency.OrdID)))
			}
			if err := validateIntegrationDependencyInput(integrationDependency, doc.PolicyLevel); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating integration dependency with ord id %q", stringPtrToString(integrationDependency.OrdID)))
				invalidIntegrationDependenciesIndices = append(invalidIntegrationDependenciesIndices, i)
				continue
			}
			if integrationDependency.OrdID != nil {
				if _, ok := resourceIDs.IntegrationDependencyIDs[*integrationDependency.OrdID]; ok && forbidDuplications {
					errs = multierror.Append(errs, errors.Errorf("found duplicate integration dependency with ord id %q", *integrationDependency.OrdID))
				}
				resourceIDs.IntegrationDependencyIDs[*integrationDependency.OrdID] = true
			}
		}

		for i, vendor := range doc.Vendors {
			if err := validateVendorInput(vendor); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating vendor with ord id %q", vendor.OrdID))
				invalidVendorsIndices = append(invalidVendorsIndices, i)
				continue
			}
			if _, ok := resourceIDs.VendorIDs[vendor.OrdID]; ok && forbidDuplications {
				errs = multierror.Append(errs, errors.Errorf("found duplicate vendor with ord id %q", vendor.OrdID))
			}
			resourceIDs.VendorIDs[vendor.OrdID] = true
		}

		for i, tombstone := range doc.Tombstones {
			if err := validateTombstoneInput(tombstone); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating tombstone with ord id %q", tombstone.OrdID))
				invalidTombstonesIndices = append(invalidTombstonesIndices, i)
			}
		}

		doc.Packages = deleteInvalidInputObjects(invalidPackagesIndices, doc.Packages)
		doc.ConsumptionBundles = deleteInvalidInputObjects(invalidBundlesIndices, doc.ConsumptionBundles)
		doc.Products = deleteInvalidInputObjects(invalidProductsIndices, doc.Products)
		doc.APIResources = deleteInvalidInputObjects(invalidApisIndices, doc.APIResources)
		doc.EventResources = deleteInvalidInputObjects(invalidEventsIndices, doc.EventResources)
		doc.Vendors = deleteInvalidInputObjects(invalidVendorsIndices, doc.Vendors)
		doc.Tombstones = deleteInvalidInputObjects(invalidTombstonesIndices, doc.Tombstones)
		doc.Capabilities = deleteInvalidInputObjects(invalidCapabilitiesIndices, doc.Capabilities)
	}

	return ResourceIDs{
		PackageIDs:          resourceIDs.PackageIDs,
		ProductIDs:          resourceIDs.ProductIDs,
		APIIDs:              resourceIDs.APIIDs,
		EventIDs:            resourceIDs.EventIDs,
		VendorIDs:           resourceIDs.VendorIDs,
		BundleIDs:           resourceIDs.BundleIDs,
		PackagePolicyLevels: resourceIDs.PackagePolicyLevels,
		CapabilityIDs:       resourceIDs.CapabilityIDs,
	}, errs
}

// Sanitize performs all the merging and rewriting rules defined in ORD. This method should be invoked after Documents are validated with the Validate method.
//   - Rewrite all relative URIs using the baseURL from the Described System Instance. If the Described System Instance baseURL is missing the provider baseURL (from the webhook) is used.
//   - Package's partOfProducts, tags, countries, industry, lineOfBusiness, labels are inherited by the resources in the package.
//   - Ensure to assign `defaultEntryPoint` if missing and there are available `entryPoints` to API's `PartOfConsumptionBundles`
//   - If some resource(Package, API or Event) doesn't have provided `policyLevel` and `customPolicyLevel`, these are inherited from the document
func (docs Documents) Sanitize(webhookBaseURL, webhookBaseProxyURL string) error {
	var err error

	// Use the ProxyURL for all relative link substitution except for the API's TargetURLs.
	// They are externally consumable and we should not expose those URLs through the Proxy but rather from webhook's BaseURL
	url := webhookBaseURL
	if webhookBaseProxyURL != "" {
		url = webhookBaseProxyURL
	}

	// Rewrite relative URIs
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if pkg.PackageLinks, err = rewriteRelativeURIsInJSON(pkg.PackageLinks, url, "url"); err != nil {
				return err
			}
			if pkg.Links, err = rewriteRelativeURIsInJSON(pkg.Links, url, "url"); err != nil {
				return err
			}
		}

		for _, bndl := range doc.ConsumptionBundles {
			if bndl.Links, err = rewriteRelativeURIsInJSON(bndl.Links, url, "url"); err != nil {
				return err
			}
			if bndl.CredentialExchangeStrategies, err = rewriteRelativeURIsInJSON(bndl.CredentialExchangeStrategies, url, "callbackUrl"); err != nil {
				return err
			}
		}

		for _, api := range doc.APIResources {
			for _, definition := range api.ResourceDefinitions {
				if !isAbsoluteURL(definition.URL) {
					definition.URL = url + definition.URL
				}
			}
			if api.APIResourceLinks, err = rewriteRelativeURIsInJSON(api.APIResourceLinks, url, "url"); err != nil {
				return err
			}
			if api.Links, err = rewriteRelativeURIsInJSON(api.Links, url, "url"); err != nil {
				return err
			}
			if api.ChangeLogEntries, err = rewriteRelativeURIsInJSON(api.ChangeLogEntries, url, "url"); err != nil {
				return err
			}
			if api.TargetURLs, err = rewriteRelativeURIsInJSONArray(api.TargetURLs, webhookBaseURL); err != nil {
				return err
			}
			rewriteDefaultTargetURL(api.PartOfConsumptionBundles, url)
		}

		for _, event := range doc.EventResources {
			if event.ChangeLogEntries, err = rewriteRelativeURIsInJSON(event.ChangeLogEntries, url, "url"); err != nil {
				return err
			}
			if event.Links, err = rewriteRelativeURIsInJSON(event.Links, url, "url"); err != nil {
				return err
			}
			for _, definition := range event.ResourceDefinitions {
				if !isAbsoluteURL(definition.URL) {
					definition.URL = url + definition.URL
				}
			}
		}

		for _, capability := range doc.Capabilities {
			for _, definition := range capability.CapabilityDefinitions {
				if !isAbsoluteURL(definition.URL) {
					definition.URL = url + definition.URL
				}
			}

			if capability.Links, err = rewriteRelativeURIsInJSON(capability.Links, url, "url"); err != nil {
				return err
			}
		}

		for _, integrationDependency := range doc.IntegrationDependencies {
			if integrationDependency.Links, err = rewriteRelativeURIsInJSON(integrationDependency.Links, url, "url"); err != nil {
				return err
			}
		}
	}

	// Package properties inheritance
	packages := make(map[string]*model.PackageInput)
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			packages[pkg.OrdID] = pkg
			if pkg.PolicyLevel == nil {
				pkg.PolicyLevel = doc.PolicyLevel
				pkg.CustomPolicyLevel = doc.CustomPolicyLevel
			}
		}
	}

	for _, doc := range docs {
		for _, api := range doc.APIResources {
			if api.PolicyLevel == nil {
				api.PolicyLevel = doc.PolicyLevel
				api.CustomPolicyLevel = doc.CustomPolicyLevel
			}

			referredPkg, ok := packages[*api.OrdPackageID]
			if !ok {
				return errors.Errorf("api with ord id %q has a reference to unknown package %q", *api.OrdID, *api.OrdPackageID)
			}
			if api.PartOfProducts, err = mergeJSONArraysOfStrings(referredPkg.PartOfProducts, api.PartOfProducts); err != nil {
				return errors.Wrapf(err, "error while merging partOfProducts for api with ord id %q", *api.OrdID)
			}
			if api.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, api.Tags); err != nil {
				return errors.Wrapf(err, "error while merging tags for api with ord id %q", *api.OrdID)
			}
			if api.Countries, err = mergeJSONArraysOfStrings(referredPkg.Countries, api.Countries); err != nil {
				return errors.Wrapf(err, "error while merging countries for api with ord id %q", *api.OrdID)
			}
			if api.Industry, err = mergeJSONArraysOfStrings(referredPkg.Industry, api.Industry); err != nil {
				return errors.Wrapf(err, "error while merging industry for api with ord id %q", *api.OrdID)
			}
			if api.LineOfBusiness, err = mergeJSONArraysOfStrings(referredPkg.LineOfBusiness, api.LineOfBusiness); err != nil {
				return errors.Wrapf(err, "error while merging lineOfBusiness for api with ord id %q", *api.OrdID)
			}
			if api.Labels, err = mergeORDLabels(referredPkg.Labels, api.Labels); err != nil {
				return errors.Wrapf(err, "error while merging labels for api with ord id %q", *api.OrdID)
			}
			assignDefaultEntryPointIfNeeded(api.PartOfConsumptionBundles, api.TargetURLs)
		}
		for _, event := range doc.EventResources {
			if event.PolicyLevel == nil {
				event.PolicyLevel = doc.PolicyLevel
				event.CustomPolicyLevel = doc.CustomPolicyLevel
			}

			referredPkg, ok := packages[*event.OrdPackageID]
			if !ok {
				return errors.Errorf("event with ord id %q has a reference to unknown package %q", *event.OrdID, *event.OrdPackageID)
			}
			if event.PartOfProducts, err = mergeJSONArraysOfStrings(referredPkg.PartOfProducts, event.PartOfProducts); err != nil {
				return errors.Wrapf(err, "error while merging partOfProducts for event with ord id %q", *event.OrdID)
			}
			if event.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, event.Tags); err != nil {
				return errors.Wrapf(err, "error while merging tags for event with ord id %q", *event.OrdID)
			}
			if event.Countries, err = mergeJSONArraysOfStrings(referredPkg.Countries, event.Countries); err != nil {
				return errors.Wrapf(err, "error while merging countries for event with ord id %q", *event.OrdID)
			}
			if event.Industry, err = mergeJSONArraysOfStrings(referredPkg.Industry, event.Industry); err != nil {
				return errors.Wrapf(err, "error while merging industry for event with ord id %q", *event.OrdID)
			}
			if event.LineOfBusiness, err = mergeJSONArraysOfStrings(referredPkg.LineOfBusiness, event.LineOfBusiness); err != nil {
				return errors.Wrapf(err, "error while merging lineOfBusiness for event with ord id %q", *event.OrdID)
			}
			if event.Labels, err = mergeORDLabels(referredPkg.Labels, event.Labels); err != nil {
				return errors.Wrapf(err, "error while merging labels for event with ord id %q", *event.OrdID)
			}
		}
		for _, capability := range doc.Capabilities {
			referredPkg, ok := packages[*capability.OrdPackageID]
			if !ok {
				return errors.Errorf("capability with ord id %q has a reference to unknown package %q", *capability.OrdID, *capability.OrdPackageID)
			}
			if capability.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, capability.Tags); err != nil {
				return errors.Wrapf(err, "error while merging tags for capability with ord id %q", *capability.OrdID)
			}
			if capability.Labels, err = mergeORDLabels(referredPkg.Labels, capability.Labels); err != nil {
				return errors.Wrapf(err, "error while merging labels for capability with ord id %q", *capability.OrdID)
			}
		}
		for _, integrationDependency := range doc.IntegrationDependencies {
			referredPkg, ok := packages[*integrationDependency.OrdPackageID]
			if !ok {
				return errors.Errorf("integration dependency with ord id %q has a reference to unknown package %q", *integrationDependency.OrdID, *integrationDependency.OrdPackageID)
			}
			if integrationDependency.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, integrationDependency.Tags); err != nil {
				return errors.Wrapf(err, "error while merging tags for integration dependency with ord id %q", *integrationDependency.OrdID)
			}
			if integrationDependency.Labels, err = mergeORDLabels(referredPkg.Labels, integrationDependency.Labels); err != nil {
				return errors.Wrapf(err, "error while merging labels for integration dependency with ord id %q", *integrationDependency.OrdID)
			}
		}
	}

	return err
}

// mergeORDLabels merges labels2 into labels1
func mergeORDLabels(labels1, labels2 json.RawMessage) (json.RawMessage, error) {
	if len(labels2) == 0 {
		return labels1, nil
	}
	parsedLabels1 := gjson.ParseBytes(labels1)
	parsedLabels2 := gjson.ParseBytes(labels2)
	if !parsedLabels1.IsObject() || !parsedLabels2.IsObject() {
		return nil, errors.New("invalid arguments: expected two json objects")
	}

	labels1Map := parsedLabels1.Map()
	labels2Map := parsedLabels2.Map()

	for k, v := range labels1Map {
		if v2, ok := labels2Map[k]; ok {
			mergedValues, err := mergeJSONArraysOfStrings(json.RawMessage(v.Raw), json.RawMessage(v2.Raw))
			if err != nil {
				return nil, errors.Wrapf(err, "while merging values for key %q", k)
			}
			labels1Map[k] = gjson.ParseBytes(mergedValues)
			delete(labels2Map, k)
		}
	}

	for k, v := range labels2Map {
		labels1Map[k] = v
	}

	result := make(map[string]interface{}, len(labels1Map))
	for k, v := range labels1Map {
		result[k] = v.Value()
	}

	return json.Marshal(result)
}

// mergeJSONArraysOfStrings merges arr2 in arr1
func mergeJSONArraysOfStrings(arr1, arr2 json.RawMessage) (json.RawMessage, error) {
	if len(arr2) == 0 {
		return arr1, nil
	}
	parsedArr1 := gjson.ParseBytes(arr1)
	parsedArr2 := gjson.ParseBytes(arr2)
	if !parsedArr1.IsArray() || !parsedArr2.IsArray() {
		return nil, errors.New("invalid arguments: expected two json arrays")
	}
	resultJSONArr := append(parsedArr1.Array(), parsedArr2.Array()...)
	result := make([]string, 0, len(resultJSONArr))
	for _, el := range resultJSONArr {
		if el.Type != gjson.String {
			return nil, errors.New("invalid arguments: expected json array of strings")
		}
		result = append(result, el.String())
	}
	result = deduplicate(result)
	return json.Marshal(result)
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

func deduplicate(s []string) []string {
	if len(s) <= 1 {
		return s
	}

	result := make([]string, 0, len(s))
	seen := make(map[string]bool)
	for _, val := range s {
		if !seen[val] {
			result = append(result, val)
			seen[val] = true
		}
	}
	return result
}

func rewriteRelativeURIsInJSONArray(j json.RawMessage, baseURL string) (json.RawMessage, error) {
	parsedJSON := gjson.ParseBytes(j)

	items := make([]interface{}, 0)
	for _, crrURI := range parsedJSON.Array() {
		if !isAbsoluteURL(crrURI.String()) {
			rewrittenURI := baseURL + crrURI.String()

			items = append(items, rewrittenURI)
		} else {
			items = append(items, crrURI.String())
		}
	}

	rewrittenJSON, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	return rewrittenJSON, nil
}

func rewriteDefaultTargetURL(bundleRefs []*model.ConsumptionBundleReference, baseURL string) {
	for _, br := range bundleRefs {
		if br.DefaultTargetURL != "" && !isAbsoluteURL(br.DefaultTargetURL) {
			br.DefaultTargetURL = baseURL + br.DefaultTargetURL
		}
	}
}

func rewriteRelativeURIsInJSON(j json.RawMessage, baseURL, jsonPath string) (json.RawMessage, error) {
	parsedJSON := gjson.ParseBytes(j)
	if parsedJSON.IsArray() {
		items := make([]interface{}, 0)
		for _, jsonElement := range parsedJSON.Array() {
			rewrittenElement, err := rewriteRelativeURIsInJSON(json.RawMessage(jsonElement.Raw), baseURL, jsonPath)
			if err != nil {
				return nil, err
			}
			items = append(items, gjson.ParseBytes(rewrittenElement).Value())
		}
		return json.Marshal(items)
	} else if parsedJSON.IsObject() {
		uriProperty := gjson.GetBytes(j, jsonPath)
		if uriProperty.Exists() && !isAbsoluteURL(uriProperty.String()) {
			u, err := url.Parse(baseURL)
			if err != nil {
				return nil, err
			}
			u.Path = path.Join(u.Path, uriProperty.String())
			return sjson.SetBytes(j, jsonPath, u.String())
		}
	}
	return j, nil
}

func assignDefaultEntryPointIfNeeded(bundleReferences []*model.ConsumptionBundleReference, targetURLs json.RawMessage) {
	lenTargetURLs := len(gjson.ParseBytes(targetURLs).Array())
	for _, br := range bundleReferences {
		if br.DefaultTargetURL == "" && lenTargetURLs > 1 {
			br.DefaultTargetURL = gjson.ParseBytes(targetURLs).Array()[0].String()
		}
	}
}

func isAbsoluteURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func stringPtrToString(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}

func deleteInvalidInputObjects[T any](invalidObjectsIndices []int, objects []T) []T {
	decreaseIndexForDeleting := 0
	for _, invalidObjectIndex := range invalidObjectsIndices {
		deleteIndex := invalidObjectIndex - decreaseIndexForDeleting
		objects = append(objects[:deleteIndex], objects[deleteIndex+1:]...)
		decreaseIndexForDeleting++
	}

	return objects
}
