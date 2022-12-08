package ord

import (
	"encoding/json"
	"net/url"
	"regexp"

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
}

// Document represents an ORD Document
type Document struct {
	Schema                string `json:"$schema"`
	OpenResourceDiscovery string `json:"openResourceDiscovery"`
	Description           string `json:"description"`

	// TODO: In the current state of ORD and it's implementation we are missing system landscape discovery and an id correlation in the system instances. Because of that in the first phase we will rely on:
	//  - DescribedSystemInstance is the application in our DB and it's baseURL should match with the one in the webhook.
	DescribedSystemInstance *model.Application `json:"describedSystemInstance"`

	Packages           []*model.PackageInput         `json:"packages"`
	ConsumptionBundles []*model.BundleCreateInput    `json:"consumptionBundles"`
	Products           []*model.ProductInput         `json:"products"`
	APIResources       []*model.APIDefinitionInput   `json:"apiResources"`
	EventResources     []*model.EventDefinitionInput `json:"eventResources"`
	Tombstones         []*model.TombstoneInput       `json:"tombstones"`
	Vendors            []*model.VendorInput          `json:"vendors"`
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

// Validate validates all the documents for a system instance
func (docs Documents) Validate(calculatedBaseURL string, apisFromDB map[string]*model.APIDefinition, eventsFromDB map[string]*model.EventDefinition, packagesFromDB map[string]*model.Package, resourceHashes map[string]uint64, globalResourcesOrdIDs map[string]bool) error {
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
		if doc.DescribedSystemInstance != nil && doc.DescribedSystemInstance.BaseURL != nil && *doc.DescribedSystemInstance.BaseURL != baseURL {
			errs = multierror.Append(errs, errors.New("describedSystemInstance should be the same as the one providing the documents"))
		}
	}

	packageIDs := make(map[string]bool)
	packagePolicyLevels := make(map[string]string)
	bundleIDs := make(map[string]bool)
	productIDs := make(map[string]bool)
	apiIDs := make(map[string]bool)
	eventIDs := make(map[string]bool)
	vendorIDs := make(map[string]bool)

	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if _, ok := packageIDs[pkg.OrdID]; ok {
				errs = multierror.Append(errs, errors.Errorf("found duplicate package with ord id %q", pkg.OrdID))
				continue
			}
			packageIDs[pkg.OrdID] = true
			packagePolicyLevels[pkg.OrdID] = pkg.PolicyLevel
		}
	}

	for _, doc := range docs {
		if err := validateDocumentInput(doc); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "error validating document"))
		}

		for _, pkg := range doc.Packages {
			if err := validatePackageInput(pkg, packagesFromDB, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating package with ord id %q", pkg.OrdID))
			}
		}

		finalBundles := make([]*model.BundleCreateInput, 0)
		for i, bndl := range doc.ConsumptionBundles {
			if bndl.Name == "" {
				continue
			}

			if err := validateBundleInput(bndl); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating bundle with ord id %q", stringPtrToString(bndl.OrdID)))
			}
			if bndl.OrdID != nil {
				if _, ok := bundleIDs[*bndl.OrdID]; ok {
					errs = multierror.Append(errs, errors.Errorf("found duplicate bundle with ord id %q", *bndl.OrdID))
				}
				bundleIDs[*bndl.OrdID] = true
			}
			finalBundles = append(finalBundles, doc.ConsumptionBundles[i])
		}
		doc.ConsumptionBundles = finalBundles
		for _, product := range doc.Products {
			if err := validateProductInput(product); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating product with ord id %q", product.OrdID))
			}
			if _, ok := productIDs[product.OrdID]; ok {
				errs = multierror.Append(errs, errors.Errorf("found duplicate product with ord id %q", product.OrdID))
			}
			productIDs[product.OrdID] = true
		}

		finalAPIs := make([]*model.APIDefinitionInput, 0)
		for i, api := range doc.APIResources {
			foundBndl := false
			for _, bndl := range api.PartOfConsumptionBundles {
				for _, finalBndl := range finalBundles {
					if *finalBndl.OrdID == bndl.BundleOrdID {
						foundBndl = true
						break
					}
				}
				if foundBndl {
					break
				}
			}

			if !foundBndl {
				continue
			}

			if err := validateAPIInput(api, packagePolicyLevels, apisFromDB, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating api with ord id %q", stringPtrToString(api.OrdID)))
			}
			if api.OrdID != nil {
				if _, ok := apiIDs[*api.OrdID]; ok {
					errs = multierror.Append(errs, errors.Errorf("found duplicate api with ord id %q", *api.OrdID))
				}
				apiIDs[*api.OrdID] = true
			}
			finalAPIs = append(finalAPIs, doc.APIResources[i])
		}
		doc.APIResources = finalAPIs

		eventsIndicesWithEmptyName := make([]int, 0)

		for i, event := range doc.EventResources {
			foundBndl := false
			for _, bndl := range event.PartOfConsumptionBundles {
				for _, finalBndl := range finalBundles {
					if *finalBndl.OrdID == bndl.BundleOrdID {
						foundBndl = true
						break
					}
				}
				if foundBndl {
					break
				}
			}

			if event.Name == "" || !foundBndl {
				eventsIndicesWithEmptyName = append(eventsIndicesWithEmptyName, i)
				continue
			}

			if err := validateEventInput(event, packagePolicyLevels, eventsFromDB, resourceHashes); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating event with ord id %q", stringPtrToString(event.OrdID)))
			}

			if event.OrdID != nil {
				if _, ok := eventIDs[*event.OrdID]; ok {
					errs = multierror.Append(errs, errors.Errorf("found duplicate event with ord id %q", *event.OrdID))
				}

				eventIDs[*event.OrdID] = true
			}
		}

		decreaseIndexForDeleting := 0
		for eventIndex := range eventsIndicesWithEmptyName {
			deleteIndex := eventIndex - decreaseIndexForDeleting
			doc.EventResources = append(doc.EventResources[:deleteIndex], doc.EventResources[deleteIndex+1:]...)
			decreaseIndexForDeleting++
		}

		for _, vendor := range doc.Vendors {
			if err := validateVendorInput(vendor); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating vendor with ord id %q", vendor.OrdID))
			}
			if _, ok := vendorIDs[vendor.OrdID]; ok {
				errs = multierror.Append(errs, errors.Errorf("found duplicate vendor with ord id %q", vendor.OrdID))
			}
			vendorIDs[vendor.OrdID] = true
		}
		for _, tombstone := range doc.Tombstones {
			if err := validateTombstoneInput(tombstone); err != nil {
				errs = multierror.Append(errs, errors.Wrapf(err, "error validating tombstone with ord id %q", tombstone.OrdID))
			}
		}
	}

	// Validate entity relations
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if pkg.Vendor != nil && !vendorIDs[*pkg.Vendor] && !globalResourcesOrdIDs[*pkg.Vendor] {
				errs = multierror.Append(errs, errors.Errorf("package with id %q has a reference to unknown vendor %q", pkg.OrdID, *pkg.Vendor))
			}
			ordIDs := gjson.ParseBytes(pkg.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !productIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					errs = multierror.Append(errs, errors.Errorf("package with id %q has a reference to unknown product %q", pkg.OrdID, productID.String()))
				}
			}
		}
		for _, product := range doc.Products {
			if !vendorIDs[product.Vendor] && !globalResourcesOrdIDs[product.Vendor] {
				errs = multierror.Append(errs, errors.Errorf("product with id %q has a reference to unknown vendor %q", product.OrdID, product.Vendor))
			}
		}
		for _, api := range doc.APIResources {
			if api.OrdPackageID != nil && !packageIDs[*api.OrdPackageID] {
				errs = multierror.Append(errs, errors.Errorf("api with id %q has a reference to unknown package %q", *api.OrdID, *api.OrdPackageID))
			}
			if api.PartOfConsumptionBundles != nil {
				for _, apiBndlRef := range api.PartOfConsumptionBundles {
					if !bundleIDs[apiBndlRef.BundleOrdID] {
						errs = multierror.Append(errs, errors.Errorf("api with id %q has a reference to unknown bundle %q", *api.OrdID, apiBndlRef.BundleOrdID))
					}
				}
			}

			ordIDs := gjson.ParseBytes(api.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !productIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					errs = multierror.Append(errs, errors.Errorf("api with id %q has a reference to unknown product %q", *api.OrdID, productID.String()))
				}
			}
		}
		for _, event := range doc.EventResources {
			if event.OrdPackageID != nil && !packageIDs[*event.OrdPackageID] {
				errs = multierror.Append(errs, errors.Errorf("event with id %q has a reference to unknown package %q", *event.OrdID, *event.OrdPackageID))
			}
			if event.PartOfConsumptionBundles != nil {
				for _, eventBndlRef := range event.PartOfConsumptionBundles {
					if !bundleIDs[eventBndlRef.BundleOrdID] {
						errs = multierror.Append(errs, errors.Errorf("event with id %q has a reference to unknown bundle %q", *event.OrdID, eventBndlRef.BundleOrdID))
					}
				}
			}

			ordIDs := gjson.ParseBytes(event.PartOfProducts).Array()
			for _, productID := range ordIDs {
				if !productIDs[productID.String()] && !globalResourcesOrdIDs[productID.String()] {
					errs = multierror.Append(errs, errors.Errorf("event with id %q has a reference to unknown product %q", *event.OrdID, productID.String()))
				}
			}
		}
	}

	return errs.ErrorOrNil()
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
			if pkg.PackageLinks, err = rewriteRelativeURIsInJSON(pkg.PackageLinks, baseURL, "url"); err != nil {
				return err
			}
			if pkg.Links, err = rewriteRelativeURIsInJSON(pkg.Links, baseURL, "url"); err != nil {
				return err
			}
		}

		for _, bndl := range doc.ConsumptionBundles {
			if bndl.Links, err = rewriteRelativeURIsInJSON(bndl.Links, baseURL, "url"); err != nil {
				return err
			}
			if bndl.CredentialExchangeStrategies, err = rewriteRelativeURIsInJSON(bndl.CredentialExchangeStrategies, baseURL, "callbackUrl"); err != nil {
				return err
			}
		}

		for _, api := range doc.APIResources {
			for _, definition := range api.ResourceDefinitions {
				if !isAbsoluteURL(definition.URL) {
					definition.URL = baseURL + definition.URL
				}
			}
			if api.APIResourceLinks, err = rewriteRelativeURIsInJSON(api.APIResourceLinks, baseURL, "url"); err != nil {
				return err
			}
			if api.Links, err = rewriteRelativeURIsInJSON(api.Links, baseURL, "url"); err != nil {
				return err
			}
			if api.ChangeLogEntries, err = rewriteRelativeURIsInJSON(api.ChangeLogEntries, baseURL, "url"); err != nil {
				return err
			}
			if api.TargetURLs, err = rewriteRelativeURIsInJSONArray(api.TargetURLs, baseURL); err != nil {
				return err
			}
			rewriteDefaultTargetURL(api.PartOfConsumptionBundles, baseURL)
		}

		for _, event := range doc.EventResources {
			if event.ChangeLogEntries, err = rewriteRelativeURIsInJSON(event.ChangeLogEntries, baseURL, "url"); err != nil {
				return err
			}
			if event.Links, err = rewriteRelativeURIsInJSON(event.Links, baseURL, "url"); err != nil {
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
	packages := make(map[string]*model.PackageInput)
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
