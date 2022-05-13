// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// SystemsService is an autogenerated mock type for the systemsService type
type SystemsService struct {
	mock.Mock
}

// GetByNameAndSystemNumber provides a mock function with given fields: ctx, name, systemNumber
func (_m *SystemsService) GetByNameAndSystemNumber(ctx context.Context, name string, systemNumber string) (*model.Application, error) {
	ret := _m.Called(ctx, name, systemNumber)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Application); ok {
		r0 = rf(ctx, name, systemNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, name, systemNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TrustedUpsert provides a mock function with given fields: ctx, in
func (_m *SystemsService) TrustedUpsert(ctx context.Context, in model.ApplicationRegisterInput) error {
	ret := _m.Called(ctx, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.ApplicationRegisterInput) error); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TrustedUpsertFromTemplate provides a mock function with given fields: ctx, in, appTemplateID
func (_m *SystemsService) TrustedUpsertFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) error {
	ret := _m.Called(ctx, in, appTemplateID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.ApplicationRegisterInput, *string) error); ok {
		r0 = rf(ctx, in, appTemplateID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSystemsService creates a new instance of SystemsService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewSystemsService(t testing.TB) *SystemsService {
	mock := &SystemsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
