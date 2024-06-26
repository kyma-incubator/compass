// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// GqlConverter is an autogenerated mock type for the gqlConverter type
type GqlConverter struct {
	mock.Mock
}

// ToGraphQL provides a mock function with given fields: in, targetTenantExternalID
func (_m *GqlConverter) ToGraphQL(in *model.AutomaticScenarioAssignment, targetTenantExternalID string) graphql.AutomaticScenarioAssignment {
	ret := _m.Called(in, targetTenantExternalID)

	var r0 graphql.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(*model.AutomaticScenarioAssignment, string) graphql.AutomaticScenarioAssignment); ok {
		r0 = rf(in, targetTenantExternalID)
	} else {
		r0 = ret.Get(0).(graphql.AutomaticScenarioAssignment)
	}

	return r0
}

// NewGqlConverter creates a new instance of GqlConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGqlConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *GqlConverter {
	mock := &GqlConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
