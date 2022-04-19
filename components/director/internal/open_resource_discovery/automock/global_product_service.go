// Code generated by mockery v2.10.4. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// GlobalProductService is an autogenerated mock type for the GlobalProductService type
type GlobalProductService struct {
	mock.Mock
}

// CreateGlobal provides a mock function with given fields: ctx, in
func (_m *GlobalProductService) CreateGlobal(ctx context.Context, in model.ProductInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.ProductInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.ProductInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteGlobal provides a mock function with given fields: ctx, id
func (_m *GlobalProductService) DeleteGlobal(ctx context.Context, id string) error {
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
func (_m *GlobalProductService) ListGlobal(ctx context.Context) ([]*model.Product, error) {
	ret := _m.Called(ctx)

	var r0 []*model.Product
	if rf, ok := ret.Get(0).(func(context.Context) []*model.Product); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Product)
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
func (_m *GlobalProductService) UpdateGlobal(ctx context.Context, id string, in model.ProductInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.ProductInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
