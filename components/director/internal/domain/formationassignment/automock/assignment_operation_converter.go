// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// AssignmentOperationConverter is an autogenerated mock type for the assignmentOperationConverter type
type AssignmentOperationConverter struct {
	mock.Mock
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *AssignmentOperationConverter) MultipleToGraphQL(in []*model.AssignmentOperation) []*graphql.AssignmentOperation {
	ret := _m.Called(in)

	if len(ret) == 0 {
		panic("no return value specified for MultipleToGraphQL")
	}

	var r0 []*graphql.AssignmentOperation
	if rf, ok := ret.Get(0).(func([]*model.AssignmentOperation) []*graphql.AssignmentOperation); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.AssignmentOperation)
		}
	}

	return r0
}

// NewAssignmentOperationConverter creates a new instance of AssignmentOperationConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAssignmentOperationConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *AssignmentOperationConverter {
	mock := &AssignmentOperationConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
