// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	processor "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// EventProcessor is an autogenerated mock type for the EventProcessor type
type EventProcessor struct {
	mock.Mock
}

// Process provides a mock function with given fields: ctx, resourceType, resourceID, bundlesFromDB, packagesFromDB, events, resourceHashes
func (_m *EventProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, events []*model.EventDefinitionInput, resourceHashes map[string]uint64) ([]*model.EventDefinition, []*processor.OrdFetchRequest, error) {
	ret := _m.Called(ctx, resourceType, resourceID, bundlesFromDB, packagesFromDB, events, resourceHashes)

	var r0 []*model.EventDefinition
	var r1 []*processor.OrdFetchRequest
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, []*model.Bundle, []*model.Package, []*model.EventDefinitionInput, map[string]uint64) ([]*model.EventDefinition, []*processor.OrdFetchRequest, error)); ok {
		return rf(ctx, resourceType, resourceID, bundlesFromDB, packagesFromDB, events, resourceHashes)
	}
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, []*model.Bundle, []*model.Package, []*model.EventDefinitionInput, map[string]uint64) []*model.EventDefinition); ok {
		r0 = rf(ctx, resourceType, resourceID, bundlesFromDB, packagesFromDB, events, resourceHashes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.EventDefinition)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, []*model.Bundle, []*model.Package, []*model.EventDefinitionInput, map[string]uint64) []*processor.OrdFetchRequest); ok {
		r1 = rf(ctx, resourceType, resourceID, bundlesFromDB, packagesFromDB, events, resourceHashes)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]*processor.OrdFetchRequest)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, resource.Type, string, []*model.Bundle, []*model.Package, []*model.EventDefinitionInput, map[string]uint64) error); ok {
		r2 = rf(ctx, resourceType, resourceID, bundlesFromDB, packagesFromDB, events, resourceHashes)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// NewEventProcessor creates a new instance of EventProcessor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEventProcessor(t interface {
	mock.TestingT
	Cleanup(func())
}) *EventProcessor {
	mock := &EventProcessor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}