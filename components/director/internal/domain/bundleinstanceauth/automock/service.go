// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, bundleID, in, defaultAuth, requestInputSchema
func (_m *Service) Create(ctx context.Context, bundleID string, in model.BundleInstanceAuthRequestInput, defaultAuth *model.Auth, requestInputSchema *string) (string, error) {
	ret := _m.Called(ctx, bundleID, in, defaultAuth, requestInputSchema)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, model.BundleInstanceAuthRequestInput, *model.Auth, *string) string); ok {
		r0 = rf(ctx, bundleID, in, defaultAuth, requestInputSchema)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.BundleInstanceAuthRequestInput, *model.Auth, *string) error); ok {
		r1 = rf(ctx, bundleID, in, defaultAuth, requestInputSchema)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateBundleInstanceAuth provides a mock function with given fields: ctx, bundleID, in, requestInputSchema
func (_m *Service) CreateBundleInstanceAuth(ctx context.Context, bundleID string, in model.BundleInstanceAuthCreateInput, requestInputSchema *string) (string, error) {
	ret := _m.Called(ctx, bundleID, in, requestInputSchema)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, model.BundleInstanceAuthCreateInput, *string) string); ok {
		r0 = rf(ctx, bundleID, in, requestInputSchema)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.BundleInstanceAuthCreateInput, *string) error); ok {
		r1 = rf(ctx, bundleID, in, requestInputSchema)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *Service) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *Service) Get(ctx context.Context, id string) (*model.BundleInstanceAuth, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.BundleInstanceAuth
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.BundleInstanceAuth); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BundleInstanceAuth)
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

// RequestDeletion provides a mock function with given fields: ctx, instanceAuth, defaultBundleInstanceAuth
func (_m *Service) RequestDeletion(ctx context.Context, instanceAuth *model.BundleInstanceAuth, defaultBundleInstanceAuth *model.Auth) (bool, error) {
	ret := _m.Called(ctx, instanceAuth, defaultBundleInstanceAuth)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, *model.BundleInstanceAuth, *model.Auth) bool); ok {
		r0 = rf(ctx, instanceAuth, defaultBundleInstanceAuth)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.BundleInstanceAuth, *model.Auth) error); ok {
		r1 = rf(ctx, instanceAuth, defaultBundleInstanceAuth)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetAuth provides a mock function with given fields: ctx, id, in
func (_m *Service) SetAuth(ctx context.Context, id string, in model.BundleInstanceAuthSetInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.BundleInstanceAuthSetInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, instanceAuth
func (_m *Service) Update(ctx context.Context, instanceAuth *model.BundleInstanceAuth) error {
	ret := _m.Called(ctx, instanceAuth)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.BundleInstanceAuth) error); ok {
		r0 = rf(ctx, instanceAuth)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewService interface {
	mock.TestingT
	Cleanup(func())
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewService(t mockConstructorTestingTNewService) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
