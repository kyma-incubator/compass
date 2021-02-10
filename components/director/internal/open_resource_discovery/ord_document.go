package open_resource_discovery

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
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
	URL                 string           `json:"url"`
	SystemInstanceAware bool             `json:"systemInstanceAware"`
	AccessStrategies    AccessStrategies `json:"accessStrategies"`
}

type Document struct {
	Schema                string `json:"$schema"`
	OpenResourceDiscovery string `json:"openResourceDiscovery"`
	Description           string `json:"description"`

	// TODO: In the current state of ORD we assume that the describedSystemInstance = providerSystemInstance due to the missing id correlation in ORD. Thus we work only with the DescribedSystemInstance.
	DescribedSystemInstance *model.Application `json:"describedSystemInstance"`
	ProviderSystemInstance  *model.Application `json:"providerSystemInstance"`

	Packages           []model.PackageInput         `json:"packages"`
	ConsumptionBundles []model.BundleCreateInput    `json:"consumptionBundles"`
	Products           []model.ProductInput         `json:"products"`
	APIResources       []model.APIDefinitionInput   `json:"apiResources"`
	EventResources     []model.EventDefinitionInput `json:"eventResources"`
	Tombstones         []model.TombstoneInput       `json:"tombstones"`
	Vendors            []model.VendorInput          `json:"vendors"`
}

type Documents []*Document

func (docs Documents) Validate() error {
	packageIDs := make(map[string]bool, 0)
	bundleIDs := make(map[string]bool, 0)
	productIDs := make(map[string]bool, 0)
	apiIDs := make(map[string]bool, 0)
	eventIDs := make(map[string]bool, 0)
	vendorIDs := make(map[string]bool, 0)

	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			packageIDs[pkg.OrdID] = true
		}
		for _, bndl := range doc.ConsumptionBundles {
			bundleIDs[*bndl.OrdID] = true
		}
		for _, product := range doc.Products {
			productIDs[product.OrdID] = true
		}
		for _, api := range doc.APIResources {
			apiIDs[*api.OrdID] = true
		}
		for _, event := range doc.EventResources {
			eventIDs[*event.OrdID] = true
		}
		for _, vendor := range doc.Vendors {
			vendorIDs[vendor.OrdID] = true
		}
	}

	// Validate entity relations
	for _, doc := range docs {
		for _, pkg := range doc.Packages {
			if !vendorIDs[*pkg.Vendor] {
				return errors.Errorf("package with id %q has a reference to unknown vendor %q", pkg.OrdID, *pkg.Vendor)
			}
			ordIDs := gjson.ParseBytes(pkg.PartOfProducts).Array() // TODO: Should be validated that partOfProducts is an array of strings (ordIDs)
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
				return errors.Errorf("api with id %q has a reference to unknown package %q", api.OrdID, *api.OrdPackageID)
			}
			if api.OrdBundleID != nil && !bundleIDs[*api.OrdBundleID] {
				return errors.Errorf("api with id %q has a reference to unknown bundle %q", api.OrdID, *api.OrdBundleID)
			}
			ordIDs := gjson.ParseBytes(api.PartOfProducts).Array() // TODO: Should be validated that partOfProducts is an array of strings (ordIDs)
			for _, productID := range ordIDs {
				if !productIDs[productID.String()] {
					return errors.Errorf("api with id %q has a reference to unknown product %q", api.OrdID, productID.String())
				}
			}
		}
		for _, event := range doc.EventResources {
			if !packageIDs[*event.OrdPackageID] {
				return errors.Errorf("event with id %q has a reference to unknown package %q", event.OrdID, *event.OrdPackageID)
			}
			if event.OrdBundleID != nil && !bundleIDs[*event.OrdBundleID] {
				return errors.Errorf("event with id %q has a reference to unknown bundle %q", event.OrdID, *event.OrdBundleID)
			}
			ordIDs := gjson.ParseBytes(event.PartOfProducts).Array() // TODO: Should be validated that partOfProducts is an array of strings (ordIDs)
			for _, productID := range ordIDs {
				if !productIDs[productID.String()] {
					return errors.Errorf("event with id %q has a reference to unknown product %q", event.OrdID, productID.String())
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

	return nil
}

// Sanitize performs all the merging and rewriting rules defined in ORD
//  - Rewrite all relative URIs using the baseURL from the System Instance
//  - Merging of products, labels etc.
func (docs Documents) Sanitize() error {
	return nil
}
