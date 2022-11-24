// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// GlobalVendorService is an autogenerated mock type for the GlobalVendorService type
type GlobalVendorService struct {
	mock.Mock
}

// CreateGlobal provides a mock function with given fields: ctx, in
func (_m *GlobalVendorService) CreateGlobal(ctx context.Context, in model.VendorInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.VendorInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.VendorInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteGlobal provides a mock function with given fields: ctx, id
func (_m *GlobalVendorService) DeleteGlobal(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListGlobal provides a mock function with given fields: ctx
func (_m *GlobalVendorService) ListGlobal(ctx context.Context) ([]*model.Vendor, error) {
	ret := _m.Called(ctx)

	var r0 []*model.Vendor
	if rf, ok := ret.Get(0).(func(context.Context) []*model.Vendor); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Vendor)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateGlobal provides a mock function with given fields: ctx, id, in
func (_m *GlobalVendorService) UpdateGlobal(ctx context.Context, id string, in model.VendorInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.VendorInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewGlobalVendorService interface {
	mock.TestingT
	Cleanup(func())
}

// NewGlobalVendorService creates a new instance of GlobalVendorService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewGlobalVendorService(t mockConstructorTestingTNewGlobalVendorService) *GlobalVendorService {
	mock := &GlobalVendorService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
