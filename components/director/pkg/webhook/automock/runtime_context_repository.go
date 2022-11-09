// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// RuntimeContextRepository is an autogenerated mock type for the runtimeContextRepository type
type RuntimeContextRepository struct {
	mock.Mock
}

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *RuntimeContextRepository) GetByID(ctx context.Context, tenant string, id string) (*model.RuntimeContext, error) {
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

// GetByRuntimeID provides a mock function with given fields: ctx, tenant, runtimeID
func (_m *RuntimeContextRepository) GetByRuntimeID(ctx context.Context, tenant string, runtimeID string) (*model.RuntimeContext, error) {
	ret := _m.Called(ctx, tenant, runtimeID)

	var r0 *model.RuntimeContext
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.RuntimeContext); ok {
		r0 = rf(ctx, tenant, runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContext)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, runtimeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRuntimeContextRepository creates a new instance of RuntimeContextRepository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeContextRepository(t testing.TB) *RuntimeContextRepository {
	mock := &RuntimeContextRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
