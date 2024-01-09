// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// AspectEventResourceConverter is an autogenerated mock type for the AspectEventResourceConverter type
type AspectEventResourceConverter struct {
	mock.Mock
}

// MultipleInputFromGraphQL provides a mock function with given fields: in
func (_m *AspectEventResourceConverter) MultipleInputFromGraphQL(in []*graphql.AspectEventDefinitionInput) ([]*model.AspectEventResourceInput, error) {
	ret := _m.Called(in)

	var r0 []*model.AspectEventResourceInput
	var r1 error
	if rf, ok := ret.Get(0).(func([]*graphql.AspectEventDefinitionInput) ([]*model.AspectEventResourceInput, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func([]*graphql.AspectEventDefinitionInput) []*model.AspectEventResourceInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AspectEventResourceInput)
		}
	}

	if rf, ok := ret.Get(1).(func([]*graphql.AspectEventDefinitionInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *AspectEventResourceConverter) MultipleToGraphQL(in []*model.AspectEventResource) ([]*graphql.AspectEventDefinition, error) {
	ret := _m.Called(in)

	var r0 []*graphql.AspectEventDefinition
	var r1 error
	if rf, ok := ret.Get(0).(func([]*model.AspectEventResource) ([]*graphql.AspectEventDefinition, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func([]*model.AspectEventResource) []*graphql.AspectEventDefinition); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.AspectEventDefinition)
		}
	}

	if rf, ok := ret.Get(1).(func([]*model.AspectEventResource) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAspectEventResourceConverter creates a new instance of AspectEventResourceConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAspectEventResourceConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *AspectEventResourceConverter {
	mock := &AspectEventResourceConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
