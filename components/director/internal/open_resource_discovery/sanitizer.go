package ord

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const unknownReferenceCode = "sap-ord-unknown-reference"
const duplicateResourceCode = "sap-ord-duplicate-resource"

// DocumentSanitizer represents the sanitizer of ORD documents
type DocumentSanitizer struct {
}

// NewDocumentSanitizer returns new instance of a document sanitizer
func NewDocumentSanitizer() DocumentSanitizer {
	return DocumentSanitizer{}
}

// Sanitize performs all the merging and rewriting rules defined in ORD. This method should be invoked after Documents are validated with the Validate method.
//   - Rewrite all relative URIs using the baseURL from the Described System Instance. If the Described System Instance baseURL is missing the provider baseURL (from the webhook) is used.
//   - Package's partOfProducts, tags, countries, industry, lineOfBusiness, labels are inherited by the resources in the package.
//   - Ensure to assign `defaultEntryPoint` if missing and there are available `entryPoints` to API's `PartOfConsumptionBundles`
//   - If some resource(Package, API, Event or Data Product) doesn't have provided `policyLevel` and `customPolicyLevel`, these are inherited from the document
func (v *DocumentSanitizer) Sanitize(docs []*Document, webhookBaseURL, webhookBaseProxyURL string) ([]*ValidationError, error) {
	valErrors := make([]*ValidationError, 0)

	var err error

	// Use the ProxyURL for all relative link substitution except for the API's TargetURLs.
	// They are externally consumable and we should not expose those URLs through the Proxy but rather from webhook's BaseURL
	baseURL := webhookBaseURL
	if webhookBaseProxyURL != "" {
		baseURL = webhookBaseProxyURL
	}

	// Rewrite relative URIs
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if pkg.PackageLinks, err = rewriteRelativeURIsInJSON(pkg.PackageLinks, baseURL, "url"); err != nil {
				return valErrors, err
			}
			if pkg.Links, err = rewriteRelativeURIsInJSON(pkg.Links, baseURL, "url"); err != nil {
				return valErrors, err
			}
		}

		for _, bndl := range doc.ConsumptionBundles {
			if bndl.Links, err = rewriteRelativeURIsInJSON(bndl.Links, baseURL, "url"); err != nil {
				return valErrors, err
			}
			if bndl.CredentialExchangeStrategies, err = rewriteRelativeURIsInJSON(bndl.CredentialExchangeStrategies, baseURL, "callbackUrl"); err != nil {
				return valErrors, err
			}
		}

		for _, api := range doc.APIResources {
			for _, definition := range api.ResourceDefinitions {
				definition.URL, err = constructResourceDefinitionURL(baseURL, definition.URL)
				if err != nil {
					return nil, err
				}
			}
			if api.APIResourceLinks, err = rewriteRelativeURIsInJSON(api.APIResourceLinks, baseURL, "url"); err != nil {
				return valErrors, err
			}
			if api.Links, err = rewriteRelativeURIsInJSON(api.Links, baseURL, "url"); err != nil {
				return valErrors, err
			}
			if api.ChangeLogEntries, err = rewriteRelativeURIsInJSON(api.ChangeLogEntries, baseURL, "url"); err != nil {
				return valErrors, err
			}
			if api.TargetURLs, err = rewriteRelativeURIsInJSONArray(api.TargetURLs, webhookBaseURL); err != nil {
				return valErrors, err
			}
			if err = rewriteDefaultTargetURL(api.PartOfConsumptionBundles, baseURL); err != nil {
				return valErrors, err
			}
		}

		for _, event := range doc.EventResources {
			if event.ChangeLogEntries, err = rewriteRelativeURIsInJSON(event.ChangeLogEntries, baseURL, "url"); err != nil {
				return valErrors, err
			}
			if event.EventResourceLinks, err = rewriteRelativeURIsInJSON(event.EventResourceLinks, baseURL, "url"); err != nil {
				return valErrors, err
			}
			if event.Links, err = rewriteRelativeURIsInJSON(event.Links, baseURL, "url"); err != nil {
				return valErrors, err
			}
			for _, definition := range event.ResourceDefinitions {
				definition.URL, err = constructResourceDefinitionURL(baseURL, definition.URL)
				if err != nil {
					return valErrors, err
				}
			}
		}

		for _, entityType := range doc.EntityTypes {
			if entityType.ChangeLogEntries, err = rewriteRelativeURIsInJSON(entityType.ChangeLogEntries, baseURL, "url"); err != nil {
				return valErrors, err
			}
			if entityType.Links, err = rewriteRelativeURIsInJSON(entityType.Links, baseURL, "url"); err != nil {
				return valErrors, err
			}
		}

		for _, capability := range doc.Capabilities {
			for _, definition := range capability.CapabilityDefinitions {
				definition.URL, err = constructResourceDefinitionURL(baseURL, definition.URL)
				if err != nil {
					return valErrors, err
				}
			}

			if capability.Links, err = rewriteRelativeURIsInJSON(capability.Links, baseURL, "url"); err != nil {
				return valErrors, err
			}
		}

		for _, integrationDependency := range doc.IntegrationDependencies {
			if integrationDependency.Links, err = rewriteRelativeURIsInJSON(integrationDependency.Links, baseURL, "url"); err != nil {
				return valErrors, err
			}
		}

		for _, dataProduct := range doc.DataProducts {
			if dataProduct.DataProductLinks, err = rewriteRelativeURIsInJSON(dataProduct.DataProductLinks, baseURL, "url"); err != nil {
				return valErrors, err
			}

			if dataProduct.ChangeLogEntries, err = rewriteRelativeURIsInJSON(dataProduct.ChangeLogEntries, baseURL, "url"); err != nil {
				return valErrors, err
			}

			if dataProduct.Links, err = rewriteRelativeURIsInJSON(dataProduct.Links, baseURL, "url"); err != nil {
				return valErrors, err
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
				valErrors = append(valErrors, newCustomValidationError(*api.OrdID, unknownReferenceCode, fmt.Sprintf("The api has a reference to unknown package %q", *api.OrdPackageID)))
				continue
			}
			if api.PartOfProducts, err = mergeJSONArraysOfStrings(referredPkg.PartOfProducts, api.PartOfProducts); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging partOfProducts for api with ord id %q", *api.OrdID)
			}
			if api.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, api.Tags); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging tags for api with ord id %q", *api.OrdID)
			}
			if api.Countries, err = mergeJSONArraysOfStrings(referredPkg.Countries, api.Countries); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging countries for api with ord id %q", *api.OrdID)
			}
			if api.Industry, err = mergeJSONArraysOfStrings(referredPkg.Industry, api.Industry); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging industry for api with ord id %q", *api.OrdID)
			}
			if api.LineOfBusiness, err = mergeJSONArraysOfStrings(referredPkg.LineOfBusiness, api.LineOfBusiness); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging lineOfBusiness for api with ord id %q", *api.OrdID)
			}
			if api.Labels, err = mergeORDLabels(referredPkg.Labels, api.Labels); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging labels for api with ord id %q", *api.OrdID)
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
				valErrors = append(valErrors, newCustomValidationError(*event.OrdID, unknownReferenceCode, fmt.Sprintf("The event has a reference to unknown package %q", *event.OrdPackageID)))
				continue
			}
			if event.PartOfProducts, err = mergeJSONArraysOfStrings(referredPkg.PartOfProducts, event.PartOfProducts); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging partOfProducts for event with ord id %q", *event.OrdID)
			}
			if event.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, event.Tags); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging tags for event with ord id %q", *event.OrdID)
			}
			if event.Countries, err = mergeJSONArraysOfStrings(referredPkg.Countries, event.Countries); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging countries for event with ord id %q", *event.OrdID)
			}
			if event.Industry, err = mergeJSONArraysOfStrings(referredPkg.Industry, event.Industry); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging industry for event with ord id %q", *event.OrdID)
			}
			if event.LineOfBusiness, err = mergeJSONArraysOfStrings(referredPkg.LineOfBusiness, event.LineOfBusiness); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging lineOfBusiness for event with ord id %q", *event.OrdID)
			}
			if event.Labels, err = mergeORDLabels(referredPkg.Labels, event.Labels); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging labels for event with ord id %q", *event.OrdID)
			}
		}

		for _, entityType := range doc.EntityTypes {
			if entityType.PolicyLevel == nil {
				entityType.PolicyLevel = doc.PolicyLevel
				entityType.CustomPolicyLevel = doc.CustomPolicyLevel
			}

			referredPkg, ok := packages[entityType.OrdPackageID]
			if !ok {
				valErrors = append(valErrors, newCustomValidationError(entityType.OrdID, unknownReferenceCode, fmt.Sprintf("The entity type has a reference to unknown package %q", entityType.OrdPackageID)))
				continue
			}
			if entityType.PartOfProducts, err = mergeJSONArraysOfStrings(referredPkg.PartOfProducts, entityType.PartOfProducts); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging partOfProducts for entity type with ord id %q", entityType.OrdID)
			}
			if entityType.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, entityType.Tags); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging tags for entity type with ord id %q", entityType.OrdID)
			}
			if entityType.Labels, err = mergeORDLabels(referredPkg.Labels, entityType.Labels); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging labels for entity type with ord id %q", entityType.OrdID)
			}
		}

		for _, capability := range doc.Capabilities {
			referredPkg, ok := packages[*capability.OrdPackageID]
			if !ok {
				valErrors = append(valErrors, newCustomValidationError(*capability.OrdID, unknownReferenceCode, fmt.Sprintf("The capability has a reference to unknown package %q", *capability.OrdPackageID)))
				continue
			}
			if capability.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, capability.Tags); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging tags for capability with ord id %q", *capability.OrdID)
			}
			if capability.Labels, err = mergeORDLabels(referredPkg.Labels, capability.Labels); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging labels for capability with ord id %q", *capability.OrdID)
			}
		}
		for _, integrationDependency := range doc.IntegrationDependencies {
			referredPkg, ok := packages[*integrationDependency.OrdPackageID]
			if !ok {
				valErrors = append(valErrors, newCustomValidationError(*integrationDependency.OrdID, unknownReferenceCode, fmt.Sprintf("The integration dependency has a reference to unknown package %q", *integrationDependency.OrdPackageID)))
				continue
			}
			if integrationDependency.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, integrationDependency.Tags); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging tags for integration dependency with ord id %q", *integrationDependency.OrdID)
			}
			if integrationDependency.Labels, err = mergeORDLabels(referredPkg.Labels, integrationDependency.Labels); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging labels for integration dependency with ord id %q", *integrationDependency.OrdID)
			}
		}
		for _, dataProduct := range doc.DataProducts {
			if dataProduct.PolicyLevel == nil {
				dataProduct.PolicyLevel = doc.PolicyLevel
				dataProduct.CustomPolicyLevel = doc.CustomPolicyLevel
			}

			referredPkg, ok := packages[*dataProduct.OrdPackageID]
			if !ok {
				valErrors = append(valErrors, newCustomValidationError(*dataProduct.OrdID, unknownReferenceCode, fmt.Sprintf("The data product has a reference to unknown package %q", *dataProduct.OrdPackageID)))
				continue
			}
			if dataProduct.Tags, err = mergeJSONArraysOfStrings(referredPkg.Tags, dataProduct.Tags); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging tags for data product with ord id %q", *dataProduct.OrdID)
			}
			if dataProduct.Industry, err = mergeJSONArraysOfStrings(referredPkg.Industry, dataProduct.Industry); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging industry for data product with ord id %q", *dataProduct.OrdID)
			}
			if dataProduct.LineOfBusiness, err = mergeJSONArraysOfStrings(referredPkg.LineOfBusiness, dataProduct.LineOfBusiness); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging lineOfBusiness for data product with ord id %q", *dataProduct.OrdID)
			}
			if dataProduct.Labels, err = mergeORDLabels(referredPkg.Labels, dataProduct.Labels); err != nil {
				return valErrors, errors.Wrapf(err, "error while merging labels for data product with ord id %q", *dataProduct.OrdID)
			}
		}
	}

	return valErrors, err
}

