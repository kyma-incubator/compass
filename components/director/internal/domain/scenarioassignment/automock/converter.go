// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// Converter is an autogenerated mock type for the Converter type
type Converter struct {
	mock.Mock
}

// FromInputGraphQL provides a mock function with given fields: in, tenant
func (_m *Converter) FromInputGraphQL(in graphql.AutomaticScenarioAssignmentSetInput, tenant string) model.AutomaticScenarioAssignment {
	ret := _m.Called(in, tenant)

	var r0 model.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(graphql.AutomaticScenarioAssignmentSetInput, string) model.AutomaticScenarioAssignment); ok {
		r0 = rf(in, tenant)
	} else {
		r0 = ret.Get(0).(model.AutomaticScenarioAssignment)
	}

	return r0
}

// ToGraphQL provides a mock function with given fields: in
func (_m *Converter) ToGraphQL(in model.AutomaticScenarioAssignment) graphql.AutomaticScenarioAssignment {
	ret := _m.Called(in)

	var r0 graphql.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(model.AutomaticScenarioAssignment) graphql.AutomaticScenarioAssignment); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(graphql.AutomaticScenarioAssignment)
	}

	return r0
}
