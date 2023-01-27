// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// ApplicationService is an autogenerated mock type for the ApplicationService type
type ApplicationService struct {
	mock.Mock
}

// ListAll provides a mock function with given fields: ctx
func (_m *ApplicationService) ListAll(ctx context.Context) ([]*model.Application, error) {
	ret := _m.Called(ctx)

	var r0 []*model.Application
	if rf, ok := ret.Get(0).(func(context.Context) []*model.Application); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewApplicationService interface {
	mock.TestingT
	Cleanup(func())
}

// NewApplicationService creates a new instance of ApplicationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewApplicationService(t mockConstructorTestingTNewApplicationService) *ApplicationService {
	mock := &ApplicationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
