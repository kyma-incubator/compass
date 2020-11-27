// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// EventingService is an autogenerated mock type for the EventingService type
type EventingService struct {
	mock.Mock
}

// CleanupAfterUnregisteringApplication provides a mock function with given fields: ctx, appID
func (_m *EventingService) CleanupAfterUnregisteringApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error) {
	ret := _m.Called(ctx, appID)

	var r0 *model.ApplicationEventingConfiguration
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *model.ApplicationEventingConfiguration); ok {
		r0 = rf(ctx, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationEventingConfiguration)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetForApplication provides a mock function with given fields: ctx, app
func (_m *EventingService) GetForApplication(ctx context.Context, app model.Application) (*model.ApplicationEventingConfiguration, error) {
	ret := _m.Called(ctx, app)

	var r0 *model.ApplicationEventingConfiguration
	if rf, ok := ret.Get(0).(func(context.Context, model.Application) *model.ApplicationEventingConfiguration); ok {
		r0 = rf(ctx, app)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationEventingConfiguration)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.Application) error); ok {
		r1 = rf(ctx, app)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
