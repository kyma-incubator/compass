// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// RuntimeContextConverter is an autogenerated mock type for the RuntimeContextConverter type
type RuntimeContextConverter struct {
	mock.Mock
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *RuntimeContextConverter) MultipleToGraphQL(in []*model.RuntimeContext) []*graphql.RuntimeContext {
	ret := _m.Called(in)

	var r0 []*graphql.RuntimeContext
	if rf, ok := ret.Get(0).(func([]*model.RuntimeContext) []*graphql.RuntimeContext); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.RuntimeContext)
		}
	}

	return r0
}

// ToGraphQL provides a mock function with given fields: in
func (_m *RuntimeContextConverter) ToGraphQL(in *model.RuntimeContext) *graphql.RuntimeContext {
	ret := _m.Called(in)

	var r0 *graphql.RuntimeContext
	if rf, ok := ret.Get(0).(func(*model.RuntimeContext) *graphql.RuntimeContext); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.RuntimeContext)
		}
	}

	return r0
}

type mockConstructorTestingTNewRuntimeContextConverter interface {
	mock.TestingT
	Cleanup(func())
}

// NewRuntimeContextConverter creates a new instance of RuntimeContextConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeContextConverter(t mockConstructorTestingTNewRuntimeContextConverter) *RuntimeContextConverter {
	mock := &RuntimeContextConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
