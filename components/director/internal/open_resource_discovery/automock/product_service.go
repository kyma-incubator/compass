// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// ProductService is an autogenerated mock type for the ProductService type
type ProductService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, resourceType, resourceID, in
func (_m *ProductService) Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.ProductInput) (string, error) {
	ret := _m.Called(ctx, resourceType, resourceID, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, model.ProductInput) string); ok {
		r0 = rf(ctx, resourceType, resourceID, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, model.ProductInput) error); ok {
		r1 = rf(ctx, resourceType, resourceID, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *ProductService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListByApplicationID provides a mock function with given fields: ctx, appID
func (_m *ProductService) ListByApplicationID(ctx context.Context, appID string) ([]*model.Product, error) {
	ret := _m.Called(ctx, appID)

	var r0 []*model.Product
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Product); ok {
		r0 = rf(ctx, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Product)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByApplicationTemplateVersionID provides a mock function with given fields: ctx, appID
func (_m *ProductService) ListByApplicationTemplateVersionID(ctx context.Context, appID string) ([]*model.Product, error) {
	ret := _m.Called(ctx, appID)

	var r0 []*model.Product
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Product); ok {
		r0 = rf(ctx, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Product)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, resourceType, id, in
func (_m *ProductService) Update(ctx context.Context, resourceType resource.Type, id string, in model.ProductInput) error {
	ret := _m.Called(ctx, resourceType, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, model.ProductInput) error); ok {
		r0 = rf(ctx, resourceType, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewProductService interface {
	mock.TestingT
	Cleanup(func())
}

// NewProductService creates a new instance of ProductService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProductService(t mockConstructorTestingTNewProductService) *ProductService {
	mock := &ProductService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
