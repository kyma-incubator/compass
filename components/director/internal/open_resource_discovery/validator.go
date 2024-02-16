package ord

import (
	"context"
	"dario.cat/mergo"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"strconv"
)

// Validator validates list of ORD documents
//
//go:generate mockery --name=Validator --output=automock --outpkg=automock --case=underscore --disable-version-string
type Validator interface {
	Validate(ctx context.Context, documents []*Document, baseURL string, globalResourcesOrdIDs map[string]bool, docsString []string) ([]ValidationError, error)
}

type DocumentValidator struct {
	client *ValidationClient
}

func NewDocumentValidator(client *ValidationClient) *DocumentValidator {
	return &DocumentValidator{
		client: client,
	}
}

func (v *DocumentValidator) Validate(ctx context.Context, documents []*Document, baseURL string, globalResourcesOrdIDs map[string]bool, docsString []string) ([]ValidationError, error) {
	var result []ValidationError

	for i := range documents {
		// validation errors coming from API Metadata validator
		log.C(ctx).Info("Calling API Metadata Validator for document")

		errors1, err := v.client.Validate("sap:base:v1", docsString[i])
		if err != nil {
			return nil, errors.Wrap(err, "while validating document with API Metadata validator")
		}

		var data interface{}
		err = json.Unmarshal([]byte(docsString[i]), &data)
		currentDocumentErrors := v.toValidationErrors(data, errors1)

		result = append(result, currentDocumentErrors...)

		deleteInvalidResourcesFromDocument(documents[i], currentDocumentErrors)
	}

	log.C(ctx).Info("Check for duplicate resources and entity relations")
	// UCL validations - check for duplicates and unknown package reference
	errors2, err := v.ValidateOld(documents, baseURL, globalResourcesOrdIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while validating document")
	}

	result = append(result, errors2...)

	return result, nil
}

func findResourceOrdIdByPath(data interface{}, path []string) (string, error) {
	current := data
	var ordId string

	for _, key := range path {
		switch t := current.(type) {
		case map[string]interface{}:
			val, ok := t[key]
			if !ok {
				return "", errors.New("Key not found in map")
			}
			fmt.Println(path, "path")
			if path[0] == "packages" {
				fmt.Println(val, "IF PKGg")
			}
			current = val
			ordIdVal, exists := t["ordId"]
			if exists {
				ordId = ordIdVal.(string)
			}
		case []interface{}:
			idx, err := strconv.Atoi(key)
			if err != nil {
				return "", errors.New("Invalid index in array")
			}
			if idx < 0 || idx >= len(t) {
				return "", errors.New("Index out of range")
			}
			current = t[idx]
			switch c := current.(type) {
			case map[string]interface{}:
				ordIdVal, exists := c["ordId"]
				if exists {
					ordId = ordIdVal.(string)
				}
			}

		default:
			return "", errors.New("Invalid data type")
		}
	}

	return ordId, nil
}

func deleteInvalidResourcesFromDocument(document *Document, documentErrors []ValidationError) {
	for _, e := range documentErrors {
		if e.Severity != ErrorSeverity {
			continue
		}
		for apiIdx, api := range document.APIResources {
			if *api.OrdID == e.OrdId {
				document.APIResources = append(document.APIResources[:apiIdx], document.APIResources[apiIdx+1:]...)
				continue
			}
		}

		for eventIdx, event := range document.EventResources {
			if *event.OrdID == e.OrdId {
				document.EventResources = append(document.EventResources[:eventIdx], document.EventResources[eventIdx+1:]...)
				continue
			}
		}

		for entityTypeIdx, entityType := range document.EntityTypes {
			if entityType.OrdID == e.OrdId {
				document.EntityTypes = append(document.EntityTypes[:entityTypeIdx], document.EntityTypes[entityTypeIdx+1:]...)
				continue
			}
		}

		for capabilityIdx, capability := range document.Capabilities {
			if *capability.OrdID == e.OrdId {
				document.Capabilities = append(document.Capabilities[:capabilityIdx], document.Capabilities[capabilityIdx+1:]...)
				continue
			}
		}

		for dataProductIdx, dataProduct := range document.DataProducts {
			if *dataProduct.OrdID == e.OrdId {
				document.DataProducts = append(document.DataProducts[:dataProductIdx], document.DataProducts[dataProductIdx+1:]...)
				continue
			}
		}

		for integrationDependencyIdx, integrationDependency := range document.IntegrationDependencies {
			if *integrationDependency.OrdID == e.OrdId {
				document.IntegrationDependencies = append(document.IntegrationDependencies[:integrationDependencyIdx], document.IntegrationDependencies[integrationDependencyIdx+1:]...)
				continue
			}
		}

		for vendorIdx, vendor := range document.Vendors {
			if vendor.OrdID == e.OrdId {
				document.Vendors = append(document.Vendors[:vendorIdx], document.Vendors[vendorIdx+1:]...)
				continue
			}
		}

		for productIdx, product := range document.Products {
			if product.OrdID == e.OrdId {
				document.Products = append(document.Products[:productIdx], document.Products[productIdx+1:]...)
				continue
			}
		}

		for pkgIdx, pkg := range document.Packages {
			if pkg.OrdID == e.OrdId {
				document.Packages = append(document.Packages[:pkgIdx], document.Packages[pkgIdx+1:]...)
				continue
			}
		}

		for bundleIdx, bundle := range document.ConsumptionBundles {
			if *bundle.OrdID == e.OrdId {
				document.ConsumptionBundles = append(document.ConsumptionBundles[:bundleIdx], document.ConsumptionBundles[bundleIdx+1:]...)
				continue
			}
		}

		for tombstoneIdx, tombstone := range document.Tombstones {
			if tombstone.OrdID == e.OrdId {
				document.Tombstones = append(document.Tombstones[:tombstoneIdx], document.Tombstones[tombstoneIdx+1:]...)
				continue
			}
		}
	}
}

