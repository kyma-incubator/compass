// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// RuntimeRepo is an autogenerated mock type for the runtimeRepo type
type RuntimeRepo struct {
	mock.Mock
}

// ListByIDs provides a mock function with given fields: ctx, tenant, ids
func (_m *RuntimeRepo) ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.Runtime, error) {
	ret := _m.Called(ctx, tenant, ids)

	var r0 []*model.Runtime
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) ([]*model.Runtime, error)); ok {
		return rf(ctx, tenant, ids)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) []*model.Runtime); ok {
		r0 = rf(ctx, tenant, ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Runtime)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tenant, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRuntimeRepo creates a new instance of RuntimeRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRuntimeRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *RuntimeRepo {
	mock := &RuntimeRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
