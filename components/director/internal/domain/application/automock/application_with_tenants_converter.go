// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// ApplicationWithTenantsConverter is an autogenerated mock type for the ApplicationWithTenantsConverter type
type ApplicationWithTenantsConverter struct {
	mock.Mock
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *ApplicationWithTenantsConverter) MultipleToGraphQL(in []*model.ApplicationWithTenants) []*graphql.ApplicationWithTenants {
	ret := _m.Called(in)

	if len(ret) == 0 {
		panic("no return value specified for MultipleToGraphQL")
	}

	var r0 []*graphql.ApplicationWithTenants
	if rf, ok := ret.Get(0).(func([]*model.ApplicationWithTenants) []*graphql.ApplicationWithTenants); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.ApplicationWithTenants)
		}
	}

	return r0
}

// NewApplicationWithTenantsConverter creates a new instance of ApplicationWithTenantsConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewApplicationWithTenantsConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *ApplicationWithTenantsConverter {
	mock := &ApplicationWithTenantsConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