func (v *DocumentValidator) toValidationErrors(document interface{}, result []ValidationResult) []ValidationError {
	valErrs := make([]ValidationError, 0)

	for _, r := range result {
		fmt.Println(r.Path)
		ordId, _ := findResourceOrdIdByPath(document, r.Path)
		valErrs = append(valErrs, ValidationError{
			OrdId:       ordId,
			Severity:    r.Severity,
			Type:        r.Code,
			Description: r.Message,
		})
	}
	return valErrs
}

// ValidateOld validates all the documents for a system instance
func (v *DocumentValidator) ValidateOld(docs []*Document, calculatedBaseURL string, globalResourcesOrdIDs map[string]bool) ([]ValidationError, error) {
	var (
		baseURL             = calculatedBaseURL
		isBaseURLConfigured = len(calculatedBaseURL) > 0
		validationErrors    []ValidationError
	)

	for _, doc := range docs {
		if !isBaseURLConfigured && (doc.DescribedSystemInstance == nil || doc.DescribedSystemInstance.BaseURL == nil) {
			validationErrors = append(validationErrors, newCustomValidationError("", ErrorSeverity, "no baseURL provided", "no baseURL was provided neither from /well-known URL, nor from config, nor from describedSystemInstance"))
			continue
		}

		if len(baseURL) == 0 {
			baseURL = *doc.DescribedSystemInstance.BaseURL
		}

		if doc.DescribedSystemInstance != nil && doc.DescribedSystemInstance.BaseURL != nil && *doc.DescribedSystemInstance.BaseURL != baseURL {
			validationErrors = append(validationErrors, newCustomValidationError("", ErrorSeverity, "", fmt.Sprintf("describedSystemInstance should be the same as the one providing the documents - %s : %s", *doc.DescribedSystemInstance.BaseURL, baseURL)))
		}
	}

	resourceIDs := ResourceIDs{
		PackageIDs:          make(map[string]bool),
		PackagePolicyLevels: make(map[string]string),
	}

	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if _, ok := resourceIDs.PackageIDs[pkg.OrdID]; ok {
				validationErrors = append(validationErrors, newCustomValidationError(pkg.OrdID, ErrorSeverity, duplicateResourceCode, fmt.Sprintf("found duplicate package with ord id %q", pkg.OrdID)))
				continue
			}
			resourceIDs.PackageIDs[pkg.OrdID] = true
			if pkg.PolicyLevel != nil {
				resourceIDs.PackagePolicyLevels[pkg.OrdID] = *pkg.PolicyLevel
			}
		}
	}

	r1, e1 := v.checkForDuplications(docs, SystemVersionPerspective, true, resourceIDs.PackagePolicyLevels)
	r2, e2 := v.checkForDuplications(docs, SystemInstancePerspective, true, resourceIDs.PackagePolicyLevels)
	r3, e3 := v.checkForDuplications(docs, "", false, resourceIDs.PackagePolicyLevels)

	if err := mergo.Merge(&resourceIDs, r1); err != nil {
		return validationErrors, err
	}
	if err := mergo.Merge(&resourceIDs, r2); err != nil {
		return validationErrors, err
	}
	if err := mergo.Merge(&resourceIDs, r3); err != nil {
		return validationErrors, err
	}

	validationErrors = append(validationErrors, e1...)
	validationErrors = append(validationErrors, e2...)
	validationErrors = append(validationErrors, e3...)

	_, e4 := v.checkEntityRelations(docs, resourceIDs, globalResourcesOrdIDs, resourceIDs.PackagePolicyLevels)

	validationErrors = append(validationErrors, e4...)

	return validationErrors, nil
}

