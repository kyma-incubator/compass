package ord

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"dario.cat/mergo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Validator validates list of ORD documents
//
//go:generate mockery --name=Validator --output=automock --outpkg=automock --case=underscore --disable-version-string
type Validator interface {
	Validate(ctx context.Context, documents []*Document, baseURL string, globalResourcesOrdIDs map[string]bool, docsString []string, ruleset, appNamespace string) ([]*ValidationError, error)
}

// DocumentValidator validates the ORD documents
type DocumentValidator struct {
	client                            ValidatorClient
	ordRuleIgnoreListMapping          map[string][]string
	validationRuleGlobalIgnoreListKey string
}

// NewDocumentValidator returns new document validator for validating ORD documents
func NewDocumentValidator(client ValidatorClient, ordRuleIgnoreListMapping map[string][]string, validationRuleGlobalIgnoreListKey string) Validator {
	return &DocumentValidator{
		client:                            client,
		ordRuleIgnoreListMapping:          ordRuleIgnoreListMapping,
		validationRuleGlobalIgnoreListKey: validationRuleGlobalIgnoreListKey,
	}
}

// Validate validates all ORD documents with the API Metadata Validator and checks resource duplications and entity relations
func (v *DocumentValidator) Validate(ctx context.Context, documents []*Document, baseURL string, globalResourcesOrdIDs map[string]bool, documentsAsString []string, ruleset, appNamespace string) ([]*ValidationError, error) {
	var combinedValidationErrors []*ValidationError

	for i := range documents {
		log.C(ctx).Info("Validate document with API Metadata Validator")
		errorsFromAPIMetadataValidator, err := v.client.Validate(ctx, ruleset, documentsAsString[i])
		if err != nil {
			return nil, errors.Wrap(err, "error while validating document with API Metadata validator")
		}

		var data interface{}
		err = json.Unmarshal([]byte(documentsAsString[i]), &data)
		if err != nil {
			return nil, errors.Wrap(err, "error while unmarshalling ORD documents as string")
		}

		currentDocumentErrors := v.toValidationErrors(data, errorsFromAPIMetadataValidator)

		if len(currentDocumentErrors) > 0 {
			log.C(ctx).Infof("There are %d validation errors from API Metadata Validator", len(currentDocumentErrors))

			ordRuleIgnoreList := v.determineIgnoreList(appNamespace)
			deleteInvalidResourcesFromDocument(ctx, documents[i], currentDocumentErrors, ordRuleIgnoreList, appNamespace)

			combinedValidationErrors = append(combinedValidationErrors, currentDocumentErrors...)
		}

		errorsFromORDConfig := validateORDConfigurations(documents[i], baseURL)
		if errorsFromORDConfig != nil {
			combinedValidationErrors = append(combinedValidationErrors, errorsFromORDConfig)
		}
	}

	resourceIDs := ResourceIDs{
		PackageIDs:          make(map[string]bool),
		PackagePolicyLevels: make(map[string]string),
	}

	duplicationErrors, err := v.checkForDuplications(documents, &resourceIDs)
	if err != nil {
		return nil, err
	}

	resourceReferenceErrors := v.checkEntityRelations(documents, &resourceIDs, globalResourcesOrdIDs)

	combinedValidationErrors = append(combinedValidationErrors, duplicationErrors...)
	combinedValidationErrors = append(combinedValidationErrors, resourceReferenceErrors...)

	return combinedValidationErrors, nil
}

// determineIgnoreList Merge the namespaced ignorelist with the global ignorelist.
// Global here means namespace-independent ignorelist
func (v *DocumentValidator) determineIgnoreList(appNamespace string) []string {
	globalIgnoreList := v.ordRuleIgnoreListMapping[v.validationRuleGlobalIgnoreListKey]
	namespacedOrdRuleIgnoreList := v.ordRuleIgnoreListMapping[appNamespace]

	return str.MergeWithoutDuplicates(globalIgnoreList, namespacedOrdRuleIgnoreList)
}

