package validator

import (
	"github.com/hashicorp/go-multierror"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type DocumentValidator struct {
	client *ValidationClient
}

func NewDocumentValidator(client *ValidationClient) *DocumentValidator {
	return &DocumentValidator{
		client: client,
	}
}

func (v *DocumentValidator) Validate(document string) ([]ValidationError, error) {
	result, err := v.client.Validate("sap:core:v1", document)
	if err != nil {
		return nil, errors.Wrap(err, "while validating document")
	}

	return v.toValidationErrors(result), nil
}

func (v *DocumentValidator) toValidationErrors(result []ValidationResult) []ValidationError {
	valErrs := make([]ValidationError, 0)
	for _, r := range result {
		valErrs = append(valErrs, ValidationError{
			OrdId:       "", // TODO
			Severity:    r.Severity,
			Type:        r.Code,
			Description: r.Message,
		})
	}
	return valErrs
}

// ValidateOld validates all the documents for a system instance
func (v *DocumentValidator) ValidateOld(docs []*ord.Document, calculatedBaseURL string, resourcesFromDB ord.ResourcesFromDB, resourceHashes map[string]uint64, globalResourcesOrdIDs map[string]bool, credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping) error {
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
			if err := ord.ValidateSystemInstanceInput(doc.DescribedSystemInstance); err != nil {
				errs = multierror.Append(errs, errors.Wrap(err, "error validating system instance"))
			}
		}

		if doc.DescribedSystemVersion != nil {
			if err := ord.ValidateSystemVersionInput(doc.DescribedSystemVersion); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating system version"))
			}
		}
		if doc.DescribedSystemInstance != nil && doc.DescribedSystemInstance.BaseURL != nil && *doc.DescribedSystemInstance.BaseURL != baseURL {
			errs = multierror.Append(errs, errors.Errorf("describedSystemInstance should be the same as the one providing the documents - %s : %s", *doc.DescribedSystemInstance.BaseURL, baseURL))
		}
	}

	resourceIDs := ord.ResourceIDs{
		PackageIDs:               make(map[string]bool),
		PackagePolicyLevels:      make(map[string]string),
		BundleIDs:                make(map[string]bool),
		ProductIDs:               make(map[string]bool),
		APIIDs:                   make(map[string]bool),
		EventIDs:                 make(map[string]bool),
		EntityTypeIDs:            make(map[string]bool),
		VendorIDs:                make(map[string]bool),
		CapabilityIDs:            make(map[string]bool),
		IntegrationDependencyIDs: make(map[string]bool),
		DataProductIDs:           make(map[string]bool),
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
	invalidEntityTypesIndices := make([]int, 0)
	invalidCapabilitiesIndices := make([]int, 0)
	invalidIntegrationDependenciesIndices := make([]int, 0)
	invalidDataProductsIndices := make([]int, 0)

	r1, e1 := v.validateAndCheckForDuplications(docs, ord.SystemVersionPerspective, true, resourcesFromDB, resourceIDs, resourceHashes, credentialExchangeStrategyTenantMappings)
	r2, e2 := v.validateAndCheckForDuplications(docs, ord.SystemInstancePerspective, true, resourcesFromDB, resourceIDs, resourceHashes, credentialExchangeStrategyTenantMappings)
	r3, e3 := v.validateAndCheckForDuplications(docs, "", false, resourcesFromDB, resourceIDs, resourceHashes, credentialExchangeStrategyTenantMappings)
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

		for i, entityType := range doc.EntityTypes {
			if !resourceIDs.PackageIDs[entityType.OrdPackageID] {
				errs = multierror.Append(errs, errors.Errorf("entity type with id %q has a reference to unknown package %q", entityType.OrdID, entityType.OrdPackageID))
				invalidEntityTypesIndices = append(invalidEntityTypesIndices, i)
			}

			ordIDs := gjson.ParseBytes(entityType.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					errs = multierror.Append(errs, errors.Errorf("entity type with id %q has a reference to unknown product %q", entityType.OrdID, productID.String()))
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

		for i, dataProduct := range doc.DataProducts {
			if dataProduct.OrdPackageID != nil && !resourceIDs.PackageIDs[*dataProduct.OrdPackageID] {
				errs = multierror.Append(errs, errors.Errorf("data product with id %q has a reference to unknown package %q", *dataProduct.OrdID, *dataProduct.OrdPackageID))
				invalidDataProductsIndices = append(invalidDataProductsIndices, i)
			}
		}

		doc.APIResources = deleteInvalidInputObjects(invalidApisIndices, doc.APIResources)
		doc.EventResources = deleteInvalidInputObjects(invalidEventsIndices, doc.EventResources)
		doc.EntityTypes = deleteInvalidInputObjects(invalidEntityTypesIndices, doc.EntityTypes)
		doc.Capabilities = deleteInvalidInputObjects(invalidCapabilitiesIndices, doc.Capabilities)
		doc.IntegrationDependencies = deleteInvalidInputObjects(invalidIntegrationDependenciesIndices, doc.IntegrationDependencies)
		doc.DataProducts = deleteInvalidInputObjects(invalidDataProductsIndices, doc.DataProducts)
		invalidApisIndices = nil
		invalidEventsIndices = nil
		invalidCapabilitiesIndices = nil
		invalidIntegrationDependenciesIndices = nil
		invalidDataProductsIndices = nil
	}

	return errs.ErrorOrNil()
}

func (v *DocumentValidator) validateAndCheckForDuplications(docs []*ord.Document, perspectiveConstraint ord.DocumentPerspective, forbidDuplications bool, resourcesFromDB ord.ResourcesFromDB, resourceID ord.ResourceIDs, resourceHashes map[string]uint64, credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping) (ord.ResourceIDs, *multierror.Error) {
	errs := &multierror.Error{}

	resourceIDs := ord.ResourceIDs{
		PackageIDs:               make(map[string]bool),
		PackagePolicyLevels:      resourceID.PackagePolicyLevels,
		BundleIDs:                make(map[string]bool),
		ProductIDs:               make(map[string]bool),
		APIIDs:                   make(map[string]bool),
		EventIDs:                 make(map[string]bool),
		EntityTypeIDs:            make(map[string]bool),
		VendorIDs:                make(map[string]bool),
		CapabilityIDs:            make(map[string]bool),
		IntegrationDependencyIDs: make(map[string]bool),
		DataProductIDs:           make(map[string]bool),
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
		invalidEntityTypesIndices := make([]int, 0)
		invalidCapabilitiesIndices := make([]int, 0)
		invalidIntegrationDependenciesIndices := make([]int, 0)
		invalidDataProductsIndices := make([]int, 0)

		if err := validateDocumentInput(doc); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "error validating document"))
		}

		for i, pkg := range doc.Packages {
			if err := validatePackageInputWithSuppressedErrors(pkg, doc.PolicyLevel, resourcesFromDB.Packages, resourceHashes); err != nil {
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
			if err := validateBundleInput(bndl, credentialExchangeStrategyTenantMappings, doc.PolicyLevel); err != nil {
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
			if err := validateProductInput(product, doc.PolicyLevel); err != nil {
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

		for i, entityType := range doc.EntityTypes {
			if err := validateEntityTypeInputWithSuppressedErrors(entityType, doc.PolicyLevel, resourcesFromDB.EntityTypes, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating entity type with ord id %q", entityType.OrdID))
			}
			if err := validateEntityTypeInput(entityType, doc.PolicyLevel); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating entity type with ord id %q", entityType.OrdID))
				invalidEntityTypesIndices = append(invalidEntityTypesIndices, i)
				continue
			}

			if _, ok := resourceIDs.EventIDs[entityType.OrdID]; ok && forbidDuplications {
				errs = multierror.Append(errs, errors.Errorf("found duplicate event with ord id %q", entityType.OrdID))
			}

			resourceIDs.EventIDs[entityType.OrdID] = true
		}

		for i, capability := range doc.Capabilities {
			if err := validateCapabilityInputWithSuppressedErrors(capability, doc.PolicyLevel, resourcesFromDB.Capabilities, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating capability with ord id %q", stringPtrToString(capability.OrdID)))
			}
			if err := validateCapabilityInput(capability, doc.PolicyLevel); err != nil {
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
			if err := validateIntegrationDependencyInputWithSuppressedErrors(integrationDependency, doc.PolicyLevel, resourcesFromDB.IntegrationDependencies, resourceHashes); err != nil {
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

		for i, dataProduct := range doc.DataProducts {
			if err := validateDataProductInputWithSuppressedErrors(dataProduct, doc.PolicyLevel, resourcesFromDB.DataProducts, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "suppressed errors validating data product with ord id %q", stringPtrToString(dataProduct.OrdID)))
			}
			if err := validateDataProductInput(dataProduct, doc.PolicyLevel); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating data product with ord id %q", stringPtrToString(dataProduct.OrdID)))
				invalidDataProductsIndices = append(invalidDataProductsIndices, i)
				continue
			}
			if dataProduct.OrdID != nil {
				if _, ok := resourceIDs.DataProductIDs[*dataProduct.OrdID]; ok && forbidDuplications {
					errs = multierror.Append(errs, errors.Errorf("found duplicate data product with ord id %q", *dataProduct.OrdID))
				}
				resourceIDs.DataProductIDs[*dataProduct.OrdID] = true
			}
		}

		for i, vendor := range doc.Vendors {
			if err := validateVendorInput(vendor, doc.PolicyLevel); err != nil {
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
		doc.EntityTypes = deleteInvalidInputObjects(invalidEntityTypesIndices, doc.EntityTypes)
		doc.Vendors = deleteInvalidInputObjects(invalidVendorsIndices, doc.Vendors)
		doc.Tombstones = deleteInvalidInputObjects(invalidTombstonesIndices, doc.Tombstones)
		doc.Capabilities = deleteInvalidInputObjects(invalidCapabilitiesIndices, doc.Capabilities)
		doc.IntegrationDependencies = deleteInvalidInputObjects(invalidIntegrationDependenciesIndices, doc.IntegrationDependencies)
		doc.DataProducts = deleteInvalidInputObjects(invalidDataProductsIndices, doc.DataProducts)
	}

	return ord.ResourceIDs{
		PackageIDs:               resourceIDs.PackageIDs,
		ProductIDs:               resourceIDs.ProductIDs,
		APIIDs:                   resourceIDs.APIIDs,
		EventIDs:                 resourceIDs.EventIDs,
		EntityTypeIDs:            resourceIDs.EntityTypeIDs,
		VendorIDs:                resourceIDs.VendorIDs,
		BundleIDs:                resourceIDs.BundleIDs,
		PackagePolicyLevels:      resourceIDs.PackagePolicyLevels,
		CapabilityIDs:            resourceIDs.CapabilityIDs,
		IntegrationDependencyIDs: resourceIDs.IntegrationDependencyIDs,
		DataProductIDs:           resourceIDs.DataProductIDs,
	}, errs
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
