package tests

import (
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
)

// WellKnownEndpoint is the single entry point for the discovery.
const WellKnownEndpoint = "/.well-known/open-resource-discovery"

// WellKnownConfig represents the whole config object
type WellKnownConfig struct {
	BaseURL                 string                  `json:"baseUrl"`
	OpenResourceDiscoveryV1 OpenResourceDiscoveryV1 `json:"openResourceDiscoveryV1"`
}

// OpenResourceDiscoveryV1 contains all Documents' details
type OpenResourceDiscoveryV1 struct {
	Documents []DocumentDetails `json:"documents"`
}

// DocumentDetails contains fields related to the fetching of each Document
type DocumentDetails struct {
	URL              string                          `json:"url"`
	AccessStrategies accessstrategy.AccessStrategies `json:"accessStrategies"`
	Perspective      string                          `json:"perspective"`
}
