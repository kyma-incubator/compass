// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import graphql "github.com/machinebox/graphql"
import mock "github.com/stretchr/testify/mock"

// GraphQLRequestBuilder is an autogenerated mock type for the GraphQLRequestBuilder type
type GraphQLRequestBuilder struct {
	mock.Mock
}

// GetApplicationsByName provides a mock function with given fields: appName
func (_m *GraphQLRequestBuilder) GetApplicationsByName(appName string) *graphql.Request {
	ret := _m.Called(appName)

	var r0 *graphql.Request
	if rf, ok := ret.Get(0).(func(string) *graphql.Request); ok {
		r0 = rf(appName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Request)
		}
	}

	return r0
}
