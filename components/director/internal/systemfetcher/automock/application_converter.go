// Code generated by mockery v2.10.4. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// ApplicationConverter is an autogenerated mock type for the applicationConverter type
type ApplicationConverter struct {
	mock.Mock
}

// CreateInputJSONToModel provides a mock function with given fields: ctx, in
func (_m *ApplicationConverter) CreateInputJSONToModel(ctx context.Context, in string) (model.ApplicationRegisterInput, error) {
	ret := _m.Called(ctx, in)

	var r0 model.ApplicationRegisterInput
	if rf, ok := ret.Get(0).(func(context.Context, string) model.ApplicationRegisterInput); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(model.ApplicationRegisterInput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
