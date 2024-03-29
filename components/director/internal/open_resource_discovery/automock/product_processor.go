// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// ProductProcessor is an autogenerated mock type for the ProductProcessor type
type ProductProcessor struct {
	mock.Mock
}

// Process provides a mock function with given fields: ctx, resourceType, resourceID, products
func (_m *ProductProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, products []*model.ProductInput) ([]*model.Product, error) {
	ret := _m.Called(ctx, resourceType, resourceID, products)

	var r0 []*model.Product
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, []*model.ProductInput) ([]*model.Product, error)); ok {
		return rf(ctx, resourceType, resourceID, products)
	}
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, []*model.ProductInput) []*model.Product); ok {
		r0 = rf(ctx, resourceType, resourceID, products)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Product)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, []*model.ProductInput) error); ok {
		r1 = rf(ctx, resourceType, resourceID, products)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewProductProcessor creates a new instance of ProductProcessor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProductProcessor(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProductProcessor {
	mock := &ProductProcessor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
