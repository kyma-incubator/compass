// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// RuntimeService is an autogenerated mock type for the RuntimeService type
type RuntimeService struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, id
func (_m *RuntimeService) Get(ctx context.Context, id string) (*model.Runtime, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Runtime); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Runtime)
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

type mockConstructorTestingTNewRuntimeService interface {
	mock.TestingT
	Cleanup(func())
}

// NewRuntimeService creates a new instance of RuntimeService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeService(t mockConstructorTestingTNewRuntimeService) *RuntimeService {
	mock := &RuntimeService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
