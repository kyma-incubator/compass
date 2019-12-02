// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"

import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"
import uuid "github.com/google/uuid"

// EventingService is an autogenerated mock type for the EventingService type
type EventingService struct {
	mock.Mock
}

// DeleteDefaultForApplication provides a mock function with given fields: ctx, appID
func (_m *EventingService) DeleteDefaultForApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error) {
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

// SetAsDefaultForApplication provides a mock function with given fields: ctx, runtimeID, appID
func (_m *EventingService) SetAsDefaultForApplication(ctx context.Context, runtimeID uuid.UUID, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error) {
	ret := _m.Called(ctx, runtimeID, appID)

	var r0 *model.ApplicationEventingConfiguration
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) *model.ApplicationEventingConfiguration); ok {
		r0 = rf(ctx, runtimeID, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationEventingConfiguration)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, uuid.UUID) error); ok {
		r1 = rf(ctx, runtimeID, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