func (v *DocumentValidator) checkEntityRelations(docs []*Document, resourceIDs ResourceIDs, globalResourcesOrdIDs map[string]bool, packagePolicyLevels map[string]string) (ResourceIDs, []ValidationError) {
	var validationErrors []ValidationError

	invalidApisIndices := make([]int, 0)
	invalidEventsIndices := make([]int, 0)
	invalidEntityTypesIndices := make([]int, 0)
	invalidCapabilitiesIndices := make([]int, 0)
	invalidIntegrationDependenciesIndices := make([]int, 0)
	invalidDataProductsIndices := make([]int, 0)

	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if pkg.Vendor != nil && !resourceIDs.VendorIDs[*pkg.Vendor] && !globalResourcesOrdIDs[*pkg.Vendor] {
				validationErrors = append(validationErrors, newCustomValidationError(pkg.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("package with id %q has a reference to unknown vendor %q", pkg.OrdID, *pkg.Vendor)))
			}
			ordIDs := gjson.ParseBytes(pkg.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					validationErrors = append(validationErrors, newCustomValidationError(pkg.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("package with id %q has a reference to unknown product %q", pkg.OrdID, productID.String())))
				}
			}
		}
		for _, product := range doc.Products {
			if !resourceIDs.VendorIDs[product.Vendor] && !globalResourcesOrdIDs[product.Vendor] {
				validationErrors = append(validationErrors, newCustomValidationError(product.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("product with id %q has a reference to unknown vendor %q", product.OrdID, product.Vendor)))
			}
		}
		for i, api := range doc.APIResources {
			if api.OrdPackageID != nil && !resourceIDs.PackageIDs[*api.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*api.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("api with id %q has a reference to unknown package %q", *api.OrdID, *api.OrdPackageID)))
				invalidApisIndices = append(invalidApisIndices, i)
			}
			if api.PartOfConsumptionBundles != nil {
				for _, apiBndlRef := range api.PartOfConsumptionBundles {
					if !resourceIDs.BundleIDs[apiBndlRef.BundleOrdID] {
						validationErrors = append(validationErrors, newCustomValidationError(*api.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("api with id %q has a reference to unknown bundle %q", *api.OrdID, apiBndlRef.BundleOrdID)))
					}
				}
			}

			ordIDs := gjson.ParseBytes(api.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					validationErrors = append(validationErrors, newCustomValidationError(*api.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("api with id %q has a reference to unknown product %q", *api.OrdID, productID.String())))
				}
			}
		}

		for i, event := range doc.EventResources {
			if event.OrdPackageID != nil && !resourceIDs.PackageIDs[*event.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*event.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("event with id %q has a reference to unknown package %q", *event.OrdID, *event.OrdPackageID)))

				invalidEventsIndices = append(invalidEventsIndices, i)
			}
			if event.PartOfConsumptionBundles != nil {
				for _, eventBndlRef := range event.PartOfConsumptionBundles {
					if !resourceIDs.BundleIDs[eventBndlRef.BundleOrdID] {
						validationErrors = append(validationErrors, newCustomValidationError(*event.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("event with id %q has a reference to unknown bundle %q", *event.OrdID, eventBndlRef.BundleOrdID)))

					}
				}
			}

			ordIDs := gjson.ParseBytes(event.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					validationErrors = append(validationErrors, newCustomValidationError(*event.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("event with id %q has a reference to unknown product %q", *event.OrdID, productID.String())))
				}
			}
		}

		for i, entityType := range doc.EntityTypes {
			if !resourceIDs.PackageIDs[entityType.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(entityType.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("entity type with id %q has a reference to unknown package %q", entityType.OrdID, entityType.OrdPackageID)))

				invalidEntityTypesIndices = append(invalidEntityTypesIndices, i)
			}

			ordIDs := gjson.ParseBytes(entityType.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					validationErrors = append(validationErrors, newCustomValidationError(entityType.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("entity type with id %q has a reference to unknown product %q", entityType.OrdID, productID.String())))

				}
			}
		}

		for i, capability := range doc.Capabilities {
			if capability.OrdPackageID != nil && !resourceIDs.PackageIDs[*capability.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*capability.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("capability with id %q has a reference to unknown package %q", *capability.OrdID, *capability.OrdPackageID)))

				invalidCapabilitiesIndices = append(invalidCapabilitiesIndices, i)
			}
		}

		for i, integrationDependency := range doc.IntegrationDependencies {
			if integrationDependency.OrdPackageID != nil && !resourceIDs.PackageIDs[*integrationDependency.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*integrationDependency.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("integration dependency with id %q has a reference to unknown package %q", *integrationDependency.OrdID, *integrationDependency.OrdPackageID)))
				invalidIntegrationDependenciesIndices = append(invalidIntegrationDependenciesIndices, i)
			}
		}

		for i, dataProduct := range doc.DataProducts {
			if dataProduct.OrdPackageID != nil && !resourceIDs.PackageIDs[*dataProduct.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*dataProduct.OrdID, ErrorSeverity, unknownReferenceCode, fmt.Sprintf("data product with id %q has a reference to unknown package %q", *dataProduct.OrdID, *dataProduct.OrdPackageID)))

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

	return ResourceIDs{
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
	}, validationErrors

}

