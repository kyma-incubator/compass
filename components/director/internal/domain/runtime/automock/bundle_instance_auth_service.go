// Code generated by mockery v2.10.5. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// BundleInstanceAuthService is an autogenerated mock type for the BundleInstanceAuthService type
type BundleInstanceAuthService struct {
	mock.Mock
}

// ListByRuntimeID provides a mock function with given fields: ctx, runtimeID
func (_m *BundleInstanceAuthService) ListByRuntimeID(ctx context.Context, runtimeID string) ([]*model.BundleInstanceAuth, error) {
	ret := _m.Called(ctx, runtimeID)

	var r0 []*model.BundleInstanceAuth
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.BundleInstanceAuth); ok {
		r0 = rf(ctx, runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BundleInstanceAuth)
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

// Update provides a mock function with given fields: ctx, instanceAuth
func (_m *BundleInstanceAuthService) Update(ctx context.Context, instanceAuth *model.BundleInstanceAuth) error {
	ret := _m.Called(ctx, instanceAuth)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.BundleInstanceAuth) error); ok {
		r0 = rf(ctx, instanceAuth)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
