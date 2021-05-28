package open_resource_discovery

import (
	"encoding/json"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	WellKnownEndpoint = "/.well-known/open-resource-discovery"
)

type WellKnownConfig struct {
	Schema                  string                  `json:"$schema"`
	OpenResourceDiscoveryV1 OpenResourceDiscoveryV1 `json:"openResourceDiscoveryV1"`
}

type OpenResourceDiscoveryV1 struct {
	Documents []DocumentDetails `json:"documents"`
}

type DocumentDetails struct {
	URL string `json:"url"`
	// TODO: Currently we cannot differentiate between system instance types reliably, therefore we cannot make use of the systemInstanceAware optimization (store it once per system type and reuse it for each system instance of that type).
	//  Once we have system landscape discovery and stable system types we can make use of this optimization. Until then we store all the information for a system instance as it is provided in the documents.
	//  Therefore we treat every resource as SystemInstanceAware = true
	SystemInstanceAware bool             `json:"systemInstanceAware"`
	AccessStrategies    AccessStrategies `json:"accessStrategies"`
}

type Document struct {
	Schema                string `json:"$schema"`
	OpenResourceDiscovery string `json:"openResourceDiscovery"`
	Description           string `json:"description"`

	// TODO: In the current state of ORD and it's implementation we are missing system landscape discovery and an id correlation in the system instances. Because of that in the first phase we will rely on:
	//  - DescribedSystemInstance is the application in our DB and it's baseURL should match with the one in the webhook.
	//  - ProviderSystemInstance is not supported since we do not support information of a system instance to be provided by a different system instance due to missing correlation.
	DescribedSystemInstance *model.Application `json:"describedSystemInstance"`
	ProviderSystemInstance  *model.Application `json:"providerSystemInstance"`

	Packages           []*model.PackageInput         `json:"packages"`
	ConsumptionBundles []*model.BundleCreateInput    `json:"consumptionBundles"`
	Products           []*model.ProductInput         `json:"products"`
	APIResources       []*model.APIDefinitionInput   `json:"apiResources"`
	EventResources     []*model.EventDefinitionInput `json:"eventResources"`
	Tombstones         []*model.TombstoneInput       `json:"tombstones"`
	Vendors            []*model.VendorInput          `json:"vendors"`
}

type Documents []*Document

