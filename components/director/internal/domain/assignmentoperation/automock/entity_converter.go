// Code generated by mockery. DO NOT EDIT.

package automock

import (
	assignmentOperation "github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// EntityConverter is an autogenerated mock type for the EntityConverter type
type EntityConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: entity
func (_m *EntityConverter) FromEntity(entity *assignmentOperation.Entity) *model.AssignmentOperation {
	ret := _m.Called(entity)

	if len(ret) == 0 {
		panic("no return value specified for FromEntity")
	}

	var r0 *model.AssignmentOperation
	if rf, ok := ret.Get(0).(func(*assignmentOperation.Entity) *model.AssignmentOperation); ok {
		r0 = rf(entity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AssignmentOperation)
		}
	}

	return r0
}

// ToEntity provides a mock function with given fields: in
func (_m *EntityConverter) ToEntity(in *model.AssignmentOperation) *assignmentOperation.Entity {
	ret := _m.Called(in)

	if len(ret) == 0 {
		panic("no return value specified for ToEntity")
	}

	var r0 *assignmentOperation.Entity
	if rf, ok := ret.Get(0).(func(*model.AssignmentOperation) *assignmentOperation.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*assignmentOperation.Entity)
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