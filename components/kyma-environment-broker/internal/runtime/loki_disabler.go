package runtime

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

// LoggingComponentName defines the component name for logging functionality
const LoggingComponentName = "logging"

// LokiDisabler provides functionality for disabling the Loki component
type LokiDisabler struct{}

// NewLokiDisabler returns new instance of LokiDisabler
func NewLokiDisabler() *LokiDisabler {
	return &LokiDisabler{}
}

// Disable disables loki form given component lists by applying special overrides
// for logging component
func (LokiDisabler) Disable(components internal.ComponentConfigurationInputList) internal.ComponentConfigurationInputList {
	disableLokiOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key:   "loki.enabled",
			Value: "false",
		},
		// TODO: enable after adding LMS configuration
		//{
		//	Key:   "fluent-bit.conf.Output.loki.enabled",
		//	Value: "false",
		//},
	}

	for _, c := range components {
		if c.Component == LoggingComponentName {
			c.Configuration = append(c.Configuration, disableLokiOverrides...)
		}
	}

	return components
}