func (v *DocumentValidator) checkForDuplications(docs []*Document, perspectiveConstraint DocumentPerspective, forbidDuplications bool, packagePolicyLevels map[string]string) (ResourceIDs, []ValidationError) {
	var validationErrors []ValidationError

	resourceIDs := ResourceIDs{
		PackageIDs:               make(map[string]bool),
		PackagePolicyLevels:      packagePolicyLevels,
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

		for _, bndl := range doc.ConsumptionBundles {
			if bndl.OrdID != nil {
				if _, ok := resourceIDs.BundleIDs[*bndl.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*bndl.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate bundle"))
				}
				resourceIDs.BundleIDs[*bndl.OrdID] = true
			}
		}

		for _, product := range doc.Products {
			if _, ok := resourceIDs.ProductIDs[product.OrdID]; ok && forbidDuplications {
				validationErrors = append(validationErrors, newCustomValidationError(product.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate product"))
			}
			resourceIDs.ProductIDs[product.OrdID] = true
		}

		for _, api := range doc.APIResources {
			if api.OrdID != nil {
				if _, ok := resourceIDs.APIIDs[*api.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*api.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate api"))
				}
				resourceIDs.APIIDs[*api.OrdID] = true
			}
		}

		for _, event := range doc.EventResources {
			if event.OrdID != nil {
				if _, ok := resourceIDs.EventIDs[*event.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*event.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate event"))
				}

				resourceIDs.EventIDs[*event.OrdID] = true
			}
		}

		for _, entityType := range doc.EntityTypes {
			if _, ok := resourceIDs.EventIDs[entityType.OrdID]; ok && forbidDuplications {
				validationErrors = append(validationErrors, newCustomValidationError(entityType.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate entity type"))
			}

			resourceIDs.EventIDs[entityType.OrdID] = true
		}

		for _, capability := range doc.Capabilities {
			if capability.OrdID != nil {
				if _, ok := resourceIDs.CapabilityIDs[*capability.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*capability.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate capability"))
				}
				resourceIDs.CapabilityIDs[*capability.OrdID] = true
			}
		}

		for _, integrationDependency := range doc.IntegrationDependencies {
			if integrationDependency.OrdID != nil {
				if _, ok := resourceIDs.IntegrationDependencyIDs[*integrationDependency.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*integrationDependency.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate integration dependency"))
				}
				resourceIDs.IntegrationDependencyIDs[*integrationDependency.OrdID] = true
			}
		}

		for _, dataProduct := range doc.DataProducts {
			if dataProduct.OrdID != nil {
				if _, ok := resourceIDs.DataProductIDs[*dataProduct.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*dataProduct.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate data product"))
				}
				resourceIDs.DataProductIDs[*dataProduct.OrdID] = true
			}
		}

		for _, vendor := range doc.Vendors {
			if _, ok := resourceIDs.VendorIDs[vendor.OrdID]; ok && forbidDuplications {
				validationErrors = append(validationErrors, newCustomValidationError(vendor.OrdID, ErrorSeverity, duplicateResourceCode, "duplicate vendor"))
			}
			resourceIDs.VendorIDs[vendor.OrdID] = true
		}

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

	return ResourceIDs{
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
	}, validationErrors
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
