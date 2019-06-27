// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"
import "github.com/stretchr/testify/mock"
import "github.com/kyma-incubator/compass/components/director/internal/model"

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
