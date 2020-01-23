package runtime

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

// OptionalComponentDisabler disables component form the given list and returns modified list
type OptionalComponentDisabler interface {
	Disable(components internal.ComponentConfigurationInputList) internal.ComponentConfigurationInputList
}

// ComponentsDisablers represents type for defining components disabler list
type ComponentsDisablers map[string]OptionalComponentDisabler

// OptionalComponentsService provides functionality for executing component disablers
type OptionalComponentsService struct {
	registered map[string]OptionalComponentDisabler
}

// NewOptionalComponentsService returns new instance of ResourceSupervisorAggregator
func NewOptionalComponentsService(initialList ComponentsDisablers) *OptionalComponentsService {
	return &OptionalComponentsService{
		registered: initialList,
	}
}

// GetAllOptionalComponentsNames returns list of registered components disablers names
func (f *OptionalComponentsService) GetAllOptionalComponentsNames() []string {
	var names []string
	for name := range f.registered {
		names = append(names, name)
	}

	return names
}

// ExecuteDisablers executes disablers on given input and returns modified list.
//
// BE AWARE: in current implementation the input is also modified.
func (f *OptionalComponentsService) ExecuteDisablers(components internal.ComponentConfigurationInputList, names ...string) (internal.ComponentConfigurationInputList, error) {
	var filterOut = components
	for _, name := range names {
		concreteDisabler, exists := f.registered[name]
		if !exists {
			return nil, fmt.Errorf("disabler for component %s was not found", name)
		}

		filterOut = concreteDisabler.Disable(filterOut)
	}

	return filterOut, nil
}

// ComputeComponentsToDisable returns disabler names that needs to be executed
func (f *OptionalComponentsService) ComputeComponentsToDisable(optComponentsToKeep []string) []string {
	var (
		allOptComponents       = f.GetAllOptionalComponentsNames()
		optComponentsToInstall = toMap(optComponentsToKeep)
		optComponentsToDisable []string
	)

	for _, name := range allOptComponents {
		if _, found := optComponentsToInstall[name]; found {
			continue
		}
		optComponentsToDisable = append(optComponentsToDisable, name)
	}
	return optComponentsToDisable
}

func toMap(in []string) map[string]struct{} {
	out := map[string]struct{}{}

	for _, entry := range in {
		out[entry] = struct{}{}
	}

	return out
}
