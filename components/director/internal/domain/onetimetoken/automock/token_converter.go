// Code generated by mockery v2.12.1. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// TokenConverter is an autogenerated mock type for the TokenConverter type
type TokenConverter struct {
	mock.Mock
}

// ToGraphQLForApplication provides a mock function with given fields: _a0
func (_m *TokenConverter) ToGraphQLForApplication(_a0 model.OneTimeToken) (graphql.OneTimeTokenForApplication, error) {
	ret := _m.Called(_a0)

	var r0 graphql.OneTimeTokenForApplication
	if rf, ok := ret.Get(0).(func(model.OneTimeToken) graphql.OneTimeTokenForApplication); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(graphql.OneTimeTokenForApplication)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.OneTimeToken) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQLForRuntime provides a mock function with given fields: _a0
func (_m *TokenConverter) ToGraphQLForRuntime(_a0 model.OneTimeToken) graphql.OneTimeTokenForRuntime {
	ret := _m.Called(_a0)

	var r0 graphql.OneTimeTokenForRuntime
	if rf, ok := ret.Get(0).(func(model.OneTimeToken) graphql.OneTimeTokenForRuntime); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(graphql.OneTimeTokenForRuntime)
	}

	return r0
}

// NewTokenConverter creates a new instance of TokenConverter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewTokenConverter(t testing.TB) *TokenConverter {
	mock := &TokenConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
