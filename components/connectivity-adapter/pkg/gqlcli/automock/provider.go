// Code generated by mockery v2.5.1. DO NOT EDIT.

package automock

import (
	http "net/http"

	gqlcli "github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"

	mock "github.com/stretchr/testify/mock"
)

// Provider is an autogenerated mock type for the Provider type
type Provider struct {
	mock.Mock
}

// GQLClient provides a mock function with given fields: rq
func (_m *Provider) GQLClient(rq *http.Request) gqlcli.GraphQLClient {
	ret := _m.Called(rq)

	var r0 gqlcli.GraphQLClient
	if rf, ok := ret.Get(0).(func(*http.Request) gqlcli.GraphQLClient); ok {
		r0 = rf(rq)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(gqlcli.GraphQLClient)
		}
	}

	return r0
}
