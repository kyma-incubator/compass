// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// IntegrationSystemConverter is an autogenerated mock type for the IntegrationSystemConverter type
type IntegrationSystemConverter struct {
	mock.Mock
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *IntegrationSystemConverter) InputFromGraphQL(in graphql.IntegrationSystemInput) model.IntegrationSystemInput {
	ret := _m.Called(in)

	var r0 model.IntegrationSystemInput
	if rf, ok := ret.Get(0).(func(graphql.IntegrationSystemInput) model.IntegrationSystemInput); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.IntegrationSystemInput)
	}

	return r0
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *IntegrationSystemConverter) MultipleToGraphQL(in []*model.IntegrationSystem) []*graphql.IntegrationSystem {
	ret := _m.Called(in)

	var r0 []*graphql.IntegrationSystem
	if rf, ok := ret.Get(0).(func([]*model.IntegrationSystem) []*graphql.IntegrationSystem); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.IntegrationSystem)
		}
	}

	return r0
}

// ToGraphQL provides a mock function with given fields: in
func (_m *IntegrationSystemConverter) ToGraphQL(in *model.IntegrationSystem) *graphql.IntegrationSystem {
	ret := _m.Called(in)

	var r0 *graphql.IntegrationSystem
	if rf, ok := ret.Get(0).(func(*model.IntegrationSystem) *graphql.IntegrationSystem); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.IntegrationSystem)
		}
	}

	return r0
}
