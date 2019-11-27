// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// EventAPIService is an autogenerated mock type for the EventAPIService type
type EventAPIService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, applicationID, in
func (_m *EventAPIService) Create(ctx context.Context, applicationID string, in model.EventAPIDefinitionInput) (string, error) {
	ret := _m.Called(ctx, applicationID, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, model.EventAPIDefinitionInput) string); ok {
		r0 = rf(ctx, applicationID, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.EventAPIDefinitionInput) error); ok {
		r1 = rf(ctx, applicationID, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *EventAPIService) Delete(ctx context.Context, id string) error {
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
func (_m *EventAPIService) Get(ctx context.Context, id string) (*model.EventAPIDefinition, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.EventAPIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.EventAPIDefinition); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EventAPIDefinition)
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

// GetForApplication provides a mock function with given fields: ctx, id, applicationID
func (_m *EventAPIService) GetForApplication(ctx context.Context, id string, applicationID string) (*model.EventAPIDefinition, error) {
	ret := _m.Called(ctx, id, applicationID)

	var r0 *model.EventAPIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.EventAPIDefinition); ok {
		r0 = rf(ctx, id, applicationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EventAPIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, applicationID, pageSize, cursor
func (_m *EventAPIService) List(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.EventAPIDefinitionPage, error) {
	ret := _m.Called(ctx, applicationID, pageSize, cursor)

	var r0 *model.EventAPIDefinitionPage
	if rf, ok := ret.Get(0).(func(context.Context, string, int, string) *model.EventAPIDefinitionPage); ok {
		r0 = rf(ctx, applicationID, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EventAPIDefinitionPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, int, string) error); ok {
		r1 = rf(ctx, applicationID, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *EventAPIService) Update(ctx context.Context, id string, in model.EventAPIDefinitionInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.EventAPIDefinitionInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