// Validate validates all the documents for a system instance
func (docs Documents) Validate(webhookURL string) error {
	// TODO: Revisit after DescribedSystemInstance vs. ProviderSystemInstance is aligned. Currently we rely on that described system instance is identical with the provider system instance. See TODO above.
	for _, doc := range docs {
		if doc.ProviderSystemInstance != nil {
			return errors.New("providerSystemInstance not supported")
		}
		if doc.DescribedSystemInstance != nil && doc.DescribedSystemInstance.BaseURL != nil && *doc.DescribedSystemInstance.BaseURL != webhookURL {
			return errors.New("describedSystemInstance should be the same as the one providing the documents or providerSystemInstance should be defined")
		}
	}

	packageIDs := make(map[string]bool, 0)
	packagePolicyLevels := make(map[string]string, 0)
	bundleIDs := make(map[string]bool, 0)
	productIDs := make(map[string]bool, 0)
	apiIDs := make(map[string]bool, 0)
	eventIDs := make(map[string]bool, 0)
	vendorIDs := make(map[string]bool, 0)

	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if _, ok := packageIDs[pkg.OrdID]; ok {
				return errors.Errorf("found duplicate package with ord id %q", pkg.OrdID)
			}
			packageIDs[pkg.OrdID] = true
			packagePolicyLevels[pkg.OrdID] = pkg.PolicyLevel
		}
	}

	for _, doc := range docs {
		if err := validateDocumentInput(doc); err != nil {
			return errors.Wrap(err, "error validating document")
		}

		for _, pkg := range doc.Packages {
			if err := validatePackageInput(pkg); err != nil {
				return errors.Wrapf(err, "error validating package with ord id %q", pkg.OrdID)
			}
		}
		for _, bndl := range doc.ConsumptionBundles {
			if err := validateBundleInput(bndl); err != nil {
				return errors.Wrapf(err, "error validating bundle with ord id %q", stringPtrToString(bndl.OrdID))
			}
			if _, ok := bundleIDs[*bndl.OrdID]; ok {
				return errors.Errorf("found duplicate bundle with ord id %q", *bndl.OrdID)
			}
			bundleIDs[*bndl.OrdID] = true
		}
		for _, product := range doc.Products {
			if err := validateProductInput(product); err != nil {
				return errors.Wrapf(err, "error validating product with ord id %q", product.OrdID)
			}
			if _, ok := productIDs[product.OrdID]; ok {
				return errors.Errorf("found duplicate product with ord id %q", product.OrdID)
			}
			productIDs[product.OrdID] = true
		}
		for _, api := range doc.APIResources {
			if err := validateAPIInput(api, packagePolicyLevels); err != nil {
				return errors.Wrapf(err, "error validating api with ord id %q", stringPtrToString(api.OrdID))
			}
			if _, ok := apiIDs[*api.OrdID]; ok {
				return errors.Errorf("found duplicate api with ord id %q", *api.OrdID)
			}
			apiIDs[*api.OrdID] = true
		}
		for _, event := range doc.EventResources {
			if err := validateEventInput(event); err != nil {
				return errors.Wrapf(err, "error validating event with ord id %q", stringPtrToString(event.OrdID))
			}
			if _, ok := eventIDs[*event.OrdID]; ok {
				return errors.Errorf("found duplicate event with ord id %q", *event.OrdID)
			}
			eventIDs[*event.OrdID] = true
		}
		for _, vendor := range doc.Vendors {
			if err := validateVendorInput(vendor); err != nil {
				return errors.Wrapf(err, "error validating vendor with ord id %q", vendor.OrdID)
			}
			if _, ok := vendorIDs[vendor.OrdID]; ok {
				return errors.Errorf("found duplicate vendor with ord id %q", vendor.OrdID)
			}
			vendorIDs[vendor.OrdID] = true
		}
		for _, tombstone := range doc.Tombstones {
			if err := validateTombstoneInput(tombstone); err != nil {
				return errors.Wrapf(err, "error validating tombstone with ord id %q", tombstone.OrdID)
			}
		}
	}

	// Validate entity relations
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if !vendorIDs[*pkg.Vendor] {
				return errors.Errorf("package with id %q has a reference to unknown vendor %q", pkg.OrdID, *pkg.Vendor)
			}
			ordIDs := gjson.ParseBytes(pkg.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !productIDs[productID.String()] {
					return errors.Errorf("package with id %q has a reference to unknown product %q", pkg.OrdID, productID.String())
				}
			}
		}
		for _, product := range doc.Products {
			if !vendorIDs[product.Vendor] {
				return errors.Errorf("product with id %q has a reference to unknown vendor %q", product.OrdID, product.Vendor)
			}
		}
		for _, api := range doc.APIResources {
			if !packageIDs[*api.OrdPackageID] {
				return errors.Errorf("api with id %q has a reference to unknown package %q", *api.OrdID, *api.OrdPackageID)
			}
			if api.PartOfConsumptionBundles != nil {
				for _, apiBndlRef := range api.PartOfConsumptionBundles {
					if !bundleIDs[apiBndlRef.BundleOrdID] {
						return errors.Errorf("api with id %q has a reference to unknown bundle %q", *api.OrdID, apiBndlRef.BundleOrdID)
					}
				}
			}

			ordIDs := gjson.ParseBytes(api.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !productIDs[productID.String()] {
					return errors.Errorf("api with id %q has a reference to unknown product %q", *api.OrdID, productID.String())
				}
			}
		}
		for _, event := range doc.EventResources {
			if !packageIDs[*event.OrdPackageID] {
				return errors.Errorf("event with id %q has a reference to unknown package %q", *event.OrdID, *event.OrdPackageID)
			}
			if event.PartOfConsumptionBundles != nil {
				for _, eventBndlRef := range event.PartOfConsumptionBundles {
					if !bundleIDs[eventBndlRef.BundleOrdID] {
						return errors.Errorf("event with id %q has a reference to unknown bundle %q", *event.OrdID, eventBndlRef.BundleOrdID)
					}
				}
			}

			ordIDs := gjson.ParseBytes(event.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !productIDs[productID.String()] {
					return errors.Errorf("event with id %q has a reference to unknown product %q", *event.OrdID, productID.String())
				}
			}
		}
		for _, tombstone := range doc.Tombstones {
			if !packageIDs[tombstone.OrdID] && !bundleIDs[tombstone.OrdID] && !productIDs[tombstone.OrdID] &&
				!apiIDs[tombstone.OrdID] && !eventIDs[tombstone.OrdID] && !vendorIDs[tombstone.OrdID] {
				return errors.Errorf("tombstone with id %q for an unknown entity", tombstone.OrdID)
			}
		}
	}

	// TODO: Validate that every change to a resource leads to version increment. If a resource in the document is different from the one in the DB and both have the same versions, then this is a validation error.

	return nil
}

