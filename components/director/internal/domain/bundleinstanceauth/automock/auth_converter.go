// Code generated by mockery v1.1.2. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// AuthConverter is an autogenerated mock type for the AuthConverter type
type AuthConverter struct {
	mock.Mock
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *AuthConverter) InputFromGraphQL(in *graphql.AuthInput) (*model.AuthInput, error) {
	ret := _m.Called(in)

	var r0 *model.AuthInput
	if rf, ok := ret.Get(0).(func(*graphql.AuthInput) *model.AuthInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AuthInput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*graphql.AuthInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in
func (_m *AuthConverter) ToGraphQL(in *model.Auth) (*graphql.Auth, error) {
	ret := _m.Called(in)

	var r0 *graphql.Auth
	if rf, ok := ret.Get(0).(func(*model.Auth) *graphql.Auth); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Auth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Auth) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