func findResourceOrdIDByPath(data interface{}, path []string) (string, error) {
	current := data
	var ordID string

	for _, key := range path {
		switch t := current.(type) {
		case map[string]interface{}:
			val, ok := t[key]
			if !ok {
				return "", errors.New("Key not found in map")
			}

			current = val
			ordIDVal, exists := t["ordId"]
			if exists {
				ordID = ordIDVal.(string)
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

			if c, ok := current.(map[string]interface{}); ok {
				if ordIDVal, exists := c["ordId"]; exists {
					ordID = ordIDVal.(string)
				}
			}
		default:
			return "", errors.New("Invalid data type")
		}
	}

	return ordID, nil
}

func deleteInvalidResourcesFromDocument(ctx context.Context, document *Document, documentErrors []*ValidationError, ordRuleIgnoreList []string, appNamespace string) {
	for _, e := range documentErrors {
		if str.ContainsInSlice(ordRuleIgnoreList, e.Type) {
			log.C(ctx).Infof("ORD error %q is present in document, but it is ignored because an ignorelist mapping exist for app namespace %q", e.Type, appNamespace)
			continue
		}

		if e.Severity != ErrorSeverity {
			continue
		}
		for apiIdx, api := range document.APIResources {
			if *api.OrdID == e.OrdID {
				document.APIResources = append(document.APIResources[:apiIdx], document.APIResources[apiIdx+1:]...)
				continue
			}
		}

		for eventIdx, event := range document.EventResources {
			if *event.OrdID == e.OrdID {
				document.EventResources = append(document.EventResources[:eventIdx], document.EventResources[eventIdx+1:]...)
				continue
			}
		}

		for entityTypeIdx, entityType := range document.EntityTypes {
			if entityType.OrdID == e.OrdID {
				document.EntityTypes = append(document.EntityTypes[:entityTypeIdx], document.EntityTypes[entityTypeIdx+1:]...)
				continue
			}
		}

		for capabilityIdx, capability := range document.Capabilities {
			if *capability.OrdID == e.OrdID {
				document.Capabilities = append(document.Capabilities[:capabilityIdx], document.Capabilities[capabilityIdx+1:]...)
				continue
			}
		}

		for dataProductIdx, dataProduct := range document.DataProducts {
			if *dataProduct.OrdID == e.OrdID {
				document.DataProducts = append(document.DataProducts[:dataProductIdx], document.DataProducts[dataProductIdx+1:]...)
				continue
			}
		}

		for integrationDependencyIdx, integrationDependency := range document.IntegrationDependencies {
			if *integrationDependency.OrdID == e.OrdID {
				document.IntegrationDependencies = append(document.IntegrationDependencies[:integrationDependencyIdx], document.IntegrationDependencies[integrationDependencyIdx+1:]...)
				continue
			}
		}

		for vendorIdx, vendor := range document.Vendors {
			if vendor.OrdID == e.OrdID {
				document.Vendors = append(document.Vendors[:vendorIdx], document.Vendors[vendorIdx+1:]...)
				continue
			}
		}

		for productIdx, product := range document.Products {
			if product.OrdID == e.OrdID {
				document.Products = append(document.Products[:productIdx], document.Products[productIdx+1:]...)
				continue
			}
		}

		for pkgIdx, pkg := range document.Packages {
			if pkg.OrdID == e.OrdID {
				document.Packages = append(document.Packages[:pkgIdx], document.Packages[pkgIdx+1:]...)
				continue
			}
		}

		for bundleIdx, bundle := range document.ConsumptionBundles {
			if *bundle.OrdID == e.OrdID {
				document.ConsumptionBundles = append(document.ConsumptionBundles[:bundleIdx], document.ConsumptionBundles[bundleIdx+1:]...)
				continue
			}
		}

		for tombstoneIdx, tombstone := range document.Tombstones {
			if tombstone.OrdID == e.OrdID {
				document.Tombstones = append(document.Tombstones[:tombstoneIdx], document.Tombstones[tombstoneIdx+1:]...)
				continue
			}
		}
	}
}

func (v *DocumentValidator) toValidationErrors(document interface{}, validationResults []model.ValidationResult) []*ValidationError {
	valErrs := make([]*ValidationError, 0)

	for _, valResult := range validationResults {
		ordID, _ := findResourceOrdIDByPath(document, valResult.Path)
		valErrs = append(valErrs, &ValidationError{
			OrdID:       ordID,
			Severity:    valResult.Severity,
			Type:        valResult.Code,
			Description: valResult.Message,
		})
	}
	return valErrs
}

func validateORDConfigurations(doc *Document, calculatedBaseURL string) *ValidationError {
	var (
		baseURL             = calculatedBaseURL
		isBaseURLConfigured = len(calculatedBaseURL) > 0
	)

	if !isBaseURLConfigured && (doc.DescribedSystemInstance == nil || doc.DescribedSystemInstance.BaseURL == nil || *doc.DescribedSystemInstance.BaseURL == "") {
		return newCustomValidationError("", "sap-ord-no-base-url", "no baseURL was provided neither from /well-known URL, nor from config, nor from describedSystemInstance")
	}

	if len(baseURL) == 0 {
		baseURL = *doc.DescribedSystemInstance.BaseURL
	}

	if doc.DescribedSystemInstance != nil && doc.DescribedSystemInstance.BaseURL != nil && *doc.DescribedSystemInstance.BaseURL != baseURL {
		return newCustomValidationError("", "sap-ord-baseUrl-mismatch", fmt.Sprintf("describedSystemInstance should be the same as the one providing the documents - %s : %s", *doc.DescribedSystemInstance.BaseURL, baseURL))
	}

	return nil
}

func (v *DocumentValidator) checkForDuplications(docs []*Document, resourceIDs *ResourceIDs) ([]*ValidationError, error) {
	var validationErrors []*ValidationError

	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if _, ok := resourceIDs.PackageIDs[pkg.OrdID]; ok {
				validationErrors = append(validationErrors, newCustomValidationError(pkg.OrdID, duplicateResourceCode, fmt.Sprintf("found duplicate package with ord id %q", pkg.OrdID)))
				continue
			}
			resourceIDs.PackageIDs[pkg.OrdID] = true
			if pkg.PolicyLevel != nil {
				resourceIDs.PackagePolicyLevels[pkg.OrdID] = *pkg.PolicyLevel
			}
		}
	}

	r1, e1 := v.checkForDuplicationsWithPerspective(docs, SystemVersionPerspective, true, resourceIDs.PackagePolicyLevels)
	r2, e2 := v.checkForDuplicationsWithPerspective(docs, SystemInstancePerspective, true, resourceIDs.PackagePolicyLevels)
	r3, e3 := v.checkForDuplicationsWithPerspective(docs, "", false, resourceIDs.PackagePolicyLevels)

	if err := mergo.Merge(resourceIDs, r1); err != nil {
		return validationErrors, err
	}
	if err := mergo.Merge(resourceIDs, r2); err != nil {
		return validationErrors, err
	}
	if err := mergo.Merge(resourceIDs, r3); err != nil {
		return validationErrors, err
	}

	validationErrors = append(validationErrors, e1...)
	validationErrors = append(validationErrors, e2...)
	validationErrors = append(validationErrors, e3...)

	return validationErrors, nil
}

