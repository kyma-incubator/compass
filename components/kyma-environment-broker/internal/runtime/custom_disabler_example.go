// This package is NOT FOR PRODUCTION USE CASE. EXAMPLE ONLY.
package runtime

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

// CustomDisablerComponentName defines the component name
const CustomDisablerComponentName = "component-x"

// CustomDisablerExample provides example how to add functionality for disabling some
// which requires more complex logic
type CustomDisablerExample struct{}

// NewCustomDisablerExample returns new instance of CustomDisablerExample
func NewCustomDisablerExample() *CustomDisablerExample {
	return &CustomDisablerExample{}
}

// Disable disables given component by adding new overrides to existing module
func (CustomDisablerExample) Disable(components internal.ComponentConfigurationInputList) internal.ComponentConfigurationInputList {
	disableOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key:   "component-x.enabled",
			Value: "false",
		},
		{
			Key:   "component-x.Output.conf.enabled",
			Value: "false",
		},
	}

	for _, c := range components {
		if c.Component == CustomDisablerComponentName {
			c.Configuration = append(c.Configuration, disableOverrides...)
		}
	}

	return components
}
