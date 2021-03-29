// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"
)

// GraphQLizer is an autogenerated mock type for the GraphQLizer type
type GraphQLizer struct {
	mock.Mock
}

// BundleInstanceAuthRequestInputToGQL provides a mock function with given fields: in
func (_m *GraphQLizer) BundleInstanceAuthRequestInputToGQL(in graphql.BundleInstanceAuthRequestInput) (string, error) {
	ret := _m.Called(in)

	var r0 string
	if rf, ok := ret.Get(0).(func(graphql.BundleInstanceAuthRequestInput) string); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(graphql.BundleInstanceAuthRequestInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
