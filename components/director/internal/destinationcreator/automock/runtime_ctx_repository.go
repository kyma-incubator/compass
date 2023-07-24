// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// RuntimeCtxRepository is an autogenerated mock type for the runtimeCtxRepository type
type RuntimeCtxRepository struct {
	mock.Mock
}

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *RuntimeCtxRepository) GetByID(ctx context.Context, tenant string, id string) (*model.RuntimeContext, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 *model.RuntimeContext
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.RuntimeContext); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContext)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewRuntimeCtxRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewRuntimeCtxRepository creates a new instance of RuntimeCtxRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeCtxRepository(t mockConstructorTestingTNewRuntimeCtxRepository) *RuntimeCtxRepository {
	mock := &RuntimeCtxRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}