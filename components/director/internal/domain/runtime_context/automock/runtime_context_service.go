// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// RuntimeContextService is an autogenerated mock type for the RuntimeContextService type
type RuntimeContextService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *RuntimeContextService) Create(ctx context.Context, in model.RuntimeContextInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.RuntimeContextInput) (string, error)); ok {
		return rf(ctx, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.RuntimeContextInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.RuntimeContextInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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

// GetByID provides a mock function with given fields: ctx, id
func (_m *RuntimeContextService) GetByID(ctx context.Context, id string) (*model.RuntimeContext, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.RuntimeContext
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.RuntimeContext, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.RuntimeContext); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContext)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListLabels provides a mock function with given fields: ctx, runtimeID
func (_m *RuntimeContextService) ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error) {
	ret := _m.Called(ctx, runtimeID)

	var r0 map[string]*model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (map[string]*model.Label, error)); ok {
		return rf(ctx, runtimeID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) map[string]*model.Label); ok {
		r0 = rf(ctx, runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, runtimeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *RuntimeContextService) Update(ctx context.Context, id string, in model.RuntimeContextInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.RuntimeContextInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewRuntimeContextService creates a new instance of RuntimeContextService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRuntimeContextService(t interface {
	mock.TestingT
	Cleanup(func())
}) *RuntimeContextService {
	mock := &RuntimeContextService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