func (v *DocumentValidator) checkEntityRelations(docs []*Document, resourceIDs *ResourceIDs, globalResourcesOrdIDs map[string]bool) []*ValidationError {
	var validationErrors []*ValidationError

	invalidApisIndices := make([]int, 0)
	invalidEventsIndices := make([]int, 0)
	invalidEntityTypesIndices := make([]int, 0)
	invalidCapabilitiesIndices := make([]int, 0)
	invalidIntegrationDependenciesIndices := make([]int, 0)
	invalidDataProductsIndices := make([]int, 0)

	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if pkg.Vendor != nil && !resourceIDs.VendorIDs[*pkg.Vendor] && !globalResourcesOrdIDs[*pkg.Vendor] {
				validationErrors = append(validationErrors, newCustomValidationError(pkg.OrdID, unknownReferenceCode, fmt.Sprintf("The package has a reference to unknown vendor %q", *pkg.Vendor)))
			}
			ordIDs := gjson.ParseBytes(pkg.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					validationErrors = append(validationErrors, newCustomValidationError(pkg.OrdID, unknownReferenceCode, fmt.Sprintf("The package has a reference to unknown product %q", productID.String())))
				}
			}
		}
		for _, product := range doc.Products {
			if !resourceIDs.VendorIDs[product.Vendor] && !globalResourcesOrdIDs[product.Vendor] {
				validationErrors = append(validationErrors, newCustomValidationError(product.OrdID, unknownReferenceCode, fmt.Sprintf("The product has a reference to unknown vendor %q", product.Vendor)))
			}
		}
		for i, api := range doc.APIResources {
			if api.OrdPackageID != nil && !resourceIDs.PackageIDs[*api.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*api.OrdID, unknownReferenceCode, fmt.Sprintf("The api has a reference to unknown package %q", *api.OrdPackageID)))

				invalidApisIndices = append(invalidApisIndices, i)
			}
			if api.PartOfConsumptionBundles != nil {
				for _, apiBndlRef := range api.PartOfConsumptionBundles {
					if !resourceIDs.BundleIDs[apiBndlRef.BundleOrdID] {
						validationErrors = append(validationErrors, newCustomValidationError(*api.OrdID, unknownReferenceCode, fmt.Sprintf("The api has a reference to unknown bundle %q", apiBndlRef.BundleOrdID)))
					}
				}
			}

			ordIDs := gjson.ParseBytes(api.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					validationErrors = append(validationErrors, newCustomValidationError(*api.OrdID, unknownReferenceCode, fmt.Sprintf("The api has a reference to unknown product %q", productID.String())))
				}
			}
		}

		for i, event := range doc.EventResources {
			if event.OrdPackageID != nil && !resourceIDs.PackageIDs[*event.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*event.OrdID, unknownReferenceCode, fmt.Sprintf("The event has a reference to unknown package %q", *event.OrdPackageID)))

				invalidEventsIndices = append(invalidEventsIndices, i)
			}
			if event.PartOfConsumptionBundles != nil {
				for _, eventBndlRef := range event.PartOfConsumptionBundles {
					if !resourceIDs.BundleIDs[eventBndlRef.BundleOrdID] {
						validationErrors = append(validationErrors, newCustomValidationError(*event.OrdID, unknownReferenceCode, fmt.Sprintf("The event has a reference to unknown bundle %q", eventBndlRef.BundleOrdID)))
					}
				}
			}

			ordIDs := gjson.ParseBytes(event.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					validationErrors = append(validationErrors, newCustomValidationError(*event.OrdID, unknownReferenceCode, fmt.Sprintf("The event has a reference to unknown product %q", productID.String())))
				}
			}
		}

		for i, entityType := range doc.EntityTypes {
			if !resourceIDs.PackageIDs[entityType.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(entityType.OrdID, unknownReferenceCode, fmt.Sprintf("The entity type has a reference to unknown package %q", entityType.OrdPackageID)))

				invalidEntityTypesIndices = append(invalidEntityTypesIndices, i)
			}

			ordIDs := gjson.ParseBytes(entityType.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !resourceIDs.ProductIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					validationErrors = append(validationErrors, newCustomValidationError(entityType.OrdID, unknownReferenceCode, fmt.Sprintf("The entity type has a reference to unknown product %q", productID.String())))
				}
			}
		}

		for i, capability := range doc.Capabilities {
			if capability.OrdPackageID != nil && !resourceIDs.PackageIDs[*capability.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*capability.OrdID, unknownReferenceCode, fmt.Sprintf("The capability has a reference to unknown package %q", *capability.OrdPackageID)))

				invalidCapabilitiesIndices = append(invalidCapabilitiesIndices, i)
			}
		}

		for i, integrationDependency := range doc.IntegrationDependencies {
			if integrationDependency.OrdPackageID != nil && !resourceIDs.PackageIDs[*integrationDependency.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*integrationDependency.OrdID, unknownReferenceCode, fmt.Sprintf("The integration dependency has a reference to unknown package %q", *integrationDependency.OrdPackageID)))

				invalidIntegrationDependenciesIndices = append(invalidIntegrationDependenciesIndices, i)
			}
		}

		for i, dataProduct := range doc.DataProducts {
			if dataProduct.OrdPackageID != nil && !resourceIDs.PackageIDs[*dataProduct.OrdPackageID] {
				validationErrors = append(validationErrors, newCustomValidationError(*dataProduct.OrdID, unknownReferenceCode, fmt.Sprintf("The data product has a reference to unknown package %q", *dataProduct.OrdPackageID)))

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

	return validationErrors
}

func (v *DocumentValidator) checkForDuplicationsWithPerspective(docs []*Document, perspectiveConstraint DocumentPerspective, forbidDuplications bool, packagePolicyLevels map[string]string) (ResourceIDs, []*ValidationError) {
	var validationErrors []*ValidationError

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

		for i, bndl := range doc.ConsumptionBundles {
			if bndl.OrdID != nil {
				if _, ok := resourceIDs.BundleIDs[*bndl.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*bndl.OrdID, duplicateResourceCode, "duplicate bundle"))

					invalidBundlesIndices = append(invalidProductsIndices, i)
				}
				resourceIDs.BundleIDs[*bndl.OrdID] = true
			}
		}

		for i, product := range doc.Products {
			if _, ok := resourceIDs.ProductIDs[product.OrdID]; ok && forbidDuplications {
				validationErrors = append(validationErrors, newCustomValidationError(product.OrdID, duplicateResourceCode, "duplicate product"))

				invalidProductsIndices = append(invalidProductsIndices, i)
			}
			resourceIDs.ProductIDs[product.OrdID] = true
		}

		for i, api := range doc.APIResources {
			if api.OrdID != nil {
				if _, ok := resourceIDs.APIIDs[*api.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*api.OrdID, duplicateResourceCode, "duplicate api"))

					invalidApisIndices = append(invalidApisIndices, i)
				}
				resourceIDs.APIIDs[*api.OrdID] = true
			}
		}

		for i, event := range doc.EventResources {
			if event.OrdID != nil {
				if _, ok := resourceIDs.EventIDs[*event.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*event.OrdID, duplicateResourceCode, "duplicate event"))

					invalidEventsIndices = append(invalidEventsIndices, i)
				}
				resourceIDs.EventIDs[*event.OrdID] = true
			}
		}

		for i, entityType := range doc.EntityTypes {
			if _, ok := resourceIDs.EventIDs[entityType.OrdID]; ok && forbidDuplications {
				validationErrors = append(validationErrors, newCustomValidationError(entityType.OrdID, duplicateResourceCode, "duplicate entity type"))
				invalidEntityTypesIndices = append(invalidEntityTypesIndices, i)
			}
			resourceIDs.EventIDs[entityType.OrdID] = true
		}

		for i, capability := range doc.Capabilities {
			if capability.OrdID != nil {
				if _, ok := resourceIDs.CapabilityIDs[*capability.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*capability.OrdID, duplicateResourceCode, "duplicate capability"))

					invalidCapabilitiesIndices = append(invalidCapabilitiesIndices, i)
				}
				resourceIDs.CapabilityIDs[*capability.OrdID] = true
			}
		}

		for i, integrationDependency := range doc.IntegrationDependencies {
			if integrationDependency.OrdID != nil {
				if _, ok := resourceIDs.IntegrationDependencyIDs[*integrationDependency.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*integrationDependency.OrdID, duplicateResourceCode, "duplicate integration dependency"))

					invalidIntegrationDependenciesIndices = append(invalidIntegrationDependenciesIndices, i)
				}
				resourceIDs.IntegrationDependencyIDs[*integrationDependency.OrdID] = true
			}
		}

		for i, dataProduct := range doc.DataProducts {
			if dataProduct.OrdID != nil {
				if _, ok := resourceIDs.DataProductIDs[*dataProduct.OrdID]; ok && forbidDuplications {
					validationErrors = append(validationErrors, newCustomValidationError(*dataProduct.OrdID, duplicateResourceCode, "duplicate data product"))

					invalidDataProductsIndices = append(invalidDataProductsIndices, i)
				}
				resourceIDs.DataProductIDs[*dataProduct.OrdID] = true
			}
		}

		for i, vendor := range doc.Vendors {
			if _, ok := resourceIDs.VendorIDs[vendor.OrdID]; ok && forbidDuplications {
				validationErrors = append(validationErrors, newCustomValidationError(vendor.OrdID, duplicateResourceCode, "duplicate vendor"))

				invalidVendorsIndices = append(invalidVendorsIndices, i)
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
