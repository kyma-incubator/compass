// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// runtimeSelfRegisterManager is an autogenerated mock type for the runtimeSelfRegisterManager type
type runtimeSelfRegisterManager struct {
	mock.Mock
}

// CleanupSelfRegisteredRuntime provides a mock function with given fields: ctx, selfRegisterLabelValue
func (_m *runtimeSelfRegisterManager) CleanupSelfRegisteredRuntime(ctx context.Context, selfRegisterLabelValue string) error {
	ret := _m.Called(ctx, selfRegisterLabelValue)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, selfRegisterLabelValue)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetSelfRegDistinguishingLabelKey provides a mock function with given fields:
func (_m *runtimeSelfRegisterManager) GetSelfRegDistinguishingLabelKey() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// PrepareRuntimeForSelfRegistration provides a mock function with given fields: ctx, in
func (_m *runtimeSelfRegisterManager) PrepareRuntimeForSelfRegistration(ctx context.Context, in model.RuntimeInput) (model.RuntimeInput, error) {
	ret := _m.Called(ctx, in)

	var r0 model.RuntimeInput
	if rf, ok := ret.Get(0).(func(context.Context, model.RuntimeInput) model.RuntimeInput); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(model.RuntimeInput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.RuntimeInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
