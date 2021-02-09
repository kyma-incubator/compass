package open_resource_discovery

import "github.com/kyma-incubator/compass/components/director/internal/model"

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

	// TODO: In the current state of ORD we assume that the describedSystemInstance = providerSystemInstance.
	DescribedSystemInstance *model.Application `json:"describedSystemInstance"`
	ProviderSystemInstance  *model.Application `json:"providerSystemInstance"`

	Packages           []*model.PackageInput         `json:"package"`
	ConsumptionBundles []*model.BundleCreateInput    `json:"consumptionBundles"`
	Products           []*model.ProductInput         `json:"products"`
	APIResources       []*model.APIDefinitionInput   `json:"apiResources"`
	EventResources     []*model.EventDefinitionInput `json:"eventResources"`
	Tombstones         []*model.TombstoneInput       `json:"tombstones"`
	Vendors            []*model.VendorInput          `json:"vendors"`
}

type Documents []*Document

func (*Documents) Validate() error {
	return nil
}
