// Code generated by mockery v2.12.1. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/machinebox/graphql"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// GraphQLClient is an autogenerated mock type for the GraphQLClient type
type GraphQLClient struct {
	mock.Mock
}

// Run provides a mock function with given fields: _a0, _a1, _a2
func (_m *GraphQLClient) Run(_a0 context.Context, _a1 *graphql.Request, _a2 interface{}) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *graphql.Request, interface{}) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewGraphQLClient creates a new instance of GraphQLClient. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewGraphQLClient(t testing.TB) *GraphQLClient {
	mock := &GraphQLClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
