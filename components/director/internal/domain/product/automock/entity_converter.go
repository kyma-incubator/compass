// Code generated by mockery. DO NOT EDIT.

package automock

import (
	product "github.com/kyma-incubator/compass/components/director/internal/domain/product"
	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// EntityConverter is an autogenerated mock type for the EntityConverter type
type EntityConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: entity
func (_m *EntityConverter) FromEntity(entity *product.Entity) (*model.Product, error) {
	ret := _m.Called(entity)

	var r0 *model.Product
	var r1 error
	if rf, ok := ret.Get(0).(func(*product.Entity) (*model.Product, error)); ok {
		return rf(entity)
	}
	if rf, ok := ret.Get(0).(func(*product.Entity) *model.Product); ok {
		r0 = rf(entity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Product)
		}
	}

	if rf, ok := ret.Get(1).(func(*product.Entity) error); ok {
		r1 = rf(entity)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToEntity provides a mock function with given fields: in
func (_m *EntityConverter) ToEntity(in *model.Product) *product.Entity {
	ret := _m.Called(in)

	var r0 *product.Entity
	if rf, ok := ret.Get(0).(func(*model.Product) *product.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*product.Entity)
		}
	}

	return r0
}

// NewEntityConverter creates a new instance of EntityConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEntityConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *EntityConverter {
	mock := &EntityConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