// Sanitize performs all the merging and rewriting rules defined in ORD. This method should be invoked after Documents are validated with the Validate method.
//  - Rewrite all relative URIs using the baseURL from the Described System Instance. If the Described System Instance baseURL is missing the provider baseURL (from the webhook) is used.
//  - Package's partOfProducts, tags, countries, industry, lineOfBusiness, labels are inherited by the resources in the package.
//  - Ensure to assign `defaultEntryPoint` if missing and there are available `entryPoints` to API's `PartOfConsumptionBundles`
func (docs Documents) Sanitize(baseURL string) error {
	var err error

	// Rewrite relative URIs
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if pkg.PackageLinks, err = rewriteRelativeURIsInJson(pkg.PackageLinks, baseURL, "url"); err != nil {
				return err
			}
			if pkg.Links, err = rewriteRelativeURIsInJson(pkg.Links, baseURL, "url"); err != nil {
				return err
			}
		}

		for _, bndl := range doc.ConsumptionBundles {
			if bndl.Links, err = rewriteRelativeURIsInJson(bndl.Links, baseURL, "url"); err != nil {
				return err
			}
			if bndl.CredentialExchangeStrategies, err = rewriteRelativeURIsInJson(bndl.CredentialExchangeStrategies, baseURL, "callbackUrl"); err != nil {
				return err
			}
		}

		for _, api := range doc.APIResources {
			for _, definition := range api.ResourceDefinitions {
				if !isAbsoluteURL(definition.URL) {
					definition.URL = baseURL + definition.URL
				}
			}
			if api.APIResourceLinks, err = rewriteRelativeURIsInJson(api.APIResourceLinks, baseURL, "url"); err != nil {
				return err
			}
			if api.Links, err = rewriteRelativeURIsInJson(api.Links, baseURL, "url"); err != nil {
				return err
			}
			if api.ChangeLogEntries, err = rewriteRelativeURIsInJson(api.ChangeLogEntries, baseURL, "url"); err != nil {
				return err
			}
			if api.TargetURLs, err = rewriteRelativeURIsInJsonArray(api.TargetURLs, baseURL); err != nil {
				return err
			}
			rewriteDefaultTargetURL(api.PartOfConsumptionBundles, baseURL)
		}

		for _, event := range doc.EventResources {
			if event.ChangeLogEntries, err = rewriteRelativeURIsInJson(event.ChangeLogEntries, baseURL, "url"); err != nil {
				return err
			}
			if event.Links, err = rewriteRelativeURIsInJson(event.Links, baseURL, "url"); err != nil {
				return err
			}
			for _, definition := range event.ResourceDefinitions {
				if !isAbsoluteURL(definition.URL) {
					definition.URL = baseURL + definition.URL
				}
			}
		}
	}

	// Package properties inheritance
	packages := make(map[string]*model.PackageInput, 0)
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			packages[pkg.OrdID] = pkg
		}
	}

	for _, doc := range docs {
		for _, api := range doc.APIResources {
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

func rewriteRelativeURIsInJsonArray(j json.RawMessage, baseURL string) (json.RawMessage, error) {
	parsedJson := gjson.ParseBytes(j)

	items := make([]interface{}, 0, 0)
	for _, crrURI := range parsedJson.Array() {
		if !isAbsoluteURL(crrURI.String()) {
			rewrittenURI := baseURL + crrURI.String()

			items = append(items, rewrittenURI)
		} else {
			items = append(items, crrURI.String())
		}
	}

	rewrittenJson, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	return rewrittenJson, nil
}

func rewriteDefaultTargetURL(bundleRefs []*model.ConsumptionBundleReference, baseURL string) {
	for _, br := range bundleRefs {
		if br.DefaultTargetURL != "" && !isAbsoluteURL(br.DefaultTargetURL) {
			br.DefaultTargetURL = baseURL + br.DefaultTargetURL
		}
	}
}

func rewriteRelativeURIsInJson(j json.RawMessage, baseURL, jsonPath string) (json.RawMessage, error) {
	parsedJson := gjson.ParseBytes(j)
	if parsedJson.IsArray() {
		items := make([]interface{}, 0, 0)
		for _, jsonElement := range parsedJson.Array() {
			rewrittenElement, err := rewriteRelativeURIsInJson(json.RawMessage(jsonElement.Raw), baseURL, jsonPath)
			if err != nil {
				return nil, err
			}
			items = append(items, gjson.ParseBytes(rewrittenElement).Value())
		}
		return json.Marshal(items)
	} else if parsedJson.IsObject() {
		uriProperty := gjson.GetBytes(j, jsonPath)
		if uriProperty.Exists() && !isAbsoluteURL(uriProperty.String()) {
			return sjson.SetBytes(j, jsonPath, baseURL+uriProperty.String())
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