func constructResourceDefinitionURL(baseURL, definitionURL string) (string, error) {
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse base URL")
	}

	defURL, err := url.Parse(definitionURL)
	if err != nil {
		return "", err
	}

	return parsedBaseURL.ResolveReference(defURL).String(), nil
}

func newCustomValidationError(ordID, errorType, description string) *ValidationError {
	return &ValidationError{
		OrdID:       ordID,
		Severity:    ErrorSeverity,
		Type:        errorType,
		Description: description,
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

// mergeJSONArraysOfStrings merges arr2 in arr1
func mergeJSONArraysOfStrings(arr1, arr2 json.RawMessage) (json.RawMessage, error) {
	if len(arr1) == 0 {
		return arr2, nil
	}
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

func isAbsoluteURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
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
		rewrittenURI, err := constructResourceDefinitionURL(baseURL, crrURI.String())
		if err != nil {
			return nil, err
		}

		items = append(items, rewrittenURI)
	}

	rewrittenJSON, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	return rewrittenJSON, nil
}

func rewriteDefaultTargetURL(bundleRefs []*model.ConsumptionBundleReference, baseURL string) error {
	for _, br := range bundleRefs {
		if br.DefaultTargetURL != "" {
			var err error
			br.DefaultTargetURL, err = constructResourceDefinitionURL(baseURL, br.DefaultTargetURL)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// mergeORDLabels merges labels2 into labels1
func mergeORDLabels(labels1, labels2 json.RawMessage) (json.RawMessage, error) {
	if len(labels1) == 0 {
		return labels2, nil
	}

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

func assignDefaultEntryPointIfNeeded(bundleReferences []*model.ConsumptionBundleReference, targetURLs json.RawMessage) {
	lenTargetURLs := len(gjson.ParseBytes(targetURLs).Array())
	for _, br := range bundleReferences {
		if br.DefaultTargetURL == "" && lenTargetURLs > 1 {
			br.DefaultTargetURL = gjson.ParseBytes(targetURLs).Array()[0].String()
		}
	}
}
