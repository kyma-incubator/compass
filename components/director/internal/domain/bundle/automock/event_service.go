// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// EventService is an autogenerated mock type for the EventService type
type EventService struct {
	mock.Mock
}

// CreateInBundle provides a mock function with given fields: ctx, appID, bundleID, in, spec
func (_m *EventService) CreateInBundle(ctx context.Context, appID string, bundleID string, in model.EventDefinitionInput, spec *model.SpecInput) (string, error) {
	ret := _m.Called(ctx, appID, bundleID, in, spec)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.EventDefinitionInput, *model.SpecInput) string); ok {
		r0 = rf(ctx, appID, bundleID, in, spec)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.EventDefinitionInput, *model.SpecInput) error); ok {
		r1 = rf(ctx, appID, bundleID, in, spec)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteAllByBundleID provides a mock function with given fields: ctx, bundleID
func (_m *EventService) DeleteAllByBundleID(ctx context.Context, bundleID string) error {
	ret := _m.Called(ctx, bundleID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, bundleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetForBundle provides a mock function with given fields: ctx, id, bundleID
func (_m *EventService) GetForBundle(ctx context.Context, id string, bundleID string) (*model.EventDefinition, error) {
	ret := _m.Called(ctx, id, bundleID)

	var r0 *model.EventDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.EventDefinition); ok {
		r0 = rf(ctx, id, bundleID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EventDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, bundleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByBundleIDs provides a mock function with given fields: ctx, bundleIDs, pageSize, cursor
func (_m *EventService) ListByBundleIDs(ctx context.Context, bundleIDs []string, pageSize int, cursor string) ([]*model.EventDefinitionPage, error) {
	ret := _m.Called(ctx, bundleIDs, pageSize, cursor)

	var r0 []*model.EventDefinitionPage
	if rf, ok := ret.Get(0).(func(context.Context, []string, int, string) []*model.EventDefinitionPage); ok {
		r0 = rf(ctx, bundleIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.EventDefinitionPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, int, string) error); ok {
		r1 = rf(ctx, bundleIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewEventService interface {
	mock.TestingT
	Cleanup(func())
}

// NewEventService creates a new instance of EventService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEventService(t mockConstructorTestingTNewEventService) *EventService {
	mock := &EventService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
