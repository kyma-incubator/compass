// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// AspectConverter is an autogenerated mock type for the AspectConverter type
type AspectConverter struct {
	mock.Mock
}

// MultipleInputFromGraphQL provides a mock function with given fields: in
func (_m *AspectConverter) MultipleInputFromGraphQL(in []*graphql.AspectInput) ([]*model.AspectInput, error) {
	ret := _m.Called(in)

	var r0 []*model.AspectInput
	var r1 error
	if rf, ok := ret.Get(0).(func([]*graphql.AspectInput) ([]*model.AspectInput, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func([]*graphql.AspectInput) []*model.AspectInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AspectInput)
		}
	}

	if rf, ok := ret.Get(1).(func([]*graphql.AspectInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *AspectConverter) MultipleToGraphQL(in []*model.Aspect) ([]*graphql.Aspect, error) {
	ret := _m.Called(in)

	var r0 []*graphql.Aspect
	var r1 error
	if rf, ok := ret.Get(0).(func([]*model.Aspect) ([]*graphql.Aspect, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func([]*model.Aspect) []*graphql.Aspect); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.Aspect)
		}
	}

	if rf, ok := ret.Get(1).(func([]*model.Aspect) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAspectConverter creates a new instance of AspectConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAspectConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *AspectConverter {
	mock := &AspectConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}