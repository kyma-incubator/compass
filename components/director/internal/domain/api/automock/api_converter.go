// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// APIConverter is an autogenerated mock type for the APIConverter type
type APIConverter struct {
	mock.Mock
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *APIConverter) InputFromGraphQL(in *graphql.APIDefinitionInput) *model.APIDefinitionInput {
	ret := _m.Called(in)

	var r0 *model.APIDefinitionInput
	if rf, ok := ret.Get(0).(func(*graphql.APIDefinitionInput) *model.APIDefinitionInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.APIDefinitionInput)
		}
	}

	return r0
}

// MultipleInputFromGraphQL provides a mock function with given fields: in
func (_m *APIConverter) MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput {
	ret := _m.Called(in)

	var r0 []*model.APIDefinitionInput
	if rf, ok := ret.Get(0).(func([]*graphql.APIDefinitionInput) []*model.APIDefinitionInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.APIDefinitionInput)
		}
	}

	return r0
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *APIConverter) MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition {
	ret := _m.Called(in)

	var r0 []*graphql.APIDefinition
	if rf, ok := ret.Get(0).(func([]*model.APIDefinition) []*graphql.APIDefinition); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.APIDefinition)
		}
	}

	return r0
}

// SpecToGraphQL provides a mock function with given fields: definitionID, in
func (_m *APIConverter) SpecToGraphQL(definitionID string, in *model.APISpec) *graphql.APISpec {
	ret := _m.Called(definitionID, in)

	var r0 *graphql.APISpec
	if rf, ok := ret.Get(0).(func(string, *model.APISpec) *graphql.APISpec); ok {
		r0 = rf(definitionID, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.APISpec)
		}
	}

	return r0
}

// ToGraphQL provides a mock function with given fields: in
func (_m *APIConverter) ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition {
	ret := _m.Called(in)

	var r0 *graphql.APIDefinition
	if rf, ok := ret.Get(0).(func(*model.APIDefinition) *graphql.APIDefinition); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.APIDefinition)
		}
	}

	return r0
}
