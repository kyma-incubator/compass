// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// SystemsService is an autogenerated mock type for the SystemsService type
type SystemsService struct {
	mock.Mock
}

// CreateManyIfNotExistsWithEventualTemplate provides a mock function with given fields: ctx, applicationInputs
func (_m *SystemsService) CreateManyIfNotExistsWithEventualTemplate(ctx context.Context, applicationInputs []model.ApplicationRegisterInputWithTemplate) error {
	ret := _m.Called(ctx, applicationInputs)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []model.ApplicationRegisterInputWithTemplate) error); ok {
		r0 = rf(ctx, applicationInputs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
