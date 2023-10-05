// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// EventConverter is an autogenerated mock type for the EventConverter type
type EventConverter struct {
	mock.Mock
}

// MultipleInputFromGraphQL provides a mock function with given fields: in
func (_m *EventConverter) MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, []*model.SpecInput, error) {
	ret := _m.Called(in)

	var r0 []*model.EventDefinitionInput
	if rf, ok := ret.Get(0).(func([]*graphql.EventDefinitionInput) []*model.EventDefinitionInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.EventDefinitionInput)
		}
	}

	var r1 []*model.SpecInput
	if rf, ok := ret.Get(1).(func([]*graphql.EventDefinitionInput) []*model.SpecInput); ok {
		r1 = rf(in)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]*model.SpecInput)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func([]*graphql.EventDefinitionInput) error); ok {
		r2 = rf(in)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MultipleToGraphQL provides a mock function with given fields: in, bundleRefs
func (_m *EventConverter) MultipleToGraphQL(in []*model.EventDefinition, bundleRefs []*model.BundleReference) ([]*graphql.EventDefinition, error) {
	ret := _m.Called(in, bundleRefs)

	var r0 []*graphql.EventDefinition
	if rf, ok := ret.Get(0).(func([]*model.EventDefinition, []*model.BundleReference) []*graphql.EventDefinition); ok {
		r0 = rf(in, bundleRefs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.EventDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*model.EventDefinition, []*model.BundleReference) error); ok {
		r1 = rf(in, bundleRefs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in, spec, bundleReference
func (_m *EventConverter) ToGraphQL(in *model.EventDefinition, spec *model.Spec, bundleReference *model.BundleReference) (*graphql.EventDefinition, error) {
	ret := _m.Called(in, spec, bundleReference)

	var r0 *graphql.EventDefinition
	if rf, ok := ret.Get(0).(func(*model.EventDefinition, *model.Spec, *model.BundleReference) *graphql.EventDefinition); ok {
		r0 = rf(in, spec, bundleReference)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.EventDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.EventDefinition, *model.Spec, *model.BundleReference) error); ok {
		r1 = rf(in, spec, bundleReference)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewEventConverter interface {
	mock.TestingT
	Cleanup(func())
}

// NewEventConverter creates a new instance of EventConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEventConverter(t mockConstructorTestingTNewEventConverter) *EventConverter {
	mock := &EventConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
