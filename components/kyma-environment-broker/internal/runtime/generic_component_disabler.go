package runtime

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

// GenericComponentDisabler provides functionality for removing configured component from given list
type GenericComponentDisabler struct {
	componentName      string
	componentNamespace string
}

// NewGenericComponentDisabler returns new instance of GenericComponentDisabler
func NewGenericComponentDisabler(name string, namespace string) *GenericComponentDisabler {
	return &GenericComponentDisabler{componentName: name, componentNamespace: namespace}
}

// Disable removes component form given lists. Filtering without allocating.
//
// source: https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
func (g *GenericComponentDisabler) Disable(components internal.ComponentConfigurationInputList) internal.ComponentConfigurationInputList {
	filterOut := components[:0]
	for _, component := range components {
		if !g.shouldRemove(component) {
			filterOut = append(filterOut, component)
		}
	}

	for i := len(filterOut); i < len(components); i++ {
		components[i] = nil
	}

	return filterOut
}

func (g *GenericComponentDisabler) shouldRemove(in *gqlschema.ComponentConfigurationInput) bool {
	if in == nil {
		return false
	}
	return in.Component == g.componentName && in.Namespace == g.componentNamespace
}
