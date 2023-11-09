// Code generated by mockery. DO NOT EDIT.

package automock

import (
	entitytypemapping "github.com/kyma-incubator/compass/components/director/internal/domain/entitytypemapping"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// EntityTypeMappingConverter is an autogenerated mock type for the EntityTypeMappingConverter type
type EntityTypeMappingConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: entity
func (_m *EntityTypeMappingConverter) FromEntity(entity *entitytypemapping.Entity) *model.EntityTypeMapping {
	ret := _m.Called(entity)

	var r0 *model.EntityTypeMapping
	if rf, ok := ret.Get(0).(func(*entitytypemapping.Entity) *model.EntityTypeMapping); ok {
		r0 = rf(entity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EntityTypeMapping)
		}
	}

	return r0
}

// ToEntity provides a mock function with given fields: in
func (_m *EntityTypeMappingConverter) ToEntity(in *model.EntityTypeMapping) *entitytypemapping.Entity {
	ret := _m.Called(in)

	var r0 *entitytypemapping.Entity
	if rf, ok := ret.Get(0).(func(*model.EntityTypeMapping) *entitytypemapping.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entitytypemapping.Entity)
		}
	}

	return r0
}

// NewEntityTypeMappingConverter creates a new instance of EntityTypeMappingConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEntityTypeMappingConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *EntityTypeMappingConverter {
	mock := &EntityTypeMappingConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}