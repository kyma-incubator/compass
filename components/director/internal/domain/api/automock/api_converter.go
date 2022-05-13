// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// APIConverter is an autogenerated mock type for the APIConverter type
type APIConverter struct {
	mock.Mock
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *APIConverter) InputFromGraphQL(in *graphql.APIDefinitionInput) (*model.APIDefinitionInput, *model.SpecInput, error) {
	ret := _m.Called(in)

	var r0 *model.APIDefinitionInput
	if rf, ok := ret.Get(0).(func(*graphql.APIDefinitionInput) *model.APIDefinitionInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.APIDefinitionInput)
		}
	}

	var r1 *model.SpecInput
	if rf, ok := ret.Get(1).(func(*graphql.APIDefinitionInput) *model.SpecInput); ok {
		r1 = rf(in)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.SpecInput)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(*graphql.APIDefinitionInput) error); ok {
		r2 = rf(in)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MultipleInputFromGraphQL provides a mock function with given fields: in
func (_m *APIConverter) MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, []*model.SpecInput, error) {
	ret := _m.Called(in)

	var r0 []*model.APIDefinitionInput
	if rf, ok := ret.Get(0).(func([]*graphql.APIDefinitionInput) []*model.APIDefinitionInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.APIDefinitionInput)
		}
	}

	var r1 []*model.SpecInput
	if rf, ok := ret.Get(1).(func([]*graphql.APIDefinitionInput) []*model.SpecInput); ok {
		r1 = rf(in)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]*model.SpecInput)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func([]*graphql.APIDefinitionInput) error); ok {
		r2 = rf(in)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MultipleToGraphQL provides a mock function with given fields: in, specs, bundleRefs
func (_m *APIConverter) MultipleToGraphQL(in []*model.APIDefinition, specs []*model.Spec, bundleRefs []*model.BundleReference) ([]*graphql.APIDefinition, error) {
	ret := _m.Called(in, specs, bundleRefs)

	var r0 []*graphql.APIDefinition
	if rf, ok := ret.Get(0).(func([]*model.APIDefinition, []*model.Spec, []*model.BundleReference) []*graphql.APIDefinition); ok {
		r0 = rf(in, specs, bundleRefs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.APIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*model.APIDefinition, []*model.Spec, []*model.BundleReference) error); ok {
		r1 = rf(in, specs, bundleRefs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in, spec, bundleRef
func (_m *APIConverter) ToGraphQL(in *model.APIDefinition, spec *model.Spec, bundleRef *model.BundleReference) (*graphql.APIDefinition, error) {
	ret := _m.Called(in, spec, bundleRef)

	var r0 *graphql.APIDefinition
	if rf, ok := ret.Get(0).(func(*model.APIDefinition, *model.Spec, *model.BundleReference) *graphql.APIDefinition); ok {
		r0 = rf(in, spec, bundleRef)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.APIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.APIDefinition, *model.Spec, *model.BundleReference) error); ok {
		r1 = rf(in, spec, bundleRef)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAPIConverter creates a new instance of APIConverter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewAPIConverter(t testing.TB) *APIConverter {
	mock := &APIConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
