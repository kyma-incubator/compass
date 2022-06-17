// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// RuntimeContextService is an autogenerated mock type for the RuntimeContextService type
type RuntimeContextService struct {
	mock.Mock
}

// Delete provides a mock function with given fields: ctx, id
func (_m *RuntimeContextService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetForRuntime provides a mock function with given fields: ctx, id, runtimeID
func (_m *RuntimeContextService) GetForRuntime(ctx context.Context, id string, runtimeID string) (*model.RuntimeContext, error) {
	ret := _m.Called(ctx, id, runtimeID)

	var r0 *model.RuntimeContext
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.RuntimeContext); ok {
		r0 = rf(ctx, id, runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContext)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, runtimeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllForRuntime provides a mock function with given fields: ctx, runtimeID
func (_m *RuntimeContextService) ListAllForRuntime(ctx context.Context, runtimeID string) ([]*model.RuntimeContext, error) {
	ret := _m.Called(ctx, runtimeID)

	var r0 []*model.RuntimeContext
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.RuntimeContext); ok {
		r0 = rf(ctx, runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.RuntimeContext)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, runtimeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByRuntimeIDs provides a mock function with given fields: ctx, runtimeIDs, pageSize, cursor
func (_m *RuntimeContextService) ListByRuntimeIDs(ctx context.Context, runtimeIDs []string, pageSize int, cursor string) ([]*model.RuntimeContextPage, error) {
	ret := _m.Called(ctx, runtimeIDs, pageSize, cursor)

	var r0 []*model.RuntimeContextPage
	if rf, ok := ret.Get(0).(func(context.Context, []string, int, string) []*model.RuntimeContextPage); ok {
		r0 = rf(ctx, runtimeIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.RuntimeContextPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, int, string) error); ok {
		r1 = rf(ctx, runtimeIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRuntimeContextService creates a new instance of RuntimeContextService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeContextService(t testing.TB) *RuntimeContextService {
	mock := &RuntimeContextService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
