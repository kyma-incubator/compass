// Code generated by mockery. DO NOT EDIT.

package automock

import (
	destination "github.com/kyma-incubator/compass/components/director/internal/domain/destination"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// EntityConverter is an autogenerated mock type for the EntityConverter type
type EntityConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: entity
func (_m *EntityConverter) FromEntity(entity *destination.Entity) *model.Destination {
	ret := _m.Called(entity)

	var r0 *model.Destination
	if rf, ok := ret.Get(0).(func(*destination.Entity) *model.Destination); ok {
		r0 = rf(entity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Destination)
		}
	}

	return r0
}

// ToEntity provides a mock function with given fields: in
func (_m *EntityConverter) ToEntity(in *model.Destination) *destination.Entity {
	ret := _m.Called(in)

	var r0 *destination.Entity
	if rf, ok := ret.Get(0).(func(*model.Destination) *destination.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*destination.Entity)
		}
	}

	return r0
}

type mockConstructorTestingTNewEntityConverter interface {
	mock.TestingT
	Cleanup(func())
}

// NewEntityConverter creates a new instance of EntityConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEntityConverter(t mockConstructorTestingTNewEntityConverter) *EntityConverter {
	mock := &EntityConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
