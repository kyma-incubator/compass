// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// RuntimeContextRepository is an autogenerated mock type for the runtimeContextRepository type
type RuntimeContextRepository struct {
	mock.Mock
}

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *RuntimeContextRepository) GetByID(ctx context.Context, tenant string, id string) (*model.RuntimeContext, error) {
	ret := _m.Called(ctx, tenant, id)

	if len(ret) == 0 {
		panic("no return value specified for GetByID")
	}

	var r0 *model.RuntimeContext
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.RuntimeContext, error)); ok {
		return rf(ctx, tenant, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.RuntimeContext); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContext)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRuntimeContextRepository creates a new instance of RuntimeContextRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRuntimeContextRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *RuntimeContextRepository {
	mock := &RuntimeContextRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
