// Code generated by mockery v2.5.1. DO NOT EDIT.

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

// GetForBundle provides a mock function with given fields: ctx, id, bundleID
func (_m *BundleInstanceAuthService) GetForBundle(ctx context.Context, id string, bundleID string) (*model.BundleInstanceAuth, error) {
	ret := _m.Called(ctx, id, bundleID)

	var r0 *model.BundleInstanceAuth
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.BundleInstanceAuth); ok {
		r0 = rf(ctx, id, bundleID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BundleInstanceAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, bundleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, id
func (_m *BundleInstanceAuthService) List(ctx context.Context, id string) ([]*model.BundleInstanceAuth, error) {
	ret := _m.Called(ctx, id)

	var r0 []*model.BundleInstanceAuth
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.BundleInstanceAuth); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BundleInstanceAuth)
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
