// Code generated by mockery v2.5.1. DO NOT EDIT.

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

// CreateInBundle provides a mock function with given fields: ctx, appID, bundleID, in, spec, eventHash
func (_m *EventService) CreateInBundle(ctx context.Context, appID string, bundleID string, in model.EventDefinitionInput, spec *model.SpecInput, eventHash uint64) (string, error) {
	ret := _m.Called(ctx, appID, bundleID, in, spec, eventHash)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.EventDefinitionInput, *model.SpecInput, uint64) string); ok {
		r0 = rf(ctx, appID, bundleID, in, spec, eventHash)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.EventDefinitionInput, *model.SpecInput, uint64) error); ok {
		r1 = rf(ctx, appID, bundleID, in, spec, eventHash)
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

// ListForBundle provides a mock function with given fields: ctx, bundleID, pageSize, cursor
func (_m *EventService) ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	ret := _m.Called(ctx, bundleID, pageSize, cursor)

	var r0 *model.EventDefinitionPage
	if rf, ok := ret.Get(0).(func(context.Context, string, int, string) *model.EventDefinitionPage); ok {
		r0 = rf(ctx, bundleID, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EventDefinitionPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, int, string) error); ok {
		r1 = rf(ctx, bundleID, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
