// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// EventDefService is an autogenerated mock type for the EventDefService type
type EventDefService struct {
	mock.Mock
}

// CreateInBundle provides a mock function with given fields: ctx, resourceType, resourceID, bundleID, in, spec
func (_m *EventDefService) CreateInBundle(ctx context.Context, resourceType resource.Type, resourceID string, bundleID string, in model.EventDefinitionInput, spec *model.SpecInput) (string, error) {
	ret := _m.Called(ctx, resourceType, resourceID, bundleID, in, spec)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, string, model.EventDefinitionInput, *model.SpecInput) string); ok {
		r0 = rf(ctx, resourceType, resourceID, bundleID, in, spec)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, string, model.EventDefinitionInput, *model.SpecInput) error); ok {
		r1 = rf(ctx, resourceType, resourceID, bundleID, in, spec)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *EventDefService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *EventDefService) Get(ctx context.Context, id string) (*model.EventDefinition, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.EventDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.EventDefinition); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EventDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListFetchRequests provides a mock function with given fields: ctx, eventDefIDs
func (_m *EventDefService) ListFetchRequests(ctx context.Context, eventDefIDs []string) ([]*model.FetchRequest, error) {
	ret := _m.Called(ctx, eventDefIDs)

	var r0 []*model.FetchRequest
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*model.FetchRequest); ok {
		r0 = rf(ctx, eventDefIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FetchRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, eventDefIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in, spec
func (_m *EventDefService) Update(ctx context.Context, id string, in model.EventDefinitionInput, spec *model.SpecInput) error {
	ret := _m.Called(ctx, id, in, spec)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.EventDefinitionInput, *model.SpecInput) error); ok {
		r0 = rf(ctx, id, in, spec)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewEventDefService interface {
	mock.TestingT
	Cleanup(func())
}

// NewEventDefService creates a new instance of EventDefService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEventDefService(t mockConstructorTestingTNewEventDefService) *EventDefService {
	mock := &EventDefService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
